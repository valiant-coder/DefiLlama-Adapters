package marketplace

import (
	"context"
	"errors"
	"exapp-go/config"
	"exapp-go/internal/db/ckhdb"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/log"
	"exapp-go/pkg/nsqutil"
	"exapp-go/pkg/oauth2"
	"strings"
	"time"
)

type UserService struct {
	db      *db.Repo
	ckhRepo *ckhdb.ClickHouseRepo
	nsqPub  *nsqutil.Publisher
}

func NewUserService() *UserService {
	nsqConf := config.Conf().Nsq
	return &UserService{
		db:      db.New(),
		ckhRepo: ckhdb.New(),
		nsqPub:  nsqutil.NewPublisher(nsqConf.Nsqds),
	}
}

func (s *UserService) Login(ctx context.Context, req entity.ReqUserLogin) (string, error) {
	cfg := config.Conf()
	var user *db.User
	switch req.Method {
	case "google":
		userInfo, err := oauth2.VerifyGoogleToken(req.IdToken, cfg.Oauth2.Google.ClientID)
		if err != nil {
			log.Logger().Errorf("verify google token error: %v,id_token: %s", err, req.IdToken)
			return "", err
		}
		user = &db.User{
			Username:    userInfo.Name,
			OauthID:     userInfo.GoogleID,
			LoginMethod: db.LoginMethodGoogle,
		}
	case "apple":
		userInfo, err := oauth2.ParseAppleIDToken(req.IdToken, cfg.Oauth2.Apple.ClientID)
		if err != nil {
			log.Logger().Errorf("verify apple token error: %v,id_token: %s", err, req.IdToken)
			return "", err
		}
		user = &db.User{
			Username:    userInfo.Name.FirstName + " " + userInfo.Name.LastName,
			OauthID:     userInfo.UserID,
			LoginMethod: db.LoginMethodApple,
		}
	default:
		return "", errors.New("invalid login method")
	}

	if err := s.db.CreateUserIfNotExist(ctx, user); err != nil {
		return "", err
	}

	return user.UID, nil
}

func (s *UserService) IsUserExist(ctx context.Context, uid string) (bool, error) {
	return s.db.IsUserExist(ctx, uid)
}

func (s *UserService) GetUserCredentials(ctx context.Context, uid string) ([]entity.RespUserCredential, error) {
	credentials, err := s.db.GetUserCredentials(ctx, uid)
	if err != nil {
		return nil, err
	}

	var dst []entity.RespUserCredential
	for _, v := range credentials {
		dst = append(dst, entity.RespUserCredential{
			UserCredential: entity.ToUserCredential(v),
			CreatedAt:      entity.Time(v.CreatedAt),
			LastUsedAt:     entity.Time(v.LastUsedAt),
			LastUsedIP:     v.LastUsedIP,
			EOSAccount:     v.EOSAccount,
			EOSPermission:  strings.Split(v.EOSPermissions, ","),
		})
	}

	return dst, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, uid string) (entity.RespUserInfo, error) {
	user, err := s.db.GetUser(ctx, uid)
	if err != nil {
		return entity.RespUserInfo{}, err
	}

	credentials, err := s.GetUserCredentials(ctx, uid)
	if err != nil {
		return entity.RespUserInfo{}, err
	}

	return entity.RespUserInfo{
		UID:      user.UID,
		UserName: user.Username,
		Passkeys: credentials,
	}, nil
}

func (s *UserService) CreateUserCredential(ctx context.Context, req entity.UserCredential, uid string) error {
	newUserCredential := db.UserCredential{
		UID:          uid,
		CredentialID: req.CredentialID,
		PublicKey:    req.PublicKey,
		Name:         req.Name,
		Synced:       req.Synced,
		DeviceID:     req.DeviceID,
	}
	if err := s.db.CreateCredentialIfNotExist(ctx, &newUserCredential); err != nil {
		return err
	}

	msg := struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{
		Type: "new_user_credential",
		Data: entity.ToUserCredential(newUserCredential),
	}
	return s.nsqPub.Publish("cdex_updates", msg)
}

func (s *UserService) UpdateUserCredentialUsage(ctx context.Context, publicKey string, ip string) error {
	credential, err := s.db.GetUserCredentialByPubkey(ctx, publicKey)
	if err != nil {
		return err
	}
	credential.LastUsedAt = time.Now()
	credential.LastUsedIP = ip
	return s.db.UpdateUserCredential(ctx, credential)
}
