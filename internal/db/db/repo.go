package db

import (
	"context"
	"crypto/tls"
	"errors"
	"exapp-go/config"
	"exapp-go/internal/db/plugins"
	"exapp-go/pkg/cache"
	"exapp-go/pkg/queryparams"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Repo struct {
	*gorm.DB
	redis        redis.Cmdable
	redisCluster bool
	mu           sync.Mutex
}

var repo *Repo
var once sync.Once

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func New() *Repo {
	once.Do(func() {
		mysqlCfg := config.Conf().Mysql
		db, err := dbConnect(mysqlCfg.User, mysqlCfg.Pass, mysqlCfg.Host, mysqlCfg.Port, mysqlCfg.Database, mysqlCfg.Loc)
		if err != nil {
			fmt.Printf("db connect err: %v\n", err)
			os.Exit(1)
		}

		cachePlugin := plugins.NewCachePlugin(cache.DefaultStore())
		err = db.Use(cachePlugin)
		if err != nil {
			fmt.Printf("db use cache plugin err: %v\n", err)
			os.Exit(1)
		}
		repo = &Repo{
			DB: db,
		}
		redisCfg := config.Conf().Redis
		repo.redisCluster = redisCfg.IsCluster
		if redisCfg.IsCluster {
			clusterRDB := redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    redisCfg.Cluster.Urls,
				Username: redisCfg.Cluster.User,
				Password: redisCfg.Cluster.Pass,
				TLSConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			})
			repo.redis = clusterRDB
			if err := clusterRDB.Ping(context.Background()).Err(); err != nil {
				fmt.Printf("redis cluster connect err: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("redis cluster connect success")
		} else {
			rdb := redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port),
				Password: redisCfg.Pass,
				DB:       redisCfg.DB,
			})
			repo.redis = rdb
			if err := rdb.Ping(context.Background()).Err(); err != nil {
				fmt.Printf("redis connect err: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("redis connect success")
		}
		if len(migrateFuncs) > 0 && mysqlCfg.Migrate {
			for _, f := range migrateFuncs {
				err := f(repo)
				if err != nil {
					fmt.Printf("db migrate err: %v\n", err)
					os.Exit(1)
				}
			}
		}

		fmt.Println("db connect success")
		go repo.ensureConnection(context.Background())
	})

	return repo

}

func dbConnect(user, pass, host, port, dbName, loc string) (*gorm.DB, error) {
	loc = url.QueryEscape(loc)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=%s",
		user,
		pass,
		host,
		port,
		dbName,
		loc)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		return nil, err
	}

	db.Set("gorm:table_options", "CHARSET=utf8mb4")

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Minute * 10)

	err = db.AutoMigrate()

	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *Repo) ensureConnection(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.checkAndReconnect(); err != nil {
				log.Printf("Database reconnection failed: %v", err)
			}
		}
	}
}

func (r *Repo) checkAndReconnect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	sqlDB, err := r.DB.DB()
	if err != nil {
		return fmt.Errorf("Failed to get database instance: %w", err)
	}

	if err := sqlDB.PingContext(context.Background()); err != nil {
		_ = sqlDB.Close()

		mysqlCfg := config.Conf().Mysql
		newDB, err := dbConnect(mysqlCfg.User, mysqlCfg.Pass, mysqlCfg.Host, mysqlCfg.Port, mysqlCfg.Database, mysqlCfg.Loc)
		if err != nil {
			return fmt.Errorf("Database reconnection failed: %w", err)
		}
		r.DB = newDB
	}
	return nil
}

func (r *Repo) Update(ctx context.Context, model interface{}) (err error) {
	return r.DB.WithContext(ctx).Updates(model).Error
}

func (r *Repo) Delete(ctx context.Context, model interface{}) (err error) {
	return r.DB.WithContext(ctx).Delete(model).Error
}

func (r *Repo) Insert(ctx context.Context, model interface{}) (err error) {
	return r.DB.WithContext(ctx).Create(model).Error
}

func (r *Repo) Get(ctx context.Context, id uint, model interface{}) (err error) {
	err = r.DB.WithContext(ctx).Where("id = ?", id).First(model).Error
	return
}

func (r *Repo) Preload(ctx context.Context, id uint, q *queryparams.QueryParams, model interface{}) (err error) {
	db := r.DB.WithContext(ctx)
	for _, preload := range q.Preload {
		db = db.Preload(preload)
	}
	err = db.Where("id = ?", id).Find(model).Error
	return
}

func (r *Repo) List(ctx context.Context, models interface{}) error {
	return r.DB.WithContext(ctx).Find(models).Error
}

func (r *Repo) Clone(db *gorm.DB) *Repo {
	return &Repo{
		DB: db,
	}
}

func (r *Repo) WithContext(ctx context.Context) *Repo {
	return &Repo{
		DB: r.DB.WithContext(ctx),
	}
}

func (r *Repo) Debug() *Repo {
	return &Repo{
		DB: r.DB.Debug(),
	}
}

func (r *Repo) WithCache(key string, expire time.Duration) *Repo {
	return r.Clone(r.DB.Set(plugins.CacheParamKey, plugins.CacheParam{
		Key:     key,
		Expires: expire,
	}))
}

func (r *Repo) ExecSQL(ctx context.Context, sql string, values ...interface{}) (err error) {
	db := r.DB.Exec(sql, values...)
	if db.Error != nil {
		err = db.Error
		return
	}
	if db.RowsAffected == 0 {
		err = errors.New("no rows affected")
	}
	return

}

func (r *Repo) Transaction(ctx context.Context, f func(repo *Repo) error) (err error) {
	return r.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return f(&Repo{DB: tx})
	})
}

func (r *Repo) Query(ctx context.Context, models interface{}, params *queryparams.QueryParams, queryable ...string) (total int64, err error) {
	if len(params.Order) == 0 {
		params.Order = "id desc"
	}

	db := r.DB.WithContext(ctx).Limit(params.Limit).Offset(params.Offset).Order(params.Order)

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

	err = db.Find(models).Error
	if err != nil {
		return
	}

	err = db.Limit(-1).Offset(-1).Count(&total).Error
	return
}

func (r *Repo) Watch(ctx context.Context, f func(tx *redis.Tx) error, keys ...string) error {
	if r.redisCluster {
		return r.redis.(*redis.ClusterClient).Watch(ctx, f, keys...)
	} else {
		return r.redis.(*redis.Client).Watch(ctx, f, keys...)
	}
}

func (r *Repo) Redis() redis.Cmdable {
	return r.redis
}

type MigrateFunc func(r *Repo) error

var migrateFuncs []MigrateFunc
var migrateOnce sync.Once

func addMigrateFunc(f MigrateFunc) {
	migrateOnce.Do(func() {
		migrateFuncs = make([]MigrateFunc, 0)
	})
	migrateFuncs = append(migrateFuncs, f)
}
