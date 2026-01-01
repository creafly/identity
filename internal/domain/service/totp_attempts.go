package service

import (
	"context"
	"strconv"
	"time"

	"github.com/creafly/identity/internal/infra/redis"
)

const (
	totpAttemptPrefix = "totp:attempts:"
	totpLockPrefix    = "totp:locked:"
)

type TOTPAttemptTracker interface {
	RecordFailedAttempt(userID string)
	GetAttempts(userID string) int
	IsLocked(userID string) bool
	GetLockoutRemaining(userID string) time.Duration
	ClearAttempts(userID string)
}

type TOTPAttemptConfig struct {
	MaxAttempts     int
	LockoutDuration time.Duration
	AttemptWindow   time.Duration
}

func DefaultTOTPAttemptConfig() TOTPAttemptConfig {
	return TOTPAttemptConfig{
		MaxAttempts:     5,
		LockoutDuration: 5 * time.Minute,
		AttemptWindow:   5 * time.Minute,
	}
}

type redisTOTPAttemptTracker struct {
	client *redis.Client
	config TOTPAttemptConfig
}

func NewTOTPAttemptTracker(client *redis.Client, config TOTPAttemptConfig) TOTPAttemptTracker {
	return &redisTOTPAttemptTracker{
		client: client,
		config: config,
	}
}

func (t *redisTOTPAttemptTracker) RecordFailedAttempt(userID string) {
	if t.config.MaxAttempts <= 0 {
		return
	}

	ctx := context.Background()
	attemptKey := totpAttemptPrefix + userID
	lockKey := totpLockPrefix + userID

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

func (t *redisTOTPAttemptTracker) GetAttempts(userID string) int {
	if t.config.MaxAttempts <= 0 {
		return 0
	}

	ctx := context.Background()
	attemptKey := totpAttemptPrefix + userID

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

func (t *redisTOTPAttemptTracker) IsLocked(userID string) bool {
	if t.config.MaxAttempts <= 0 {
		return false
	}

	ctx := context.Background()
	lockKey := totpLockPrefix + userID

	exists, err := t.client.Exists(ctx, lockKey)
	if err != nil {
		return false
	}

	return exists
}

func (t *redisTOTPAttemptTracker) GetLockoutRemaining(userID string) time.Duration {
	if t.config.MaxAttempts <= 0 {
		return 0
	}

	ctx := context.Background()
	lockKey := totpLockPrefix + userID

	ttl, err := t.client.TTL(ctx, lockKey)
	if err != nil || ttl < 0 {
		return 0
	}

	return ttl
}

func (t *redisTOTPAttemptTracker) ClearAttempts(userID string) {
	ctx := context.Background()
	attemptKey := totpAttemptPrefix + userID
	lockKey := totpLockPrefix + userID

	_ = t.client.Del(ctx, attemptKey, lockKey)
}
