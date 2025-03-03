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
	repo           *db.Repo
	ckhRepo        *ckhdb.ClickHouseRepo
	nsqPub         *nsqutil.Publisher
	priceCache     map[string]string
	priceCacheTime time.Time
}

func NewUserService() *UserService {
	nsqConf := config.Conf().Nsq
	return &UserService{
		repo:       db.New(),
		ckhRepo:    ckhdb.New(),
		nsqPub:     nsqutil.NewPublisher(nsqConf.Nsqds),
		priceCache: make(map[string]string),
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
			Avatar:      userInfo.Picture,
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

	if err := s.repo.CreateUserIfNotExist(ctx, user); err != nil {
		return "", err
	}

	return user.UID, nil
}

func (s *UserService) IsUserExist(ctx context.Context, uid string) (bool, error) {
	return s.repo.IsUserExist(ctx, uid)
}

func (s *UserService) GetUserCredentials(ctx context.Context, uid string) ([]entity.RespUserCredential, error) {
	credentials, err := s.repo.GetUserCredentials(ctx, uid)
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
	user, err := s.repo.GetUser(ctx, uid)
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
		AAGuid:       req.AAGuid,
	}
	if err := s.repo.CreateCredentialIfNotExist(ctx, &newUserCredential); err != nil {
		return err
	}

	go func() {
		msg := struct {
			Type string      `json:"type"`
			Data interface{} `json:"data"`
		}{
			Type: "new_user_credential",
			Data: entity.ToUserCredential(newUserCredential),
		}
		err := s.nsqPub.Publish("cdex_updates", msg)
		if err != nil {
			log.Logger().Errorf("publish new user credential error: %v", err)
		}
	}()
	return nil
}

func (s *UserService) UpdateUserCredentialUsage(ctx context.Context, publicKey string, ip string) error {
	credential, err := s.repo.GetUserCredentialByPubkey(ctx, publicKey)
	if err != nil {
		return err
	}
	credential.LastUsedAt = time.Now()
	credential.LastUsedIP = ip
	return s.repo.UpdateUserCredential(ctx, credential)
}

func (s *UserService) DeleteUserCredential(ctx context.Context, uid string, credentialID string) error {
	credentials, err := s.repo.GetUserCredentials(ctx, uid)
	if err != nil {
		return err
	}

	var targetCredential *db.UserCredential
	for _, c := range credentials {
		if c.CredentialID == credentialID {
			targetCredential = &c
			break
		}
	}

	if targetCredential == nil {
		return errors.New("credential not found or not belong to user")
	}

	return s.repo.DeleteUserCredential(ctx, targetCredential)
}
