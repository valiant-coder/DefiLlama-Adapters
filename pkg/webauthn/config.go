package webauthn

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
)

type WebAuthnSDK struct {
	webauthn *webauthn.WebAuthn
	store    SessionStore
}

type Config struct {
	RPDisplayName string 
	RPID          string 
	RPOrigin      string 
	Timeout       int    
}

func NewWebAuthnSDK(config Config, store SessionStore) (*WebAuthnSDK, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: config.RPDisplayName,
		RPID:          config.RPID,
		RPOrigins:     []string{config.RPOrigin},
		Timeouts: webauthn.TimeoutsConfig{
			Registration: webauthn.TimeoutConfig{
				Enforce: true,
				Timeout: time.Duration(config.Timeout) * time.Second,
			},
			Login: webauthn.TimeoutConfig{
				Enforce: true,
				Timeout: time.Duration(config.Timeout) * time.Second,
			},
		},
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, err
	}

	return &WebAuthnSDK{
		webauthn: w,
		store:    store,
	}, nil
}
