// Package redis provides Redis token storage
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RefreshTokenRepository manages refresh tokens in Redis
type RefreshTokenRepository struct {
	client *redis.Client
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(client *redis.Client) *RefreshTokenRepository {
	return &RefreshTokenRepository{client: client}
}

// Store stores a refresh token
func (r *RefreshTokenRepository) Store(ctx context.Context, userID, tokenID string, ttl time.Duration) error {
	key := fmt.Sprintf("refresh_token:%s", userID)
	return r.client.Set(ctx, key, tokenID, ttl).Err()
}

// Get retrieves a refresh token
func (r *RefreshTokenRepository) Get(ctx context.Context, userID string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s", userID)
	return r.client.Get(ctx, key).Result()
}

// Delete removes a refresh token
func (r *RefreshTokenRepository) Delete(ctx context.Context, userID string) error {
	key := fmt.Sprintf("refresh_token:%s", userID)
	return r.client.Del(ctx, key).Err()
}

// Verify checks if a refresh token matches the stored value
func (r *RefreshTokenRepository) Verify(ctx context.Context, userID, tokenID string) (bool, error) {
	stored, err := r.Get(ctx, userID)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return stored == tokenID, nil
}
