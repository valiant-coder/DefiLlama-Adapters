package webauthn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

func (w *WebAuthnSDK) BeginAuthentication(ctx context.Context, user User) (*protocol.CredentialAssertion, string, error) {
	options, sessionData, err := w.webauthn.BeginLogin(user)
	if err != nil {
		return nil, "", fmt.Errorf("begin authentication failed: %w", err)
	}

	sessionBytes, err := json.Marshal(sessionData)
	if err != nil {
		return nil, "", fmt.Errorf("marshal session data failed: %w", err)
	}

	sessionID := generateSessionID()
	err = w.store.Store(ctx, sessionID, sessionBytes, 5*time.Minute)
	if err != nil {
		return nil, "", fmt.Errorf("store session failed: %w", err)
	}

	return options, sessionID, nil
}


func (w *WebAuthnSDK) FinishAuthentication(ctx context.Context, user User, sessionID string, response *http.Request) error {
	sessionBytes, err := w.store.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get session failed: %w", err)
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionBytes, &sessionData); err != nil {
		return fmt.Errorf("unmarshal session data failed: %w", err)
	}

	_, err = w.webauthn.FinishLogin(user, sessionData, response)
	if err != nil {
		return fmt.Errorf("finish authentication failed: %w", err)
	}

	_ = w.store.Delete(ctx, sessionID)
	return nil
}
