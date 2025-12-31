package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/smart-kart/framework/env"
	"github.com/smart-kart/framework/logger"
	protov1 "github.com/smart-kart/proto/gen/go/proto/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Gateway wraps HTTP/REST server
type Gateway struct {
	server  *http.Server
	service interface{}
	logger  logger.Logger
	mux     *runtime.ServeMux
	ctx     context.Context
}

// ServiceRegistrar defines the interface for service registration
type ServiceRegistrar interface {
	RegisterWithHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
}

// corsMiddleware adds CORS headers to allow frontend requests
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from frontend origins
		origin := r.Header.Get("Origin")

		// Get allowed origins from environment variable or use defaults
		allowedOriginsEnv := env.GetOrDefault("CORS_ALLOWED_ORIGINS", "")
		allowedOrigins := []string{
			"http://localhost:8000",
			"http://localhost:3000",
			"http://localhost:5173", // Vite default
			"http://localhost:8083", // Admin dashboard
		}

		// Parse additional origins from environment variable (comma-separated)
		if allowedOriginsEnv != "" {
			envOrigins := strings.Split(allowedOriginsEnv, ",")
			for _, o := range envOrigins {
				trimmed := strings.TrimSpace(o)
				if trimmed != "" {
					allowedOrigins = append(allowedOrigins, trimmed)
				}
			}
		}

		// Check if origin is allowed
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-Session-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Disable caching for API responses to ensure fresh data (especially inventory)
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue with the next handler
		next.ServeHTTP(w, r)
	})
}

// customErrorHandler handles gRPC errors and removes @type from details
func customErrorHandler(_ context.Context, _ *runtime.ServeMux, _ runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Convert error to gRPC status
	st := status.Convert(err)
	pb := st.Proto()

	// Build error response without @type field
	details := make([]map[string]interface{}, 0)
	for _, detail := range pb.GetDetails() {
		var errDetail protov1.Err
		if unmarshalErr := proto.Unmarshal(detail.GetValue(), &errDetail); unmarshalErr == nil {
			details = append(details, map[string]interface{}{
				"code":    errDetail.Code,
				"message": errDetail.Message,
				"remarks": errDetail.Remarks,
			})
		}
	}

	response := map[string]interface{}{
		"code":    int(pb.GetCode()),
		"message": pb.GetMessage(),
		"details": details,
	}

	// Set HTTP status code based on gRPC code
	w.WriteHeader(runtime.HTTPStatusFromCode(st.Code()))

	// Marshal and write response
	buf, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"code":13,"message":"Internal server error","details":[]}`)
		return
	}

	_, _ = w.Write(buf)
}

// outgoingHeaderMatcher allows specific headers (like Set-Cookie) to be forwarded from gRPC to HTTP
func outgoingHeaderMatcher(key string) (string, bool) {
	switch key {
	case "set-cookie":
		return "Set-Cookie", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// incomingHeaderMatcher allows specific headers (like Cookie) to be forwarded from HTTP to gRPC
func incomingHeaderMatcher(key string) (string, bool) {
	// Convert to lowercase for case-insensitive matching
	lowerKey := strings.ToLower(key)

	switch lowerKey {
	case "cookie":
		// Forward cookie header to gRPC metadata
		// Use "grpcgateway-cookie" to be consistent with other forwarded headers
		return "grpcgateway-cookie", true
	case "authorization":
		// Also forward authorization header
		return "grpcgateway-authorization", true
	case "x-user-id":
		// Forward user ID header (set by JWT middleware after token validation)
		return "x-user-id", true
	case "x-user-role":
		// Forward user role header (set by JWT middleware for admin authentication)
		return "x-user-role", true
	case "x-session-id":
		// Forward session ID header for guest cart operations
		return "x-session-id", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// NewGateway creates a new HTTP gateway
func NewGateway(ctx context.Context) (*Gateway, error) {
	// Create gRPC-gateway runtime mux with custom error handler and metadata forwarders
	mux := runtime.NewServeMux(
		runtime.WithErrorHandler(customErrorHandler),
		runtime.WithOutgoingHeaderMatcher(outgoingHeaderMatcher),
		runtime.WithIncomingHeaderMatcher(incomingHeaderMatcher),
	)

	return &Gateway{
		logger: logger.FromContext(ctx),
		mux:    mux,
		ctx:    ctx,
	}, nil
}

// WithServiceHandler registers the service handler
func (g *Gateway) WithServiceHandler(ctx context.Context, svc interface{}) (*Gateway, error) {
	g.service = svc

	// Get gRPC port to connect to
	grpcPort := env.GetOrDefault(env.GRPCPort, "9090")
	grpcAddr := fmt.Sprintf("localhost:%s", grpcPort)

	// Create connection to gRPC server
	conn, err := grpc.NewClient(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	// Register service with gateway if it implements ServiceRegistrar
	if registrar, ok := svc.(ServiceRegistrar); ok {
		if err := registrar.RegisterWithHandler(ctx, g.mux, conn); err != nil {
			return nil, fmt.Errorf("failed to register service handlers: %w", err)
		}
	}

	// Wrap mux with CORS middleware
	corsHandler := corsMiddleware(g.mux)

	// Create HTTP server
	port := env.GetOrDefault(env.ServerPort, "8080")
	g.server = &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      corsHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return g, nil
}

// WrapHandler wraps the current handler with a custom wrapper function
// This is useful for adding custom routes or middleware that need to intercept requests
func (g *Gateway) WrapHandler(wrapper func(http.Handler) http.Handler) {
	g.server.Handler = wrapper(g.server.Handler)
}

// ListenAndServe starts the HTTP server
func (g *Gateway) ListenAndServe() error {
	g.logger.Info("HTTP server listening on %s", g.server.Addr)
	return g.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (g *Gateway) Shutdown(ctx context.Context) error {
	return g.server.Shutdown(ctx)
}