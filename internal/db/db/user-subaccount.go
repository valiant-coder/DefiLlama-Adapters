package db

import (
	"context"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(
			&UserSubAccount{},
		)
	})
}

type UserSubAccount struct {
	gorm.Model
	UID        string                      `gorm:"column:uid;type:varchar(255);not null;index:idx_uid_name"`
	EOSAccount string                      `gorm:"column:eos_account;type:varchar(255);not null;index:idx_eos_account"`
	Name       string                      `gorm:"column:name;type:varchar(255);not null;index:idx_uid_name"`
	Permission string                      `gorm:"column:permission;type:varchar(255);not null"`
	APIKey     string                      `gorm:"column:api_key;type:varchar(255);not null;index:idx_api_key"`
	PublicKeys datatypes.JSONSlice[string] `gorm:"column:public_keys;type:json;not null"`
}

func (UserSubAccount) TableName() string {
	return "user_subaccounts"
}

func (r *Repo) CreateUserSubAccount(ctx context.Context, userSubAccount *UserSubAccount) error {
	return r.WithContext(ctx).Create(userSubAccount).Error
}

// GetUserSubAccounts retrieves all sub-accounts for a user by UID
func (r *Repo) GetUserSubAccounts(ctx context.Context, uid string) ([]UserSubAccount, error) {
	var subAccounts []UserSubAccount
	err := r.WithContext(ctx).Where("uid = ?", uid).Find(&subAccounts).Error
	return subAccounts, err
}

// GetUserSubAccountByAPIKey retrieves a sub-account by its API key
func (r *Repo) GetUserSubAccountByAPIKey(ctx context.Context, apiKey string) (*UserSubAccount, error) {
	var subAccount UserSubAccount
	err := r.WithContext(ctx).Where("api_key = ?", apiKey).First(&subAccount).Error
	return &subAccount, err
}

// GetUserSubAccount retrieves a specific sub-account by UID and name
func (r *Repo) GetUserSubAccount(ctx context.Context, uid string, name string) (*UserSubAccount, error) {
	var subAccount UserSubAccount
	err := r.WithContext(ctx).Where("uid = ? AND name = ?", uid, name).First(&subAccount).Error
	return &subAccount, err
}

// DeleteUserSubAccount deletes a sub-account by UID and name
func (r *Repo) DeleteUserSubAccount(ctx context.Context, uid string, name string) error {
	return r.WithContext(ctx).Where("uid = ? AND name = ?", uid, name).Delete(&UserSubAccount{}).Error
}
