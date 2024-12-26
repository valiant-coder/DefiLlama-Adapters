package db

import (
	"context"
	"errors"
	"exapp-go/config"
	"exapp-go/internal/db/plugins"
	"exapp-go/pkg/cache"
	"exapp-go/pkg/queryparams"
	"exapp-go/pkg/utils"
	"fmt"
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
	rdb struct {
		isCluster bool
		single    *redis.Client
		cluster   *redis.ClusterClient
	}
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
		if redisCfg.IsCluster {
			repo.rdb.isCluster = true
			clusterRDB := redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    redisCfg.Cluster.Urls,
				Username: redisCfg.Cluster.User,
				Password: redisCfg.Cluster.Pass,
			})
			repo.rdb.cluster = clusterRDB

		} else {
			rdb := redis.NewClient(&redis.Options{
				Addr:     fmt.Sprintf("%s:%d", redisCfg.Host, redisCfg.Port),
				Password: redisCfg.Pass,
				DB:       redisCfg.DB,
			})
			repo.rdb.single = rdb
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

	err = db.AutoMigrate()

	if err != nil {
		return nil, err
	}

	return db, nil
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

func (r *Repo) CacheSet(ctx context.Context, key string, value interface{}, expire time.Duration) error {
	if r.rdb.isCluster {
		return r.rdb.cluster.Set(ctx, key, utils.MarshalE(value), expire).Err()
	} else {
		return r.rdb.single.Set(ctx, key, utils.MarshalE(value), expire).Err()
	}
}

func (r *Repo) CacheGet(ctx context.Context, key string, value interface{}) error {
	if r.rdb.isCluster {
		v, err := r.rdb.cluster.Get(ctx, key).Bytes()
		if err != nil {
			return err
		}
		return utils.Unmarshal(v, value)
	} else {
		v, err := r.rdb.single.Get(ctx, key).Bytes()
		if err != nil {
			return err
		}
		return utils.Unmarshal(v, value)
	}
}

func (r *Repo) CacheExist(ctx context.Context, key string) (bool, error) {
	if r.rdb.isCluster {
		result, err := r.rdb.cluster.Exists(ctx, key).Result()
		if err != nil {
			return false, err
		}
		return result > 0, nil
	} else {
		result, err := r.rdb.single.Exists(ctx, key).Result()
		if err != nil {
			return false, err
		}
		return result > 0, nil
	}
}

func (r *Repo) CacheDel(ctx context.Context, key string) error {
	if r.rdb.isCluster {
		return r.rdb.cluster.Del(ctx, key).Err()
	} else {
		return r.rdb.single.Del(ctx, key).Err()
	}
}
func (r *Repo) CacheSAdd(ctx context.Context, key string, members ...interface{}) error {
	if r.rdb.isCluster {
		return r.rdb.cluster.SAdd(ctx, key, members...).Err()
	} else {
		return r.rdb.single.SAdd(ctx, key, members...).Err()
	}
}
func (r *Repo) CacheSRem(ctx context.Context, key string, members ...interface{}) error {
	if r.rdb.isCluster {
		return r.rdb.cluster.SRem(ctx, key, members...).Err()
	} else {
		return r.rdb.single.SRem(ctx, key, members...).Err()
	}
}
func (r *Repo) CacheSPop(ctx context.Context, key string) (string, error) {
	if r.rdb.isCluster {
		return r.rdb.cluster.SPop(ctx, key).Result()
	} else {
		return r.rdb.single.SPop(ctx, key).Result()
	}
}
func (r *Repo) CacheSPopN(ctx context.Context, key string, count int64) ([]string, error) {
	if r.rdb.isCluster {
		return r.rdb.cluster.SPopN(ctx, key, count).Result()
	} else {
		return r.rdb.single.SPopN(ctx, key, count).Result()
	}
}
func (r *Repo) CacheSCard(ctx context.Context, key string) (int64, error) {
	if r.rdb.isCluster {
		return r.rdb.cluster.SCard(ctx, key).Result()
	} else {
		return r.rdb.single.SCard(ctx, key).Result()
	}
}

func (r *Repo) CacheZAdd(ctx context.Context, key string, members ...redis.Z) error {
	if r.rdb.isCluster {
		return r.rdb.cluster.ZAdd(ctx, key, members...).Err()
	} else {
		return r.rdb.single.ZAdd(ctx, key, members...).Err()
	}
}
func (r *Repo) CacheZRem(ctx context.Context, key string, members ...interface{}) error {
	if r.rdb.isCluster {
		return r.rdb.cluster.ZRem(ctx, key, members...).Err()
	} else {
		return r.rdb.single.ZRem(ctx, key, members...).Err()
	}
}
func (r *Repo) GetZScore(ctx context.Context, key string, member string) (float64, error) {
	if r.rdb.isCluster {
		return r.rdb.cluster.ZScore(ctx, key, member).Result()
	} else {
		return r.rdb.single.ZScore(ctx, key, member).Result()
	}
}

func (r *Repo) Watch(ctx context.Context, f func(tx *redis.Tx) error, keys ...string) error {
	if r.rdb.isCluster {
		return r.rdb.cluster.Watch(ctx, f, keys...)
	} else {
		return r.rdb.single.Watch(ctx, f, keys...)
	}
}

func (r *Repo) RedisClient() *redis.Client {
	return r.rdb.single
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
