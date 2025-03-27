package db

import (
	"context"
	"exapp-go/data"
	"fmt"

	"gorm.io/gorm"
)

func init() {

	addMigrateFunc(func(r *Repo) error {

		return r.AutoMigrate(&UserPointsPair{})
	})
}

type UserPointsPair struct {
	gorm.Model
	PoolId      uint64 `json:"pool_id" gorm:"column:pool_id;type:bigint(20);not null;uniqueIndex:idx_pool_id"`
	BaseCoin    string `json:"base_coin" gorm:"column:base_coin;type:varchar(255);not null"`
	QuoteCoin   string `json:"quote_coin" gorm:"column:quote_coin;type:varchar(255);not null"`
	Coefficient uint64 `json:"coefficient" gorm:"column:coefficient;type:bigint(20);not null;default:1"`
}

func UserPointsPairRedisKey(poolId uint64) string {

	return "user_points_pair:detail:" + fmt.Sprintf("%d", poolId)
}

func (p *UserPointsPair) RedisKey() string {

	return UserPointsPairRedisKey(p.PoolId)
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

func (r *Repo) GetUserPointsPair(ctx context.Context, poolId uint64) (*UserPointsPair, error) {

	return Get[UserPointsPair](&UserPointsPair{PoolId: poolId})
}

func (r *Repo) CreateUserPointsPair(ctx context.Context, pair *UserPointsPair) error {

	return nil
}
