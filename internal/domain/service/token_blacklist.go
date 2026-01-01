package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/creafly/identity/internal/infra/redis"
)

const (
	tokenBlacklistPrefix = "token:blacklist:"
	userRevokePrefix     = "user:revoke:"
)

type TokenBlacklist interface {
	Add(tokenHash string, expiresAt time.Time)
	IsBlacklisted(tokenHash string) bool
	RevokeAllForUser(userID string, expiresAt time.Time)
	IsUserRevoked(userID string) bool
}

func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

type redisBlacklist struct {
	client *redis.Client
}

func NewTokenBlacklist(client *redis.Client) TokenBlacklist {
	return &redisBlacklist{client: client}
}

func (bl *redisBlacklist) Add(tokenHash string, expiresAt time.Time) {
	ctx := context.Background()
	key := tokenBlacklistPrefix + tokenHash
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return
	}
	_ = bl.client.Set(ctx, key, "1", ttl)
}

func (bl *redisBlacklist) IsBlacklisted(tokenHash string) bool {
	ctx := context.Background()
	key := tokenBlacklistPrefix + tokenHash
	exists, err := bl.client.Exists(ctx, key)
	if err != nil {
		return false
	}
	return exists
}

func (bl *redisBlacklist) RevokeAllForUser(userID string, expiresAt time.Time) {
	ctx := context.Background()
	key := userRevokePrefix + userID
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return
	}
	_ = bl.client.Set(ctx, key, fmt.Sprintf("%d", time.Now().Unix()), ttl)
}

func (bl *redisBlacklist) IsUserRevoked(userID string) bool {
	ctx := context.Background()
	key := userRevokePrefix + userID
	exists, err := bl.client.Exists(ctx, key)
	if err != nil {
		return false
	}
	return exists
}
