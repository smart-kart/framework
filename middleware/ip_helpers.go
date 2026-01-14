package middleware

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/metadata"
)

var (
	// ErrIPNotFound is returned when client IP cannot be extracted
	ErrIPNotFound = errors.New("client IP address not found")
)

// ExtractClientIP gets the client IP address from trusted proxy headers
// Priority: X-Real-IP (most reliable) > X-Forwarded-For rightmost (client IP)
// Returns error if no IP can be determined
func ExtractClientIP(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata in context")
	}

	// PRIORITY 1: X-Real-IP header (set by trusted reverse proxies like nginx, cloudflare)
	// This is the most reliable source for client IP
	if xrip := md.Get("x-real-ip"); len(xrip) > 0 && xrip[0] != "" {
		return strings.TrimSpace(xrip[0]), nil
	}

	// PRIORITY 2: X-Forwarded-For header
	// Format: "client, proxy1, proxy2, ..."
	// We want the leftmost IP (original client), but validate it's not empty
	if xff := md.Get("x-forwarded-for"); len(xff) > 0 && xff[0] != "" {
		// Split by comma and get first IP (client IP)
		parts := strings.Split(xff[0], ",")
		if len(parts) > 0 {
			clientIP := strings.TrimSpace(parts[0])
			if clientIP != "" {
				return clientIP, nil
			}
		}
	}

	// No valid IP found
	return "", ErrIPNotFound
}

// ExtractClientIPOrEmpty returns the client IP or empty string if not found
// Use this when IP is optional
func ExtractClientIPOrEmpty(ctx context.Context) string {
	ip, err := ExtractClientIP(ctx)
	if err != nil {
		return ""
	}
	return ip
}

// GetUserAgentFromMetadata extracts the User-Agent header from context
func GetUserAgentFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	if userAgents := md.Get("user-agent"); len(userAgents) > 0 {
		return userAgents[0]
	}

	// Also try grpcgateway-user-agent (HTTP gateway)
	if userAgents := md.Get("grpcgateway-user-agent"); len(userAgents) > 0 {
		return userAgents[0]
	}

	return ""
}

// ExtractIdentifierForRateLimit extracts a unique identifier for rate limiting
// Priority: user_id (authenticated) > IP address (anonymous)
// This is used by rate limiting middleware
func ExtractIdentifierForRateLimit(ctx context.Context) (string, error) {
	// PRIORITY 1: Try to get user ID (for authenticated requests)
	userID, err := GetUserIDFromContext(ctx)
	if err == nil && userID != "" {
		return "user:" + userID, nil
	}

	// PRIORITY 2: Try to get IP address (for anonymous requests)
	ip, err := ExtractClientIP(ctx)
	if err == nil && ip != "" {
		return "ip:" + ip, nil
	}

	// SECURITY: Return error instead of fallback to prevent rate limit bypass
	// where all unidentified users would share the same bucket
	return "", errors.New("unable to identify client: no user_id or IP address found")
}

// GetRequestMetadata extracts common request metadata for logging/debugging
type RequestMetadata struct {
	ClientIP  string
	UserAgent string
	UserID    string
}

// ExtractRequestMetadata extracts all common request metadata at once
func ExtractRequestMetadata(ctx context.Context) RequestMetadata {
	return RequestMetadata{
		ClientIP:  ExtractClientIPOrEmpty(ctx),
		UserAgent: GetUserAgentFromMetadata(ctx),
		UserID:    GetUserIDOrEmpty(ctx),
	}
}
