package middleware

import (
	"context"
	"time"

	"github.com/smart-kart/framework/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor returns a gRPC interceptor that logs request/response details
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Get logger from context
		log := logger.FromContext(ctx)

		// Extract correlation ID if present
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ids := md.Get("x-correlation-id"); len(ids) > 0 {
				ctx = logger.WithContext(ctx, log)
			}
		}

		// Extract user ID from context if present
		userID := ""
		if uid, ok := ctx.Value("user_id").(string); ok {
			userID = uid
		}

		// Log request start
		log.Info("gRPC request started",
			"method", info.FullMethod,
			"user_id", userID,
		)

		// Call handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Extract status
		st, _ := status.FromError(err)
		statusCode := st.Code()

		// Log request completion
		if err != nil {
			log.Error("gRPC request failed",
				"method", info.FullMethod,
				"user_id", userID,
				"duration_ms", duration.Milliseconds(),
				"status_code", statusCode.String(),
				"error", err.Error(),
			)
		} else {
			log.Info("gRPC request completed",
				"method", info.FullMethod,
				"user_id", userID,
				"duration_ms", duration.Milliseconds(),
				"status_code", "OK",
			)
		}

		return resp, err
	}
}

// HTTPLoggingMiddleware logs HTTP gateway requests
func HTTPLoggingMiddleware(next grpc.UnaryHandler) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		start := time.Now()
		log := logger.FromContext(ctx)

		// Extract HTTP metadata
		var path, method string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if paths := md.Get(":path"); len(paths) > 0 {
				path = paths[0]
			}
			if methods := md.Get(":method"); len(methods) > 0 {
				method = methods[0]
			}
		}

		// Call handler
		resp, err := next(ctx, req)

		duration := time.Since(start)

		// Determine status code
		statusCode := codes.OK
		if err != nil {
			st, _ := status.FromError(err)
			statusCode = st.Code()
		}

		// Log
		if err != nil {
			log.Error("HTTP request failed",
				"method", method,
				"path", path,
				"duration_ms", duration.Milliseconds(),
				"status", statusCode.String(),
				"error", err.Error(),
			)
		} else {
			log.Info("HTTP request completed",
				"method", method,
				"path", path,
				"duration_ms", duration.Milliseconds(),
				"status", "OK",
			)
		}

		return resp, err
	}
}
