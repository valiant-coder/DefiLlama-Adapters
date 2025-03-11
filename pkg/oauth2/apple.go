package oauth2

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

type AppleUserInfo struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Name   struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	} `json:"name"`
	EmailVerified  bool `json:"email_verified"`
	IsPrivateEmail bool `json:"is_private_email"`
}

func ParseAppleIDToken(idToken, clientID string) (*AppleUserInfo, error) {
	appleKeys, err := getApplePublicKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to get Apple public keys: %v", err)
	}

	token, err := jwt.Parse(idToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		kid := token.Header["kid"].(string)
		for _, key := range appleKeys.Keys {
			if key.Kid == kid {
				nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
				if err != nil {
					return nil, fmt.Errorf("failed to decode modulus: %v", err)
				}
				eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
				if err != nil {
					return nil, fmt.Errorf("failed to decode exponent: %v", err)
				}

				n := new(big.Int).SetBytes(nBytes)
				e := new(big.Int).SetBytes(eBytes)

				pub := &rsa.PublicKey{
					N: n,
					E: int(e.Int64()),
				}
				return pub, nil
			}
		}
		return nil, fmt.Errorf("matching public key not found")
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if claims["iss"].(string) != "https://appleid.apple.com" {
		return nil, fmt.Errorf("invalid token issuer")
	}

	if claims["aud"].(string) != clientID {
		return nil, fmt.Errorf("invalid client ID")
	}

	userInfo := &AppleUserInfo{
		UserID: claims["sub"].(string),
	}

	if email, ok := claims["email"].(string); ok {
		userInfo.Email = email
	}
	if emailVerified, ok := claims["email_verified"].(bool); ok {
		userInfo.EmailVerified = emailVerified
	}
	if isPrivateEmail, ok := claims["is_private_email"].(bool); ok {
		userInfo.IsPrivateEmail = isPrivateEmail
	}

	return userInfo, nil
}

var (
	appleKeysCache     *ApplePublicKeys
	appleKeysCacheTime time.Time
	appleKeysCacheTTL  = 24 * time.Hour
)

func getApplePublicKeys() (*ApplePublicKeys, error) {
	if appleKeysCache != nil && time.Since(appleKeysCacheTime) < appleKeysCacheTTL {
		return appleKeysCache, nil
	}

	resp, err := http.Get("https://appleid.apple.com/auth/keys")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var keys ApplePublicKeys
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}

	appleKeysCache = &keys
	appleKeysCacheTime = time.Now()

	return &keys, nil
}

type ApplePublicKeys struct {
	Keys []struct {
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	} `json:"keys"`
}
