package middleware

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

var (
	// ErrUserIDNotFound is returned when user_id is not found in context
	ErrUserIDNotFound = errors.New("user_id not found in context")
)

// GetUserIDFromContext tries to extract user_id from both metadata and context value
// This is the preferred method for extracting user_id in handlers
// Priority: metadata (set by auth middleware) > context value (legacy)
func GetUserIDFromContext(ctx context.Context) (string, error) {
	// PRIORITY 1: Try metadata first (set by auth middleware)
	userID, err := GetUserIDFromMetadata(ctx)
	if err == nil && userID != "" {
		return userID, nil
	}

	// PRIORITY 2: Try context value (legacy support)
	if uid, ok := ctx.Value("user_id").(string); ok && uid != "" {
		return uid, nil
	}

	return "", ErrUserIDNotFound
}

// GetUserIDFromMetadata extracts user_id from gRPC metadata
// This is set by the auth middleware after JWT validation
func GetUserIDFromMetadata(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("no metadata in context")
	}

	userIDs := md.Get("user_id")
	if len(userIDs) == 0 || userIDs[0] == "" {
		return "", ErrUserIDNotFound
	}

	return userIDs[0], nil
}

// SetUserIDInContext adds user_id to both context value and metadata
// This is useful for propagating user_id to downstream services
func SetUserIDInContext(ctx context.Context, userID string) context.Context {
	// Add to context value
	ctx = context.WithValue(ctx, "user_id", userID)

	// Add to metadata for gRPC propagation
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	md = metadata.Join(md, metadata.Pairs("user_id", userID))
	ctx = metadata.NewIncomingContext(ctx, md)

	return ctx
}

// GetUserIDOrEmpty returns user_id or empty string if not found
// Use this when user_id is optional (e.g., endpoints that work for both authenticated and anonymous users)
func GetUserIDOrEmpty(ctx context.Context) string {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return ""
	}
	return userID
}

// RequireUserID extracts user_id from context and returns an error if not found
// Use this for endpoints that require authentication
func RequireUserID(ctx context.Context) (string, error) {
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		return "", errors.New("authentication required: user_id not found in context")
	}
	if userID == "" {
		return "", errors.New("authentication required: user_id is empty")
	}
	return userID, nil
}
