package services

import (
	"context"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
)

const (
	// RateLimitMaxAttempts is the maximum number of login attempts allowed
	RateLimitMaxAttempts = 5
	// RateLimitWindowMinutes is the time window for rate limiting in minutes
	RateLimitWindowMinutes = 15
)

type RateLimitService struct {
	client valkey.Client
	window time.Duration
	max    int
}

func NewRateLimitService(client valkey.Client) *RateLimitService {
	return &RateLimitService{
		client: client,
		window: RateLimitWindowMinutes * time.Minute,
		max:    RateLimitMaxAttempts,
	}
}

// rateLimitKey generates the Redis/Valkey key for rate limiting
func (s *RateLimitService) rateLimitKey(identifier string) string {
	return fmt.Sprintf("ratelimit:admin_login:%s", identifier)
}

// CheckRateLimit checks if the identifier has exceeded the rate limit
// Returns:
//   - allowed: true if the request is allowed
//   - remaining: number of remaining attempts
//   - retryAfter: seconds until rate limit resets (only relevant if not allowed)
//   - error: any error that occurred
func (s *RateLimitService) CheckRateLimit(ctx context.Context, identifier string) (allowed bool, remaining int, retryAfter int64, err error) {
	key := s.rateLimitKey(identifier)

	// Increment the counter
	incrCmd := s.client.B().Incr().Key(key).Build()
	count, err := s.client.Do(ctx, incrCmd).ToInt64()
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to increment rate limit counter: %w", err)
	}

	// If this is the first attempt, set the expiration
	if count == 1 {
		expireCmd := s.client.B().Expire().Key(key).Seconds(int64(s.window.Seconds())).Build()
		err = s.client.Do(ctx, expireCmd).Error()
		if err != nil {
			return false, 0, 0, fmt.Errorf("failed to set rate limit expiration: %w", err)
		}
	}

	// Check if rate limit exceeded
	if count > int64(s.max) {
		// Get TTL for retry-after header
		ttlCmd := s.client.B().Ttl().Key(key).Build()
		ttl, err := s.client.Do(ctx, ttlCmd).ToInt64()
		if err != nil {
			ttl = int64(s.window.Seconds())
		}
		return false, 0, ttl, nil
	}

	remaining = s.max - int(count)
	return true, remaining, 0, nil
}

// ResetRateLimit resets the rate limit counter for an identifier
// This should be called after a successful login
func (s *RateLimitService) ResetRateLimit(ctx context.Context, identifier string) error {
	key := s.rateLimitKey(identifier)
	delCmd := s.client.B().Del().Key(key).Build()
	err := s.client.Do(ctx, delCmd).Error()
	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}
	return nil
}
