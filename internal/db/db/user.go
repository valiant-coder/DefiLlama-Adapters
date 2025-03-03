package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

func init() {
	addMigrateFunc(func(r *Repo) error {
		return r.AutoMigrate(
			&User{},
			&UserCredential{},
		)
	})
}

type LoginMethod string

const (
	LoginMethodGoogle LoginMethod = "google"
	LoginMethodApple  LoginMethod = "apple"
)

type User struct {
	gorm.Model
	Username    string      `gorm:"column:username;type:varchar(255);not null;index:idx_username"`
	UID         string      `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid"`
	LoginMethod LoginMethod `gorm:"column:login_method;type:varchar(255);not null;uniqueIndex:idx_login_method_oauth_id"`
	Avatar      string      `gorm:"column:avatar;type:varchar(255);not null;default:''"`
	OauthID     string      `gorm:"column:oauth_id;type:varchar(255);not null;uniqueIndex:idx_login_method_oauth_id"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := r.Intn(90000000) + 10000000
	u.UID = fmt.Sprintf("%d", randomNum)
	return
}

func (r *Repo) UpsertUser(ctx context.Context, user *User) error {
	var existingUser User
	err := r.DB.WithContext(ctx).Where(
		"login_method = ? AND oauth_id = ?",
		user.LoginMethod,
		user.OauthID,
	).First(&existingUser).Error
	if err == nil {
		existingUser.Avatar = user.Avatar
		existingUser.Username = user.Username
		user = &existingUser
		return r.DB.WithContext(ctx).Save(user).Error
	}
	return r.DB.WithContext(ctx).Create(user).Error
}

func (r *Repo) IsUserExist(ctx context.Context, uid string) (bool, error) {
	var user User
	result := r.DB.WithContext(ctx).Where("uid = ?", uid).First(&user)
	return result.RowsAffected > 0, result.Error
}

func (r *Repo) GetUser(ctx context.Context, uid string) (User, error) {
	var user User
	result := r.DB.WithContext(ctx).Where("uid = ?", uid).First(&user)
	return user, result.Error
}

func (r *Repo) GetTotalUserCount(ctx context.Context) (int64, error) {
	var totalUserCount int64
	result := r.DB.WithContext(ctx).Model(&User{}).Count(&totalUserCount).Error
	if result != nil {
		return 0, result
	}
	return totalUserCount, nil
}

type UserCredential struct {
	gorm.Model
	UID            string    `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	CredentialID   string    `gorm:"column:credential_id;type:varchar(255);uniqueIndex:idx_credential_id"`
	PublicKey      string    `gorm:"column:public_key;type:varchar(255);not null;index:idx_public_key"`
	Name           string    `gorm:"column:name;type:varchar(255);not null"`
	LastUsedAt     time.Time `gorm:"column:last_used_at;default:null;type:timestamp"`
	LastUsedIP     string    `gorm:"column:last_used_ip;type:varchar(255)"`
	Synced         bool      `gorm:"column:synced;type:tinyint(1);not null;default:0"`
	EOSAccount     string    `gorm:"column:eos_account;type:varchar(255);index:idx_eos_account"`
	EOSPermissions string    `gorm:"column:eos_permissions;type:varchar(512)"`
	DeviceID       string    `gorm:"column:device_id;default:null;type:varchar(255);index:idx_device_id"`
	BlockNumber    uint64    `gorm:"column:block_number;default:0;type:bigint(20)"`
	AAGuid         string    `gorm:"column:aaguid;type:varchar(255);default:null"`
}

func (UserCredential) TableName() string {
	return "user_credentials"
}

func (r *Repo) CreateCredentialIfNotExist(ctx context.Context, credential *UserCredential) error {
	var existingCredential UserCredential
	result := r.DB.WithContext(ctx).Where("credential_id = ?", credential.CredentialID).First(&existingCredential)
	if result.Error == gorm.ErrRecordNotFound {
		return r.DB.WithContext(ctx).Create(credential).Error
	}
	return nil
}

func (r *Repo) GetUserCredentials(ctx context.Context, uid string) ([]UserCredential, error) {
	var credentials []UserCredential
	result := r.DB.WithContext(ctx).Where("uid = ?", uid).Find(&credentials)
	return credentials, result.Error
}

func (r *Repo) GetUserCredentialByPubkey(ctx context.Context, pubkey string) (*UserCredential, error) {
	var credential UserCredential
	result := r.DB.WithContext(ctx).Where("public_key = ?", pubkey).First(&credential)
	return &credential, result.Error
}

func (r *Repo) GetUserCredentialsByEOSAccount(ctx context.Context, eosAccount string) ([]*UserCredential, error) {
	var credentials []*UserCredential
	err := r.DB.WithContext(ctx).Where("eos_account = ?", eosAccount).Find(&credentials).Error
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func (r *Repo) GetEosAccountByUID(ctx context.Context, uid string) (string, error) {
	var credential UserCredential
	err := r.DB.WithContext(ctx).Where("uid = ? and eos_account != ''", uid).First(&credential).Error
	if err != nil {
		return "", err
	}
	return credential.EOSAccount, nil
}

func (r *Repo) GetUserCredentialsByKeys(ctx context.Context, keys []string) ([]*UserCredential, error) {
	var credentials []*UserCredential
	err := r.DB.WithContext(ctx).Where("public_key IN (?)", keys).Find(&credentials).Error
	if err != nil {
		return nil, err
	}
	return credentials, nil
}

func (r *Repo) UpdateUserCredential(ctx context.Context, credential *UserCredential) error {
	return r.DB.WithContext(ctx).Model(&UserCredential{}).Where("id = ?", credential.ID).Updates(credential).Error
}

func (r *Repo) DeleteUserCredential(ctx context.Context, credential *UserCredential) error {
	return r.DB.WithContext(ctx).Where("id = ?", credential.ID).Delete(&UserCredential{}).Error
}

func (r *Repo) GetUIDByEOSAccount(ctx context.Context, eosAccount string) (string, error) {
	var credential UserCredential
	result := r.DB.WithContext(ctx).Where("eos_account = ?", eosAccount).First(&credential)
	return credential.UID, result.Error
}

func (r *Repo) GetUserCredentialMaxBlockNumber(ctx context.Context) (uint64, error) {
	var blockNumber *uint64
	err := r.WithContext(ctx).Model(&UserCredential{}).Select("COALESCE(MAX(block_number), 0)").Scan(&blockNumber).Error
	if err != nil {
		return 0, err
	}
	if blockNumber == nil {
		return 0, nil
	}
	return *blockNumber, nil
}

type EOSAccountInfo struct {
	UID        string `gorm:"column:uid;"`
	EOSAccount string `gorm:"column:eos_account;"`
}

func (r *Repo) GetAllEOSAccounts(ctx context.Context) ([]EOSAccountInfo, error) {
	var accounts []EOSAccountInfo
	err := r.DB.WithContext(ctx).
		Model(&UserCredential{}).
		Select("DISTINCT uid, eos_account").
		Where("eos_account != ''").
		Find(&accounts).Error
	return accounts, err
}

func (r *Repo) GetUsersByUIDs(ctx context.Context, uids []string) (map[string]User, error) {
	var users []User
	err := r.DB.WithContext(ctx).Where("uid IN ?", uids).Find(&users).Error
	if err != nil {
		return nil, err
	}

	userMap := make(map[string]User)
	for _, user := range users {
		userMap[user.UID] = user
	}
	return userMap, nil
}
