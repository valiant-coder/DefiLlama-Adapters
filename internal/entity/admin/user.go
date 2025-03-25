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
