package marketplace

import (
	"context"
	"errors"
	"exapp-go/config"
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
	"exapp-go/pkg/log"
	"exapp-go/pkg/oauth2"
)

type UserService struct {
	db *db.Repo
}

func NewUserService() *UserService {
	return &UserService{db: db.New()}
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
