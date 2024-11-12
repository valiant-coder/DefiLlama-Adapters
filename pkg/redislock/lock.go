package redislock

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// src from:https://github.com/bsm/redis-lock
var luaRefresh = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("pexpire", KEYS[1], ARGV[2]) else return 0 end`)
var luaRelease = redis.NewScript(`if redis.call("get", KEYS[1]) == ARGV[1] then return redis.call("del", KEYS[1]) else return 0 end`)

// ErrLockNotObtained may be returned by Obtain() and Run()
// if a lock could not be obtained.
var (
	ErrLockUnlockFailed     = errors.New("lock unlock failed")
	ErrLockNotObtained      = errors.New("lock not obtained")
	ErrLockDurationExceeded = errors.New("lock duration exceeded")
)

// Locker allows (repeated) distributed locking.
type Locker struct {
	client *redis.Client
	key    string
	opts   Options

	token string
	mutex sync.Mutex

	timer  *time.Timer
	cancel chan struct{}
}

// Options describe the options for the lock
type Options struct {
	// The maximum duration to lock a key for
	// Default: 5s
	LockTimeout time.Duration

	// The number of time the acquisition of a lock will be retried.
	// Default: 0 = do not retry
	RetryCount int
}

func (o *Options) normalize() *Options {
	if o.LockTimeout < 1 {
		o.LockTimeout = 5 * time.Second
	}
	if o.RetryCount < 0 {
		o.RetryCount = 0
	}
	return o
}

// New creates a new distributed locker on a given key.
func New(client *redis.Client, key string, opts *Options) *Locker {
	var o Options
	if opts != nil {
		o = *opts
	}
	o.normalize()

	return &Locker{client: client, key: key, opts: o}
}

// IsLocked returns true if a lock is still being held.
func (l *Locker) IsLocked() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	return l.token != ""
}

func (l *Locker) IsLockedStrict(ctx context.Context) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.token == "" {
		return false
	}
	token, _ := l.client.Get(ctx, l.key).Result()
	return l.token == token
}

// Lock applies the lock, don't forget to defer the Unlock() function to release the lock after usage.
func (l *Locker) Lock(ctx context.Context) (b bool, err error) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.token != "" {
		if l.cancel != nil {
			close(l.cancel)
		}
		b, err = l.refresh(ctx)
	} else {
		b, err = l.create(ctx)
	}
	return
}

// Unlock releases the lock
func (l *Locker) Unlock(ctx context.Context) error {
	l.mutex.Lock()
	err := l.release(ctx)
	if l.cancel != nil {
		close(l.cancel)
	}
	l.reset()
	l.mutex.Unlock()
	return err
}

func (l *Locker) GetKey() string {
	return l.key
}

func (l *Locker) autoExpire() {
	l.mutex.Lock()
	if l.timer == nil {
		l.timer = time.NewTimer(l.opts.LockTimeout)
	} else {
		l.timer.Reset(l.opts.LockTimeout)
	}
	l.cancel = make(chan struct{})
	l.mutex.Unlock()

	select {
	case <-l.cancel:
		//
	case <-l.timer.C:
		l.resetWithLock()
	}
	l.mutex.Lock()
	l.cancel = nil
	l.mutex.Unlock()
}

func (l *Locker) create(ctx context.Context) (bool, error) {
	l.reset()

	// NewCache a random token
	token, err := randomToken()
	if err != nil {
		return false, err
	}

	attempts := l.opts.RetryCount + 1

	for {

		ok, err := l.obtain(ctx, token)
		if err != nil {
			return false, err
		} else if ok {
			l.token = token
			go l.autoExpire()
			return true, nil
		}

		if attempts--; attempts <= 0 {
			return false, nil
		}

		time.Sleep(time.Duration(sleepTime(l.opts.RetryCount-attempts-1)) * time.Millisecond)
	}
}

func (l *Locker) refresh(ctx context.Context) (bool, error) {
	ttl := strconv.FormatInt(int64(l.opts.LockTimeout/time.Millisecond), 10)
	status, err := luaRefresh.Run(ctx, l.client, []string{l.key}, l.token, ttl).Result()
	if err != nil {
		return false, err
	} else if status == int64(1) {
		return true, nil
	}
	return l.create(ctx)
}

func (l *Locker) obtain(ctx context.Context, token string) (bool, error) {
	ok, err := l.client.SetNX(ctx, l.key, token, l.opts.LockTimeout).Result()
	if err == redis.Nil {
		err = nil
	}
	return ok, err
}

func (l *Locker) release(ctx context.Context) error {
	defer l.reset()
	res, err := luaRelease.Run(ctx, l.client, []string{l.key}, l.token).Result()
	if err == redis.Nil {
		return ErrLockUnlockFailed
	}

	if i, ok := res.(int64); !ok || i != 1 {
		return ErrLockUnlockFailed
	}

	return err
}

func (l *Locker) reset() {
	l.token = ""
}

func (l *Locker) resetWithLock() {
	l.mutex.Lock()
	l.token = ""
	l.mutex.Unlock()
}

func randomToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

func sleepTime(count int) int {
	if count <= 1 {
		return 2
	}
	if count > 10 {
		return 100
	} else {
		return count * count
	}
}

func CalcRetryTimes(expireTime int) int {
	sum := 0
	count := 1
	for sum <= expireTime*1000 {
		sum += sleepTime(count)
		count++
	}
	return count
}
