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
		return r.AutoMigrate(&User{})
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
	OauthID     string      `gorm:"column:oauth_id;type:varchar(255);not null;uniqueIndex:idx_login_method_oauth_id	"`
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

type UserCredential struct {
	gorm.Model
	UID          string `gorm:"column:uid;type:varchar(255);not null;"`
	CredentialID string `gorm:"column:credential_id;type:text;not null;uniqueIndex:idx_credential_id"`
	PublicKey    string `gorm:"column:public_key;type:text;not null"`
}

func (UserCredential) TableName() string {
	return "td_user_credentials"
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
	result := r.DB.WithContext(ctx).Where("user_id = ?", uid).Find(&credentials)
	return credentials, result.Error
}
