package middleware

import (
	"context"
	"strings"

	"github.com/cozy-hub-app/framework/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor creates a gRPC unary interceptor that extracts user ID from JWT
// and adds it to the request context metadata.
// This allows authenticated endpoints to access the user ID via metadata.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract authorization header from metadata
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			// No metadata, continue without auth
			return handler(ctx, req)
		}

		// Get authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			// Also try grpcgateway-authorization (from HTTP gateway)
			authHeaders = md.Get("grpcgateway-authorization")
		}

		if len(authHeaders) == 0 {
			// No auth header, continue without setting user_id
			return handler(ctx, req)
		}

		// Parse Bearer token
		authHeader := authHeaders[0]
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid format, continue without auth
			return handler(ctx, req)
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			return handler(ctx, req)
		}

		// Validate token and extract user ID
		jwtManager := jwt.GetJWTManager()
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without setting user_id
			// The endpoint handler will return unauthorized if user_id is required
			return handler(ctx, req)
		}

		// Add user_id to metadata
		if claims.UserID != "" {
			md = metadata.Join(md, metadata.Pairs("user_id", claims.UserID))
			ctx = metadata.NewIncomingContext(ctx, md)
		}

		return handler(ctx, req)
	}
}
