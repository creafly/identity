package service

import (
	"context"
	"strconv"
	"time"

	"github.com/creafly/identity/internal/infra/redis"
)

const (
	loginAttemptPrefix = "login:attempts:"
	loginLockPrefix    = "login:locked:"
)

type LoginAttemptTracker interface {
	RecordFailedAttempt(identifier string)
	GetAttempts(identifier string) int
	IsLocked(identifier string) bool
	GetLockoutRemaining(identifier string) time.Duration
	ClearAttempts(identifier string)
}

type LoginAttemptConfig struct {
	MaxAttempts     int
	LockoutDuration time.Duration
	AttemptWindow   time.Duration
}

func DefaultLoginAttemptConfig() LoginAttemptConfig {
	return LoginAttemptConfig{
		MaxAttempts:     5,
		LockoutDuration: 15 * time.Minute,
		AttemptWindow:   15 * time.Minute,
	}
}

type redisLoginAttemptTracker struct {
	client *redis.Client
	config LoginAttemptConfig
}

func NewLoginAttemptTracker(client *redis.Client, config LoginAttemptConfig) LoginAttemptTracker {
	return &redisLoginAttemptTracker{
		client: client,
		config: config,
	}
}

func (t *redisLoginAttemptTracker) RecordFailedAttempt(identifier string) {
	if t.config.MaxAttempts <= 0 {
		return
	}

	ctx := context.Background()
	attemptKey := loginAttemptPrefix + identifier
	lockKey := loginLockPrefix + identifier

	locked, _ := t.client.Exists(ctx, lockKey)
	if locked {
		return
	}

	count, err := t.client.Incr(ctx, attemptKey)
	if err != nil {
		return
	}

	if count == 1 {
		_ = t.client.Expire(ctx, attemptKey, t.config.AttemptWindow)
	}

	if int(count) >= t.config.MaxAttempts {
		_ = t.client.Set(ctx, lockKey, "1", t.config.LockoutDuration)
		_ = t.client.Del(ctx, attemptKey)
	}
}

func (t *redisLoginAttemptTracker) GetAttempts(identifier string) int {
	if t.config.MaxAttempts <= 0 {
		return 0
	}

	ctx := context.Background()
	attemptKey := loginAttemptPrefix + identifier

	val, err := t.client.Get(ctx, attemptKey)
	if err != nil {
		return 0
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	return count
}

func (t *redisLoginAttemptTracker) IsLocked(identifier string) bool {
	if t.config.MaxAttempts <= 0 {
		return false
	}

	ctx := context.Background()
	lockKey := loginLockPrefix + identifier

	exists, err := t.client.Exists(ctx, lockKey)
	if err != nil {
		return false
	}

	return exists
}

func (t *redisLoginAttemptTracker) GetLockoutRemaining(identifier string) time.Duration {
	if t.config.MaxAttempts <= 0 {
		return 0
	}

	ctx := context.Background()
	lockKey := loginLockPrefix + identifier

	ttl, err := t.client.TTL(ctx, lockKey)
	if err != nil || ttl < 0 {
		return 0
	}

	return ttl
}

func (t *redisLoginAttemptTracker) ClearAttempts(identifier string) {
	ctx := context.Background()
	attemptKey := loginAttemptPrefix + identifier
	lockKey := loginLockPrefix + identifier

	_ = t.client.Del(ctx, attemptKey, lockKey)
}
