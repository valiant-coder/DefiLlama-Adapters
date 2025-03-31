package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
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
	SID        string                      `gorm:"column:sid;type:varchar(255);default:null;uniqueIndex:idx_sid"`
	EOSAccount string                      `gorm:"column:eos_account;type:varchar(255);not null;index:idx_eos_account;uniqueIndex:idx_eos_account_permission"`
	Name       string                      `gorm:"column:name;type:varchar(255);not null;index:idx_uid_name"`
	Permission string                      `gorm:"column:permission;type:varchar(255);not null;uniqueIndex:idx_eos_account_permission"`
	APIKey     string                      `gorm:"column:api_key;type:varchar(255);not null;index:idx_api_key"`
	PublicKeys datatypes.JSONSlice[string] `gorm:"column:public_keys;type:json;not null"`
}

func (UserSubAccount) TableName() string {
	return "user_subaccounts"
}

func (u *UserSubAccount) BeforeCreate(tx *gorm.DB) (err error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := r.Intn(90000000) + 10000000
	u.SID = fmt.Sprintf("%d", randomNum)
	return
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
	err := r.WithCache(fmt.Sprintf("subaccount:%s", apiKey), 1*time.Hour).WithContext(ctx).Where("api_key = ?", apiKey).First(&subAccount).Error
	return &subAccount, err
}

// GetUserSubAccount retrieves a specific sub-account by UID and name
func (r *Repo) GetUserSubAccount(ctx context.Context, uid string, name string) (*UserSubAccount, error) {
	var subAccount UserSubAccount
	err := r.WithContext(ctx).Where("uid = ? AND name = ?", uid, name).First(&subAccount).Error
	return &subAccount, err
}

// DeleteUserSubAccount deletes a sub-account by UID and name
func (r *Repo) DeleteUserSubAccount(ctx context.Context, sid string) error {
	return r.WithContext(ctx).Where("sid = ?", sid).Delete(&UserSubAccount{}).Error
}

func (r *Repo) GetUserSubAccountByEOSAccountAndPermission(ctx context.Context, eosAccount string, permission string) (*UserSubAccount, error) {
	var subAccount UserSubAccount
	err := r.WithContext(ctx).Where("eos_account = ? AND permission = ?", eosAccount, permission).First(&subAccount).Error
	return &subAccount, err
}

func (r *Repo) UpdateUserSubAccount(ctx context.Context, subAccount *UserSubAccount) error {
	return r.WithContext(ctx).Model(&UserSubAccount{}).Where("id = ?", subAccount.ID).Updates(subAccount).Error
}

type UserSubAccountBalance struct {
	gorm.Model
	UID        string          `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	EOSAccount string          `gorm:"column:eos_account;type:varchar(255);not null;index:idx_eos_permission"`
	Permission string          `gorm:"column:permission;type:varchar(255);not null;index:idx_eos_permission"`
	Coin       string          `gorm:"column:coin;type:varchar(255);not null;index:idx_coin"`
	Balance    decimal.Decimal `gorm:"column:balance;type:decimal(36,18);not null"`
}

func (UserSubAccountBalance) TableName() string {
	return "user_subaccount_balances"
}

func (r *Repo) GetUserSubAccountBalance(ctx context.Context, eosAccount string, permission string) ([]*UserSubAccountBalance, error) {
	var balances []*UserSubAccountBalance
	err := r.WithContext(ctx).Where("eos_account = ? AND permission = ?", eosAccount, permission).Find(&balances).Error
	return balances, err
}
