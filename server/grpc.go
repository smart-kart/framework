package server

import (
	"fmt"
	"net"

	"github.com/smart-kart/framework/env"
	"github.com/smart-kart/framework/logger"
	"google.golang.org/grpc"
)

// GRPCServer wraps gRPC server
type GRPCServer struct {
	server       *grpc.Server
	service      interface{}
	interceptors []grpc.UnaryServerInterceptor
	logger       logger.Logger
	registerFunc func(*grpc.Server, interface{})
}

// GRPCServiceRegistrar is a function that registers a service with a gRPC server
type GRPCServiceRegistrar func(*grpc.Server, interface{})

// NewGRPC creates a new gRPC server
func NewGRPC() *GRPCServer {
	return &GRPCServer{
		logger:       logger.New(),
		interceptors: []grpc.UnaryServerInterceptor{},
	}
}

// WithServiceInterceptors adds interceptors to the server
func (s *GRPCServer) WithServiceInterceptors(interceptors ...interface{}) *GRPCServer {
	// TODO: Convert interface{} to actual gRPC interceptors
	return s
}

// WithServiceServer registers the service implementation
func (s *GRPCServer) WithServiceServer(svc interface{}) *GRPCServer {
	s.service = svc
	return s
}

// WithGRPCRegistrar sets the function to register the service with gRPC server
func (s *GRPCServer) WithGRPCRegistrar(fn GRPCServiceRegistrar) *GRPCServer {
	s.registerFunc = fn
	return s
}

// ListenAndServe starts the gRPC server
func (s *GRPCServer) ListenAndServe() error {
	port := env.GetOrDefault(env.GRPCPort, "50051")
	addr := fmt.Sprintf(":%s", port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Create gRPC server with interceptors
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(s.interceptors...),
	}
	s.server = grpc.NewServer(opts...)

	// Register service with gRPC server using the registrar function
	if s.registerFunc != nil && s.service != nil {
		s.registerFunc(s.server, s.service)
	}

	s.logger.Info("gRPC server listening on %s", addr)
	return s.server.Serve(listener)
}

// Shutdown gracefully shuts down the server
func (s *GRPCServer) Shutdown() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}