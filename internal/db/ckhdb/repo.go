package ckhdb

import (
	"context"
	"exapp-go/config"
	"exapp-go/internal/db/plugins"
	"exapp-go/pkg/cache"
	"exapp-go/pkg/queryparams"
	"fmt"
	"os"
	"sync"
	"time"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	clickhouseDB *gorm.DB
	ckOnce       sync.Once
)

func ckhDB() *gorm.DB {
	ckOnce.Do(func() {
		config := config.Conf().ClickHouse

		dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%s/%s?dial_timeout=10s&read_timeout=20s", config.User, config.Pass, config.Host, config.Port, config.Database)
		var err error
		clickhouseDB, err = gorm.Open(clickhouse.Open(dsn), &gorm.Config{
			Logger:      logger.Default.LogMode(logger.Silent),
			PrepareStmt: false,
		})
		if err != nil {
			panic("failed to connect clickhouse database")
		}
		cachePlugin := plugins.NewCachePlugin(cache.DefaultStore())
		err = clickhouseDB.Use(cachePlugin)
		if err != nil {
			fmt.Printf("db use cache plugin err: %v\n", err)
			os.Exit(1)
		}

		sqlDB, err := clickhouseDB.DB()
		if err != nil {
			panic("failed to get database instance")
		}

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
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

	subQuery := db
	if err = subQuery.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return
	}

	if err = db.Limit(params.Limit).Offset(params.Offset).Order(params.Order).Find(models).Error; err != nil {
		return
	}

	return
}

func (r *ClickHouseRepo) WithCache(key string, expire time.Duration) *ClickHouseRepo {
	return r.Clone(r.DB.Set(plugins.CacheParamKey, plugins.CacheParam{
		Key:     key,
		Expires: expire,
	}))
}

func (r *ClickHouseRepo) Clone(db *gorm.DB) *ClickHouseRepo {
	return &ClickHouseRepo{
		DB: db,
	}
}
