package oauth2

import (
	"context"
	"fmt"

	"google.golang.org/api/idtoken"
)

type GoogleUserInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	GoogleID      string `json:"sub"`
}

func VerifyGoogleToken(idToken, clientID string) (*GoogleUserInfo, error) {

	payload, err := idtoken.Validate(context.Background(), idToken, clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %v", err)
	}

	userInfo := &GoogleUserInfo{
		Email:         payload.Claims["email"].(string),
		EmailVerified: payload.Claims["email_verified"].(bool),
		Name:          payload.Claims["name"].(string),
		Picture:       payload.Claims["picture"].(string),
		GoogleID:      payload.Claims["sub"].(string),
	}

	return userInfo, nil
}

