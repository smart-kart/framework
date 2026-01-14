package middleware

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	limiters map[string]*bucket
	mu       sync.RWMutex
	rate     int           // requests per window
	window   time.Duration // time window
	cleanup  time.Duration // cleanup interval
}

type bucket struct {
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: maximum number of requests per window
// window: time window for rate limiting (e.g., 15 minutes)
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*bucket),
		rate:     rate,
		window:   window,
		cleanup:  window * 2, // cleanup stale entries
	}

	// Start cleanup goroutine
	go rl.cleanupRoutine()

	return rl
}

// cleanupRoutine periodically removes stale rate limit entries
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.limiters {
			b.mu.Lock()
			if now.Sub(b.lastRefill) > rl.cleanup {
				delete(rl.limiters, key)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// getBucket returns or creates a bucket for the given key
func (rl *RateLimiter) getBucket(key string) *bucket {
	rl.mu.RLock()
	b, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		return b
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if b, exists := rl.limiters[key]; exists {
		return b
	}

	b = &bucket{
		tokens:     rl.rate,
		lastRefill: time.Now(),
	}
	rl.limiters[key] = b
	return b
}

// allow checks if a request should be allowed
func (rl *RateLimiter) allow(key string) bool {
	b := rl.getBucket(key)
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill)

	// Refill bucket if window has passed
	if elapsed >= rl.window {
		b.tokens = rl.rate
		b.lastRefill = now
	}

	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// extractIdentifier extracts IP address or user identifier from context
// SECURITY: Returns error instead of fallback to prevent rate limit bypass
// where all unidentified users would share the same bucket ("ip:unknown")
func extractIdentifier(ctx context.Context) (string, error) {
	// PRIORITY 1: Try to get user ID from metadata (set by auth middleware)
	// This is the most reliable identifier for authenticated requests
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userIDs := md.Get("user_id"); len(userIDs) > 0 && userIDs[0] != "" {
			return fmt.Sprintf("user:%s", userIDs[0]), nil
		}
	}

	// PRIORITY 2: Try to get user ID from context (legacy support)
	if userID, ok := ctx.Value("user_id").(string); ok && userID != "" {
		return fmt.Sprintf("user:%s", userID), nil
	}

	// PRIORITY 3: Get IP address from trusted headers
	// Note: X-Forwarded-For can be spoofed, so we only use X-Real-IP from trusted proxy
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// X-Real-IP is set by trusted reverse proxies (nginx, cloudflare)
		// This is more reliable than X-Forwarded-For for rate limiting
		if xrip := md.Get("x-real-ip"); len(xrip) > 0 && xrip[0] != "" {
			return fmt.Sprintf("ip:%s", xrip[0]), nil
		}

		// Fallback to X-Forwarded-For first IP (client IP)
		// Note: This can be spoofed, but better than nothing for unauthenticated requests
		if xff := md.Get("x-forwarded-for"); len(xff) > 0 && xff[0] != "" {
			// X-Forwarded-For format: "client, proxy1, proxy2"
			// We only want the first IP (client IP)
			clientIP := strings.Split(xff[0], ",")[0]
			clientIP = strings.TrimSpace(clientIP)
			if clientIP != "" {
				return fmt.Sprintf("ip:%s", clientIP), nil
			}
		}
	}

	// SECURITY: No fallback to "ip:unknown" - reject request instead
	// This prevents all unidentified users from sharing the same rate limit bucket
	// In production, ALL requests should have either user_id or IP from proxy
	return "", fmt.Errorf("unable to identify client: no user_id or IP address found")
}

// UnaryServerInterceptor returns a gRPC unary server interceptor for rate limiting
func (rl *RateLimiter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract identifier (IP or user ID)
		// SECURITY: Reject requests that cannot be identified to prevent rate limit bypass
		identifier, err := extractIdentifier(ctx)
		if err != nil {
			return nil, status.Errorf(
				codes.FailedPrecondition,
				"unable to identify client for rate limiting: %v", err,
			)
		}

		key := fmt.Sprintf("%s:%s", info.FullMethod, identifier)

		// Check rate limit
		if !rl.allow(key) {
			return nil, status.Errorf(
				codes.ResourceExhausted,
				"rate limit exceeded: maximum %d requests per %v",
				rl.rate,
				rl.window,
			)
		}

		// Continue with request
		return handler(ctx, req)
	}
}

// MethodRateLimiter allows different rate limits for different methods
type MethodRateLimiter struct {
	limiters map[string]*RateLimiter
	default_ *RateLimiter
}

// NewMethodRateLimiter creates a method-specific rate limiter
func NewMethodRateLimiter(defaultRate int, defaultWindow time.Duration) *MethodRateLimiter {
	return &MethodRateLimiter{
		limiters: make(map[string]*RateLimiter),
		default_: NewRateLimiter(defaultRate, defaultWindow),
	}
}

// AddMethodLimit adds a specific rate limit for a method
func (mrl *MethodRateLimiter) AddMethodLimit(method string, rate int, window time.Duration) {
	mrl.limiters[method] = NewRateLimiter(rate, window)
}

// UnaryServerInterceptor returns a gRPC unary server interceptor
func (mrl *MethodRateLimiter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Get rate limiter for this method
		limiter := mrl.default_
		if methodLimiter, exists := mrl.limiters[info.FullMethod]; exists {
			limiter = methodLimiter
		}

		// Extract identifier
		// SECURITY: Reject requests that cannot be identified to prevent rate limit bypass
		identifier, err := extractIdentifier(ctx)
		if err != nil {
			return nil, status.Errorf(
				codes.FailedPrecondition,
				"unable to identify client for rate limiting: %v", err,
			)
		}

		key := fmt.Sprintf("%s:%s", info.FullMethod, identifier)

		// Check rate limit
		if !limiter.allow(key) {
			return nil, status.Errorf(
				codes.ResourceExhausted,
				"rate limit exceeded: maximum %d requests per %v",
				limiter.rate,
				limiter.window,
			)
		}

		// Continue with request
		return handler(ctx, req)
	}
}
