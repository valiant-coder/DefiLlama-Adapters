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
	"fmt"
	"strconv"
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

// LoginHandler defines the interface for different login methods
type LoginHandler interface {
	Handle(req entity.ReqUserLogin) (*db.User, error)
}

// GoogleLoginHandler handles Google login
type GoogleLoginHandler struct {
	clientID string
}

func (h *GoogleLoginHandler) Handle(req entity.ReqUserLogin) (*db.User, error) {
	userInfo, err := oauth2.VerifyGoogleToken(req.IdToken, h.clientID)
	if err != nil {
		return nil, fmt.Errorf("verify google token: %w", err)
	}
	return &db.User{
		Username:    userInfo.Name,
		OauthID:     userInfo.GoogleID,
		LoginMethod: db.LoginMethodGoogle,
		Avatar:      userInfo.Picture,
		Email:       userInfo.Email,
	}, nil
}

// AppleLoginHandler handles Apple login
type AppleLoginHandler struct {
	clientID string
}

func (h *AppleLoginHandler) Handle(req entity.ReqUserLogin) (*db.User, error) {
	userInfo, err := oauth2.ParseAppleIDToken(req.IdToken, h.clientID)
	if err != nil {
		return nil, fmt.Errorf("verify apple token: %w", err)
	}
	return &db.User{
		Username:    req.UserName,
		OauthID:     userInfo.UserID,
		LoginMethod: db.LoginMethodApple,
		Email:       userInfo.Email,
	}, nil
}

// TelegramLoginHandler handles Telegram login
type TelegramLoginHandler struct {
	botToken string
}

func (h *TelegramLoginHandler) Handle(req entity.ReqUserLogin) (*db.User, error) {
	userInfo, err := oauth2.VerifyTelegramLogin(h.botToken, oauth2.TelegramData{
		ID:        req.TelegramData.ID,
		FirstName: req.TelegramData.FirstName,
		LastName:  req.TelegramData.LastName,
		Username:  req.TelegramData.Username,
		PhotoURL:  req.TelegramData.PhotoURL,
		Hash:      req.TelegramData.Hash,
		AuthDate:  req.TelegramData.AuthDate,
	})
	if err != nil {
		return nil, fmt.Errorf("verify telegram data: %w", err)
	}
	username := userInfo.Username
	if username == "" {
		username = fmt.Sprintf("%s %s", userInfo.FirstName, userInfo.LastName)
	}
	return &db.User{
		Username:    username,
		OauthID:     strconv.FormatInt(userInfo.ID, 10),
		LoginMethod: db.LoginMethodTelegram,
		Avatar:      userInfo.PhotoURL,
	}, nil
}

// EVMLoginHandler handles EVM login
type EVMLoginHandler struct {
}

func (h *EVMLoginHandler) Handle(req entity.ReqUserLogin) (*db.User, error) {

	return &db.User{
		Username:    req.EVMAddress,
		LoginMethod: db.LoginMethodEVM,
		OauthID:     req.EVMAddress,
		EVMAddress:  req.EVMAddress,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req entity.ReqUserLogin) (string, error) {
	cfg := config.Conf()

	handlers := map[db.LoginMethod]LoginHandler{
		db.LoginMethodGoogle: &GoogleLoginHandler{
			clientID: cfg.Oauth2.Google.ClientID,
		},
		db.LoginMethodApple: &AppleLoginHandler{
			clientID: cfg.Oauth2.Apple.ClientID,
		},
		db.LoginMethodTelegram: &TelegramLoginHandler{
			botToken: cfg.Oauth2.Telegram.BotToken,
		},
		db.LoginMethodEVM: &EVMLoginHandler{},
	}

	handler, ok := handlers[db.LoginMethod(req.Method)]
	if !ok {
		return "", errors.New("invalid login method")
	}

	user, err := handler.Handle(req)
	if err != nil {
		log.Logger().Errorf("login failed for method %s: %v", req.Method, err)
		return "", err
	}

	if err := s.repo.UpsertUser(ctx, user); err != nil {
		return "", fmt.Errorf("upsert user: %w", err)
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

	var eosAccount, permission string
	if user.LoginMethod == db.LoginMethodEVM {
		eosAccount = user.EOSAccount
		permission = user.Permission
	} else {
		for _, c := range credentials {
			if c.EOSAccount != "" {
				eosAccount = c.EOSAccount
				permission = "active"
				break
			}
		}
	}

	return entity.RespUserInfo{
		UID:      user.UID,
		UserName: user.Username,
		Passkeys: credentials,
		Email:    user.Email,

		// for evm user
		EVMAddress: user.EVMAddress,
		EOSAccount: eosAccount,
		Permission: permission,
	}, nil
}

func (s *UserService) GetEOSAccountAndPermissionByUID(ctx context.Context, uid string) (string, string, error) {
	return s.repo.GetEOSAccountAndPermissionByUID(ctx, uid)
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
