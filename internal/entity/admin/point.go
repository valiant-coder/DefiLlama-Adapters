package entity_admin

import (
	"exapp-go/internal/db/db"
	"exapp-go/internal/entity"
)

type ReqUserPointsGrant struct {
	UID    string `json:"uid"`
	Amount uint64 `json:"amount"`
}

type RespUserPointsGrant struct {
	ID        uint        `json:"id"`
	CreatedAt entity.Time `json:"created_at"`
	Admin     string      `json:"admin"`
	GrantTime entity.Time `json:"grant_time"`
	UID       string      `json:"uid"`
	Amount    uint64      `json:"amount"`
	Status    uint8       `json:"status"`
}

func (r *RespUserPointsGrant) Fill(a *db.UserPointsGrant) *RespUserPointsGrant {
	r.UID = a.UID
	r.Amount = a.Amount
	r.GrantTime = entity.Time(a.GrantTime)
	r.Status = uint8(a.Status)
	r.CreatedAt = entity.Time(a.CreatedAt)
	r.ID = a.ID
	r.Admin = a.Admin
	return r
}

type ReqUpdateUserPointsGrantStatus struct {
	Status uint8 `json:"status"`
}

type ReqBatchUserPointsGrantAccept struct {
	IDs []uint `json:"ids"`
}
