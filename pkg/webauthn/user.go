package webauthn

import "github.com/go-webauthn/webauthn/webauthn"

// User implements webauthn.User interface
type User interface {
	// WebAuthnID returns the user's unique identifier
	WebAuthnID() []byte
	// WebAuthnName returns the username
	WebAuthnName() string
	// WebAuthnDisplayName returns the user's display name
	WebAuthnDisplayName() string
	// WebAuthnIcon returns the user's icon URL
	WebAuthnIcon() string
	// WebAuthnCredentials returns the user's registered credentials
	WebAuthnCredentials() []webauthn.Credential
}
