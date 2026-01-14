package middleware

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// CorrelationIDKey is the context key for correlation ID
type contextKey string

const CorrelationIDKey contextKey = "correlation_id"

// CorrelationIDInterceptor returns a gRPC interceptor that adds correlation IDs to requests
// This allows tracing requests across microservices
func CorrelationIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract correlation ID from incoming metadata
		var correlationID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get("x-correlation-id"); len(ids) > 0 {
				correlationID = ids[0]
			}
		}

		// Generate new correlation ID if not present
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Add correlation ID to context
		ctx = context.WithValue(ctx, CorrelationIDKey, correlationID)

		// Add correlation ID to outgoing metadata
		md := metadata.Pairs("x-correlation-id", correlationID)
		ctx = metadata.NewOutgoingContext(ctx, md)

		// Continue with request
		return handler(ctx, req)
	}
}

// GetCorrelationID extracts correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}
