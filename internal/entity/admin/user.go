package entity_admin

import (
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
)

type RespUser struct {
	ID             uint           `json:"id"`
	Username       string         `json:"username"`
	UID            string         `json:"uid"`
	CreatedAt      entity.Time    `json:"created_at"`
	FirstDepositAt entity.Time    `json:"first_reposit_at"`
	LoginMethod    db.LoginMethod `json:"login_method"`
	LastUsedAt     entity.Time    `json:"last_used_at"`
	PasskeyCount   int            `json:"passkey_count"`
	SecurityLevel  uint8          `json:"security_level"`
	LastDepositAt  entity.Time    `json:"last_reposit_at"`
	LastWithdrawAt entity.Time    `json:"last_withdraw_at"`
}

func (r *RespUser) Fill(a *db.UserList) *RespUser {
	r.ID = a.ID
	r.Username = a.Username
	r.UID = a.UID
	r.LoginMethod = a.LoginMethod
	r.PasskeyCount = a.PasskeyCount
	r.CreatedAt = entity.Time(a.CreatedAt)
	r.LastUsedAt = entity.Time(a.LastUsedAt)
	r.FirstDepositAt = entity.Time(a.FirstDepositAt)
	r.LastDepositAt = entity.Time(a.LastDepositAt)
	r.LastWithdrawAt = entity.Time(a.LastWithdrawAt)

	switch r.PasskeyCount {
	case 0:
		r.SecurityLevel = 0
	case 1, 2:
		r.SecurityLevel = 1
	default:
		r.SecurityLevel = 2
	}
	return r
}

type RespPasskey struct {
	UID            string      `json:"id"`
	CredentialID   string      `json:"credential_id"`
	PublicKey      string      `json:"public_key"`
	Name           string      `json:"name"`
	LastUsedAt     entity.Time `json:"last_used_at"`
	LastUsedIP     string      `json:"last_used_ip"`
	Synced         bool        `json:"synced"`
	EOSAccount     string      `json:"eos_account"`
	EOSPermissions string      `json:"eos_permissions"`
	DeviceID       string      `json:"device_id"`
	BlockNumber    uint64      `json:"block_number"`
	AAGuid         string      `json:"aaguid"`
}

func (r *RespPasskey) Fill(a *db.UserCredential) *RespPasskey {
	r.UID = a.UID
	r.CredentialID = a.CredentialID
	r.PublicKey = a.PublicKey
	r.Name = a.Name
	r.LastUsedAt = entity.Time(a.LastUsedAt)
	r.LastUsedIP = a.LastUsedIP
	r.Synced = a.Synced
	r.EOSAccount = a.EOSAccount
	r.EOSPermissions = a.EOSPermissions
	r.DeviceID = a.DeviceID
	r.BlockNumber = a.BlockNumber
	r.AAGuid = a.AAGuid
	return r
}

type RespTransactionsRecord struct {
	ID             uint        `json:"id"`
	DepositAt      entity.Time `json:"deposit_at"`
	WithdrawAt     entity.Time `json:"withdraw_at"`
	Symbol         string      `json:"symbol"`
	CoinName       string      `json:"coin_name"`
	EVMAddress     string      `json:"evm_address"`
	UID            string      `json:"uid"`
	Free           float64     `json:"free"`
	DepositChain   string      `json:"deposit_chain"`
	WithdrawChain  string      `json:"withdraw_chain"`
	DepositAmount  float64     `json:"deposit_amount"`
	WithdrawAmount float64     `json:"withdraw_amount"`
	TxHash         string      `json:"tx_hash"`
}
