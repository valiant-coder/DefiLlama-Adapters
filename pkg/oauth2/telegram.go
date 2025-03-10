package oauth2

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TelegramUserInfo struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
}

type TelegramData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
	Hash      string `json:"hash"`
	AuthDate  string `json:"auth_date"`
}

// VerifyTelegramLogin verifies Telegram login data
func VerifyTelegramLogin(botToken string, data TelegramData) (*TelegramUserInfo, error) {

	// Check required fields
	checkHash := data.Hash
	authDate, err := strconv.ParseInt(data.AuthDate, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth date: %v", err)
	}

	// Check if timestamp is expired (valid within 24 hours)
	if time.Now().Unix()-authDate > 86400 {
		return nil, errors.New("auth date expired")
	}

	// Generate data check string
	dataCheckString := make([]string, 0, 5)
	dataCheckString = append(dataCheckString, fmt.Sprintf("id=%d", data.ID))
	if data.FirstName != "" {
		dataCheckString = append(dataCheckString, fmt.Sprintf("first_name=%s", data.FirstName))
	}
	if data.LastName != "" {
		dataCheckString = append(dataCheckString, fmt.Sprintf("last_name=%s", data.LastName))
	}
	if data.Username != "" {
		dataCheckString = append(dataCheckString, fmt.Sprintf("username=%s", data.Username))
	}
	if data.PhotoURL != "" {
		dataCheckString = append(dataCheckString, fmt.Sprintf("photo_url=%s", data.PhotoURL))
	}
	dataCheckString = append(dataCheckString, fmt.Sprintf("auth_date=%s", data.AuthDate))
	sort.Strings(dataCheckString)

	// Generate key using bot token
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(strings.Join(dataCheckString, "\n")))
	hash := hex.EncodeToString(mac.Sum(nil))

	// Verify hash
	if hash != checkHash {
		return nil, errors.New("data hash mismatch")
	}

	// Parse user info
	userInfo := &TelegramUserInfo{
		ID:        data.ID,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Username:  data.Username,
		PhotoURL:  data.PhotoURL,
	}

	return userInfo, nil
}
