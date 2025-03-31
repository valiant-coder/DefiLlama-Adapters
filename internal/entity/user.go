package entity

import "exapp-go/internal/db/db"

type ReqUserLogin struct {
	// google,apple,telegram,evm
	Method  string `json:"method"`
	IdToken string `json:"id_token"`
	// for apple
	UserName string `json:"user_name"`

	TelegramData TelegramData `json:"telegram_data,omitempty"`

	// for evm user
	EVMAddress string `json:"evm_address"`

	// Invite Code
	InviteCode string `json:"invite_code"`
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

type RespUserInfo struct {
	UID      string               `json:"uid"`
	UserName string               `json:"user_name"`
	Passkeys []RespUserCredential `json:"passkeys"`
	Email    string               `json:"email"`

	// for evm user
	EVMAddress string `json:"evm_address"`
	EOSAccount string `json:"eos_account"`
	Permission string `json:"permission"`
}

type UserCredential struct {
	DeviceID     string `json:"device_id"`
	AAGuid       string `json:"aaguid"`
	CredentialID string `json:"credential_id"`
	PublicKey    string `json:"public_key"`
	Name         string `json:"name"`
	Synced       bool   `json:"synced"`
	EOSAccount   string `json:"eos_account,omitempty"`
	UID          string `json:"uid"`
	Storage      string `json:"storage"`
}

func ToUserCredential(credential db.UserCredential) UserCredential {
	return UserCredential{
		DeviceID:     credential.DeviceID,
		CredentialID: credential.CredentialID,
		PublicKey:    credential.PublicKey,
		Name:         credential.Name,
		Synced:       credential.Synced,
		AAGuid:       credential.AAGuid,
		EOSAccount:   credential.EOSAccount,
		UID:          credential.UID,
	}
}

type RespUserCredential struct {
	UserCredential
	CreatedAt     Time     `json:"created_at"`
	LastUsedIP    string   `json:"last_used_ip"`
	LastUsedAt    Time     `json:"last_used_at"`
	EOSAccount    string   `json:"eos_account"`
	EOSPermission []string `json:"eos_permission"`
}

type UserBalance struct {
	Coin        string        `json:"coin"`
	Balance     string        `json:"balance"`
	USDTPrice   string        `json:"usdt_price"`
	Locked      string        `json:"locked"`
	Withdrawing string        `json:"withdrawing"`
	Depositing  string        `json:"depositing"`
	Locks       []LockBalance `json:"locks"`
}

type LockBalance struct {
	PoolID     uint64 `json:"pool_id"`
	PoolSymbol string `json:"pool_symbol"`
	Balance    string `json:"balance"`
}
