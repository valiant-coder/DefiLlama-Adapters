package ckhdb

import (
	"context"
	"fmt"
	"exapp-go/config"
	"exapp-go/pkg/queryparams"
	"sync"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var clickhouseDB *gorm.DB
var ckOnce sync.Once

func ckhDB() *gorm.DB {
	ckOnce.Do(func() {
		config := config.Conf().ClickHouse

		dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?dial_timeout=10s&read_timeout=20s", config.User, config.Pass, config.Host, config.Port, config.Database)
		var err error
		clickhouseDB, err = gorm.Open(clickhouse.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		if err != nil {
			panic("failed to connect clickhouse database")
		}
	})
	return clickhouseDB
}

type ClickHouseRepo struct {
	*gorm.DB
}

func New() *ClickHouseRepo {
	return &ClickHouseRepo{
		DB: ckhDB(),
	}
}

func (c *ClickHouseRepo) Query(ctx context.Context, models interface{}, params *queryparams.QueryParams, queryable ...string) (total int64, err error) {
	if len(params.Order) == 0 {
		params.Order = "id desc"
	}

	db := c.DB.WithContext(ctx)

	if params.Select != "" {
		db = db.Select(params.Select)
	}

	if len(params.Joins) > 0 {
		db = db.Joins(params.Joins)
	}

	if len(params.Group) > 0 {
		db = db.Group(params.Group)
	}

	if len(params.Having) > 0 {
		db = db.Having(params.Having)
	}

	plains := params.Query.Plains(queryable...)
	if len(plains) > 0 {
		db = db.Where(plains[0], plains[1:]...)
	}

	if len(params.CustomQuery) != 0 {
		for queryStr, queryValue := range params.CustomQuery {
			if len(queryValue) == 0 {
				db = db.Where(queryStr)
			} else {
				db = db.Where(queryStr, queryValue...)
			}
		}
	}
	if len(params.Preload) > 0 {
		for _, populate := range params.Preload {
			db = db.Preload(populate)
		}
	}
	if len(params.TableName) > 0 {
		db = db.Table(params.TableName)
	}
	err = db.Limit(-1).Offset(-1).Count(&total).Error
	if err != nil {
		return
	}

	err = db.Limit(params.Limit).Offset(params.Offset).Order(params.Order).Find(models).Error
	if err != nil {
		return
	}

	return
}
