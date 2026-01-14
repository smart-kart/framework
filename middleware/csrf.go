package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// CSRFProtection implements CSRF token validation
type CSRFProtection struct {
	tokens map[string]*csrfToken
	mu     sync.RWMutex
	ttl    time.Duration
}

type csrfToken struct {
	token     string
	createdAt time.Time
	userID    string
}

// NewCSRFProtection creates a new CSRF protection middleware
func NewCSRFProtection(ttl time.Duration) *CSRFProtection {
	csrf := &CSRFProtection{
		tokens: make(map[string]*csrfToken),
		ttl:    ttl,
	}

	// Start cleanup routine
	go csrf.cleanupRoutine()

	return csrf
}

// cleanupRoutine removes expired tokens
func (c *CSRFProtection) cleanupRoutine() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, token := range c.tokens {
			if now.Sub(token.createdAt) > c.ttl {
				delete(c.tokens, key)
			}
		}
		c.mu.Unlock()
	}
}

// GenerateToken generates a new CSRF token for a user
func (c *CSRFProtection) GenerateToken(userID string) (string, error) {
	// Generate random token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.URLEncoding.EncodeToString(b)

	// Store token
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tokens[token] = &csrfToken{
		token:     token,
		createdAt: time.Now(),
		userID:    userID,
	}

	return token, nil
}

// ValidateToken validates a CSRF token
func (c *CSRFProtection) ValidateToken(token, userID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	csrfToken, exists := c.tokens[token]
	if !exists {
		return false
	}

	// Check if token expired
	if time.Since(csrfToken.createdAt) > c.ttl {
		return false
	}

	// Check if token belongs to user
	if csrfToken.userID != userID {
		return false
	}

	return true
}

// InvalidateToken removes a CSRF token
func (c *CSRFProtection) InvalidateToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.tokens, token)
}

// UnaryServerInterceptor returns a gRPC interceptor for CSRF protection
func (c *CSRFProtection) UnaryServerInterceptor(protectedMethods []string) grpc.UnaryServerInterceptor {
	// Create method map for fast lookup
	methodMap := make(map[string]bool)
	for _, method := range protectedMethods {
		methodMap[method] = true
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip CSRF check for unprotected methods
		if !methodMap[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract CSRF token from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.InvalidArgument, "missing metadata")
		}

		csrfTokens := md.Get("x-csrf-token")
		if len(csrfTokens) == 0 {
			return nil, status.Error(codes.InvalidArgument, "missing CSRF token")
		}
		csrfToken := csrfTokens[0]

		// Extract user ID from context
		userID, ok := ctx.Value("user_id").(string)
		if !ok || userID == "" {
			return nil, status.Error(codes.Unauthenticated, "user not authenticated")
		}

		// Validate token
		if !c.ValidateToken(csrfToken, userID) {
			return nil, status.Error(codes.PermissionDenied, "invalid or expired CSRF token")
		}

		return handler(ctx, req)
	}
}

// HTTPMiddleware provides CSRF protection for HTTP endpoints
func (c *CSRFProtection) HTTPMiddleware(protectedPaths []string) func(next func(ctx context.Context, req interface{}) (interface{}, error)) func(ctx context.Context, req interface{}) (interface{}, error) {
	pathMap := make(map[string]bool)
	for _, path := range protectedPaths {
		pathMap[path] = true
	}

	return func(next func(ctx context.Context, req interface{}) (interface{}, error)) func(ctx context.Context, req interface{}) (interface{}, error) {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Extract path from context
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return next(ctx, req)
			}

			paths := md.Get(":path")
			if len(paths) == 0 {
				return next(ctx, req)
			}

			// Skip if path not protected
			if !pathMap[paths[0]] {
				return next(ctx, req)
			}

			// Extract and validate CSRF token
			csrfTokens := md.Get("x-csrf-token")
			if len(csrfTokens) == 0 {
				return nil, fmt.Errorf("missing CSRF token")
			}

			userID, ok := ctx.Value("user_id").(string)
			if !ok || userID == "" {
				return nil, fmt.Errorf("user not authenticated")
			}

			if !c.ValidateToken(csrfTokens[0], userID) {
				return nil, fmt.Errorf("invalid or expired CSRF token")
			}

			return next(ctx, req)
		}
	}
}
