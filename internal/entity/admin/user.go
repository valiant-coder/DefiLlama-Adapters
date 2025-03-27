package entity_admin

import (
	"exapp-go/internal/db/db"
)

type RespUser struct {
	ID             uint           `json:"id"`
	Username       string         `json:"username"`
	UID            string         `json:"uid"`
	CreatedAt      string         `json:"created_at"`
	FirstDepositAt string         `json:"first_reposit_at"`
	LoginMethod    db.LoginMethod `json:"login_method"`
	LastUsedAt     string         `json:"last_used_at"`
	PasskeyCount   int            `json:"passkey_count"`
	SecurityLevel  uint8          `json:"security_level"`
	LastDepositAt  string         `json:"last_reposit_at"`
	LastWithdrawAt string         `json:"last_withdraw_at"`
}

func (r *RespUser) Fill(a *db.UserList) *RespUser {
	r.ID = a.ID
	r.Username = a.Username
	r.UID = a.UID
	r.LoginMethod = a.LoginMethod
	r.PasskeyCount = a.PasskeyCount
	r.CreatedAt = a.CreatedAt.Format("2006-01-02 15:04:05")
	r.LastUsedAt = a.LastUsedAt.Format("2006-01-02 15:04:05")
	r.FirstDepositAt = a.FirstDepositAt.Format("2006-01-02 15:04:05")
	r.LastDepositAt = a.LastDepositAt.Format("2006-01-02 15:04:05")
	r.LastWithdrawAt = a.LastWithdrawAt.Format("2006-01-02 15:04:05")

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
	Name       string `json:"name"`
	IsAuth     bool   `json:"is_auto"`
	Storage    string `json:"storage"`
	Synced     bool   `json:"synced"`
	LastUsedAt string `json:"last_used_at"`
	LastUsedIP string `json:"last_used_ip"`
	SignupAt   string `json:"signin_at"`
}

func (r *RespPasskey) Fill(a *db.UserCredential) *RespPasskey {
	r.Name = a.Name
	r.IsAuth = false
	r.Storage = a.Storage
	r.Synced = a.Synced
	r.SignupAt = a.CreatedAt.Format("2006-01-02 15:04:05")
	r.LastUsedAt = a.LastUsedAt.Format("2006-01-02 15:04:05")
	r.LastUsedIP = a.LastUsedIP

	if a.EOSAccount != "" {
		r.IsAuth = true
	}
	return r
}

const (
	TimeDimensionMonth string = "month"
	TimeDimensionWeek  string = "week"
	TimeDimensionDay   string = "day"
)

const (
	DataTypeAddUserCount    string = "add_user_count"
	DataTypeAddPasskeyCount string = "add_passkey_count"
	DataTypeAddEvmCount     string = "add_evm_count"
	DateTypeAddDepositCount string = "add_deposit_count"
)
