package admin

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"exapp-go/internal/db/db"
	entity_admin "exapp-go/internal/entity/admin"
	"exapp-go/internal/errno"
	"exapp-go/pkg/utils"

	uuid "github.com/satori/go.uuid"
	"github.com/spf13/cast"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const loginKeyPrefix = "admin:login:%s"

func (s *AdminServices) Login(ctx context.Context, req *entity_admin.ReqLogin) (*entity_admin.RespLogin, error) {
	admin, err := s.repo.GetAdminByName(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.DefaultParamsError("User does not exist")
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password))
	if err != nil {
		return nil, errno.DefaultParamsError("Incorrect password")
	}

	key := uuid.NewV4().String()
	loginKey := fmt.Sprintf(loginKeyPrefix, key)
	err = s.repo.Redis().Set(ctx, loginKey, admin.ID, time.Minute*5).Err()
	if err != nil {
		return nil, err
	}
	return &entity_admin.RespLogin{
		Key:          key,
		IsFirstLogin: admin.FirstLogin,
	}, nil
}

func (s *AdminServices) Auth(ctx context.Context, req *entity_admin.ReqAuth) (string, error) {
	var adminID uint
	err := s.repo.Redis().Get(ctx, fmt.Sprintf(loginKeyPrefix, req.Key)).Scan(&adminID)
	if err != nil {
		return "", err
	}

	var admin *db.Admin
	admin, err = s.repo.GetAdmin(ctx, cast.ToString(adminID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errno.DefaultParamsError("User does not exist")
		}
		return "", err
	}

	if admin.GoogleAuthSecret == "" {
		return "", errno.DefaultParamsError("Please bind Google Authenticator first")
	}

	ok, err := utils.VerifyGoogleAuth(admin.GoogleAuthSecret, req.GoogleVerifyCode)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errno.DefaultParamsError("Incorrect verification code")
	}
	now := time.Now()
	admin.LastLoginAt = &now
	admin.FirstLogin = false
	err = s.repo.SaveAdmin(ctx, admin.ID, map[string]interface{}{
		"last_login_at": admin.LastLoginAt,
		"first_login":   admin.FirstLogin})
	if err != nil {
		return "", err
	}
	return admin.Name, nil
}

func (s *AdminServices) CheckAdminIsExist(ctx context.Context, name string) (bool, error) {
	ok, err := s.repo.IsAdminExists(ctx, name)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	return true, nil
}

func (s *AdminServices) ResetPassword(ctx context.Context, req *entity_admin.ReqResetPassword) error {
	admin, err := s.repo.GetAdminByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errno.DefaultParamsError("User does not exist")
		}
		return err
	}

	if strings.EqualFold(req.NewPassword, req.OldPassword) {
		return errno.DefaultParamsError("New password cannot be the same as old password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword))
	if err != nil {
		return errno.DefaultParamsError("Incorrect old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin.Password = string(hashedPassword)
	return s.repo.SaveAdmin(ctx,
		admin.ID,
		map[string]interface{}{"password": admin.Password})
}

func (s *AdminServices) GetGoogleAuthSecret(ctx context.Context, name string) (*entity_admin.RespGoogleAuth, error) {
	admin, err := s.repo.GetAdminByName(ctx, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errno.DefaultParamsError("User does not exist")
		}
		return nil, err
	}

	if admin.GoogleAuthSecret != "" && !admin.FirstLogin {
		return nil, errno.DefaultParamsError("Google Authenticator already bound")
	}

	secret, err := utils.GenerateSecretKey(uuid.NewV4().String())
	if err != nil {
		return nil, err
	}

	admin.GoogleAuthSecret = secret

	err = s.repo.SaveAdmin(ctx,
		admin.ID,
		map[string]interface{}{"google_auth_secret": admin.GoogleAuthSecret})
	if err != nil {
		return nil, err
	}

	return &entity_admin.RespGoogleAuth{
		Secret: secret,
		QRData: fmt.Sprintf("otpauth://totp/%s?issuer=OneDex.Admin&secret=%s", name, secret),
	}, nil

}
