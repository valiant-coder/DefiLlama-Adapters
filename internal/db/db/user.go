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
	Username    string      `gorm:"column:username;type:varchar(255);not null;"`
	UID         string      `gorm:"column:uid;type:varchar(255);not null;uniqueIndex:idx_uid"`
	LoginMethod LoginMethod `gorm:"column:login_method;type:varchar(255);not null;uniqueIndex:idx_login_method_oauth_id"`
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

func (r *Repo) CreateUserIfNotExist(ctx context.Context, user *User) error {
	var existingUser User
	result := r.DB.WithContext(ctx).Where(
		"login_method = ? AND oauth_id = ?",
		user.LoginMethod,
		user.OauthID,
	).First(&existingUser)

	if result.Error == gorm.ErrRecordNotFound {
		if err := r.DB.WithContext(ctx).Create(user).Error; err != nil {
			if r.DB.WithContext(ctx).Where(
				"login_method = ? AND oauth_id = ?",
				user.LoginMethod,
				user.OauthID,
			).First(&existingUser).Error == nil {
				*user = existingUser
				return nil
			}
			return fmt.Errorf("create user failed: %w", err)
		}
		return nil
	}
	*user = existingUser
	return nil
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

type UserCredential struct {
	gorm.Model
	UID            string    `gorm:"column:uid;type:varchar(255);not null;index:idx_uid"`
	CredentialID   string    `gorm:"column:credential_id;type:varchar(255);uniqueIndex:idx_credential_id"`
	PublicKey      string    `gorm:"column:public_key;type:text;not null"`
	Name           string    `gorm:"column:name;type:varchar(255);not null"`
	LastUsedAt     time.Time `gorm:"column:last_used_at;default:null;type:timestamp"`
	LastUsedIP     string    `gorm:"column:last_used_ip;type:varchar(255)"`
	Synced         bool      `gorm:"column:synced;type:tinyint(1);not null;default:0"`
	EOSAccount     string    `gorm:"column:eos_account;type:varchar(255)"`
	EOSPermissions string    `gorm:"column:eos_permissions;type:varchar(512)"`
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
