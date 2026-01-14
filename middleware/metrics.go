package middleware

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	// Request counter
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	// Request duration histogram
	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// Active requests gauge
	activeRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_active_requests",
			Help: "Number of active gRPC requests",
		},
		[]string{"method"},
	)

	// Request size histogram
	requestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_size_bytes",
			Help:    "gRPC request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method"},
	)

	// Response size histogram
	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_response_size_bytes",
			Help:    "gRPC response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method"},
	)

	// Error counter by type
	errorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_errors_total",
			Help: "Total number of gRPC errors",
		},
		[]string{"method", "error_code"},
	)
)

// MetricsInterceptor returns a gRPC interceptor that collects Prometheus metrics
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := info.FullMethod
		start := time.Now()

		// Increment active requests
		activeRequests.WithLabelValues(method).Inc()
		defer activeRequests.WithLabelValues(method).Dec()

		// Call handler
		resp, err := handler(ctx, req)

		// Record duration
		duration := time.Since(start).Seconds()
		requestDuration.WithLabelValues(method).Observe(duration)

		// Determine status
		statusCode := "OK"
		if err != nil {
			st, _ := status.FromError(err)
			statusCode = st.Code().String()
			errorsTotal.WithLabelValues(method, statusCode).Inc()
		}

		// Increment request counter
		requestsTotal.WithLabelValues(method, statusCode).Inc()

		return resp, err
	}
}

// InitMetrics initializes custom application metrics
func InitMetrics() {
	// Database connection pool metrics
	promauto.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
		func() float64 {
			// This will be updated by the application
			return 0
		},
	)

	// Cache metrics
	promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	// Authentication metrics
	promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "result"},
	)

	// Rate limit metrics
	promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of requests that exceeded rate limits",
		},
	)
}
