package redislock

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)


func NewLocker(redisClient *redis.Client, key string, blockTime ...time.Duration) *Locker {
	defTime := time.Second * 5
	if len(blockTime) > 0 {
		defTime = blockTime[0]
	}
	return New(redisClient, key, &Options{LockTimeout: defTime})
}


func NewBlockLocker(redisClient *redis.Client, key string, blockTime ...time.Duration) *Locker {
	defTime := time.Second * 5
	if len(blockTime) > 0 {
		defTime = blockTime[0]
	}
	retryCount := CalcRetryTimes(int(defTime / time.Second))
	return New(redisClient, key, &Options{LockTimeout: defTime, RetryCount: retryCount})
}

func DoWithBlockLock(ctx context.Context, redisClient *redis.Client, fun func(), keys ...string) error {
	lock := NewBlockLocker(redisClient, "lock:"+strings.Join(keys, ":"))
	b, err := lock.Lock(ctx)
	defer func() {
		_ = lock.Unlock(ctx)
	}()
	if err != nil {
		return err
	}
	if b {
		fun()
	} else {
		return fmt.Errorf("get lock failed:%s", strings.Join(keys, ":"))
	}
	return nil
}
