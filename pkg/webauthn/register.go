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

// BeginRegistration starts the registration process
func (w *WebAuthnSDK) BeginRegistration(ctx context.Context, user User) (*protocol.CredentialCreation, string, error) {
	options, sessionData, err := w.webauthn.BeginRegistration(user)
	if err != nil {
		return nil, "", fmt.Errorf("begin registration failed: %w", err)
	}

	// Serialize session data
	sessionBytes, err := json.Marshal(sessionData)
	if err != nil {
		return nil, "", fmt.Errorf("marshal session data failed: %w", err)
	}

	// Generate session ID
	sessionID := generateSessionID()

	// Store session data
	err = w.store.Store(ctx, sessionID, sessionBytes, 5*time.Minute)
	if err != nil {
		return nil, "", fmt.Errorf("store session failed: %w", err)
	}

	return options, sessionID, nil
}

// FinishRegistration completes the registration process
func (w *WebAuthnSDK) FinishRegistration(ctx context.Context, user User, sessionID string, response *http.Request) (*webauthn.Credential, error) {
	// Get session data
	sessionBytes, err := w.store.Get(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session failed: %w", err)
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionBytes, &sessionData); err != nil {
		return nil, fmt.Errorf("unmarshal session data failed: %w", err)
	}

	credential, err := w.webauthn.FinishRegistration(user, sessionData, response)
	if err != nil {
		return nil, fmt.Errorf("finish registration failed: %w", err)
	}

	// Clean up session data
	_ = w.store.Delete(ctx, sessionID)

	return credential, nil
}
