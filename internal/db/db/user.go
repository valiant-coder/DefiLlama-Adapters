package db

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

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
	return "td_users"
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
