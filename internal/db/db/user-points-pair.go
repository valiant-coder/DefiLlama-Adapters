package db

import (
	"context"
	"exapp-go/data"

	"gorm.io/gorm"
)

func init() {

	addMigrateFunc(func(r *Repo) error {

		return r.AutoMigrate(&UserPointsPair{})
	})
}

type UserPointsPair struct {
	gorm.Model
	Pair        string `gorm:"column:pair;type:varchar(255);not null;uniqueIndex:idx_pair"`
	Coefficient uint64 `gorm:"column:coefficient;type:bigint(20);not null;default:1"`
}

func UserPointsPairRedisKey(pair string) string {

	return "user_points_pair:detail:" + pair
}

func (p *UserPointsPair) RedisKey() string {

	return UserPointsPairRedisKey(p.Pair)
}

type UserPointsPairListRes ListResult[UserPointsPair]

func (r *Repo) ListUserPointsPair(ctx context.Context, params data.UserPointsPairListParam) (*UserPointsPairListRes, error) {

	res, err := List[data.UserPointsPairListParam, UserPointsPair](params)
	if err != nil {

		return nil, err
	}

	result := UserPointsPairListRes(res)
	return &result, nil
}

func (r *Repo) GetUserPointsPair(ctx context.Context, pair string) (*UserPointsPair, error) {

	return Get[UserPointsPair](&UserPointsPair{Pair: pair})
}

func (r *Repo) CreateUserPointsPair(ctx context.Context, pair *UserPointsPair) error {

	return nil
}
