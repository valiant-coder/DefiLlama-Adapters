package db

import (
	"time"

	"gorm.io/gorm"
)

func init() {

	addMigrateFunc(func(r *Repo) error {

		return r.AutoMigrate(&UserPointsGrant{})
	})
}

type GrantStatus uint8

const (
	GrantStatusPending GrantStatus = iota
	GrantStatusAccept
	GrantStatusReject
	GrantStatusCancel
)

type UserPointsGrant struct {
	gorm.Model
	GrantTime time.Time   `json:"grant_time" gorm:"column:grant_time;type:bigint(20);not null"`
	Admin     string      `json:"admin" gorm:"column:admin;type:varchar(255);not null"`
	UID       string      `json:"uid" gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	Amount    uint64      `json:"amount" gorm:"column:amount;type:bigint(20);not null"`
	Status    GrantStatus `json:"status" gorm:"column:status;type:int(11);not null;default:0"`
}

func (*UserPointsGrant) TableName() string {
	return "user_points_grant"
}
