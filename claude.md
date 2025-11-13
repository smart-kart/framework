# Smart Kart Framework - Complete Documentation

## Table of Contents
1. [Framework Overview](#framework-overview)
2. [Directory Structure](#directory-structure)
3. [Core Modules](#core-modules)
4. [Shared Components](#shared-components)
5. [Configuration Management](#configuration-management)
6. [Server Management](#server-management)
7. [Response Handling](#response-handling)
8. [Validation System](#validation-system)
9. [Database Integration](#database-integration)
10. [Build System and Make Targets](#build-system-and-make-targets)
11. [Dependencies and External Packages](#dependencies-and-external-packages)
12. [Service Integration Guidelines](#service-integration-guidelines)
13. [Development Guidelines](#development-guidelines)

---

## Framework Overview

### Purpose
The Smart Kart Framework is a shared library/module that provides common utilities, configurations, and base functionality to all microservices in the Smart Kart ecosystem (Account service, Products service, etc.).

### Key Responsibilities
- Provide standardized gRPC and HTTP server implementations
- Manage database connections and lifecycle
- Handle error responses with custom error codes
- Validate request data with standardized validation
- Configure environment variables
- Provide structured logging
- Support graceful profiling for performance analysis
- Enable inter-service communication patterns

### Module Path
```
github.com/smart-kart/framework
```

---

## Directory Structure

```
framework/
├── application/           # Application initialization and lifecycle
│   └── application.go     # Main application struct and builder pattern
├── env/                   # Environment configuration management
│   └── env.go            # Environment variable loading and validation
├── logger/               # Logging utilities
│   └── logger.go         # Structured logging interface and implementation
├── pgx/                  # PostgreSQL database driver
│   └── driver.go         # PostgreSQL connection pool management
├── response/             # HTTP/gRPC response and error handling
│   ├── constant.go       # Response constants and HTTP status mappings
│   ├── error_code_msg.go # Error code to message mappings
│   └── grpc.go          # gRPC response builders (Success, Error, etc.)
├── server/              # Server implementations
│   ├── http.go          # HTTP/REST gateway using gRPC-gateway
│   ├── grpc.go          # gRPC server wrapper
│   └── profiler.go      # Performance profiler server
├── utils/               # Utility functions
│   └── generic/         # Generic type utilities
│       ├── generic.go   # Generic helper functions
│       └── generic_test.go
├── validate/            # Request validation
│   ├── validate.go      # Validator interface and request validation
│   └── constant.go      # Validation constants (country codes, etc.)
├── make/                # Build system
│   └── proto.mk         # Makefile targets for proto generation
├── Makefile             # Root Makefile
├── go.mod               # Go module dependencies
└── go.sum               # Locked dependency versions
```

---

## Core Modules

### 1. application/application.go

**Purpose**: Manages application lifecycle and dependency initialization using the builder pattern.

**Key Struct**: `Application`
```go
type Application struct {
    logger         logger.Logger
    pgxRegistrar   func(context.Context) error
    redisConfig    *RedisConfig
    awsConfig      *AWSConfig
    errorHandler   ErrorHandler
    customValidator interface{}
}
```

**Key Methods**:
- `New()` - Creates a new Application instance
- `WithPgx(registrar)` - Registers PostgreSQL initialization function
- `WithRedis()` - Configures Redis connection (Host, Port, Password from env)
- `WithAWS()` - Configures AWS services (SQS, S3) from env
- `WithErrorCode(errMsg, validationErr)` - Sets error and validation error mappings
- `WithCustomValidator(validator)` - Adds custom validation logic
- `Run(ctx)` - Initializes all configured dependencies in order

**Configuration Structs**:
- `RedisConfig` - Redis connection settings
- `AWSConfig` - AWS service configuration
- `ErrorHandler` - Error message and validation error mappings

**Usage Pattern**:
```go
app := application.New().
    WithPgx(pgx.Init).
    WithRedis().
    WithAWS().
    WithErrorCode(errorMessages, validationErrors)

if err := app.Run(ctx); err != nil {
    // handle initialization error
}
```

---

### 2. env/env.go

**Purpose**: Centralized environment variable management with support for .env files and YAML configs.

**Key Constants** (Environment Variable Keys):
```
Service, Environment, ServerPort, GRPCPort
DBHost, DBPort, DBUser, DBPassword, DBName, DBDrivers
RedisHost, RedisPort, RedisPassword
AWSRegion, SQSQueueURL, S3Bucket
JWTSecretKey, JWTAccessTokenTTL, JWTRefreshTokenTTL, JWTIssuer
```

**Environment Types**:
- `UnitTest` - "unittest"
- `Dev` - "dev"
- `Staging` - "staging"
- `Prod` - "prod"

**Key Functions**:
- `Get(key string) string` - Get environment variable (empty if not set)
- `GetOrDefault(key, defaultValue string) string` - Get or return default value
- `Set(key, value string) error` - Set environment variable
- `GetList(key string) []string` - Get comma-separated values as slice
- `LoadFromEnv(filePath string) error` - Load from .env file (KEY=VALUE format)
- `LoadFromYAML(filePath string) error` - Load from YAML configuration file
- `Validate(requiredVars ...string) error` - Verify required variables are set

**Features**:
- Automatically parses .env files with support for comments (# prefix)
- Skips quotes in values ("value" becomes value)
- Never overrides already-set environment variables
- Returns nil error if .env file doesn't exist (optional)

---

### 3. logger/logger.go

**Purpose**: Standardized logging interface for all services.

**Logger Interface**:
```go
type Logger interface {
    Info(msg string, args ...interface{})
    Error(msg string, args ...interface{})
    Fatal(msg string, args ...interface{})
    Debug(msg string, args ...interface{})
}
```

**Default Implementation**:
- Uses Go's standard `log` package
- INFO logs to stdout with file:line prefix
- ERROR logs to stderr with file:line prefix
- DEBUG logs to stdout with file:line prefix
- FATAL logs to stderr then exits the process

**Key Functions**:
- `New() Logger` - Create a new logger instance
- `WithContext(ctx, logger) context.Context` - Add logger to context
- `FromContext(ctx) Logger` - Extract logger from context (returns new logger if not found)
- `RestrictedGet() Logger` - Get logger for framework internal usage

**Usage Pattern**:
```go
logger := logger.New()
logger.Info("Application started")
logger.Error("Database connection failed: %v", err)

// Context-based usage
ctx = logger.WithContext(ctx, logger)
log := logger.FromContext(ctx)
```

---

### 4. pgx/driver.go

**Purpose**: PostgreSQL connection pool management using pgx driver.

**Key Struct**: `DataSource`
```go
type DataSource struct {
    pool *pgxpool.Pool
}
```

**Global State**:
- `_ds *DataSource` - Global datasource singleton

**Key Functions**:
- `Init(ctx) error` - Initialize PostgreSQL connection pool
  - Reads from: DBHost, DBPort, DBUser, DBPassword, DBName environment variables
  - Creates connection string: `host=X port=Y user=Z password=A dbname=B sslmode=disable`
  - Tests connection with `Ping()`
  - Stores in global `_ds` variable

- `GetDS() *DataSource` - Get the global datasource instance
- `(ds *DataSource) GetPool() *pgxpool.Pool` - Access the connection pool
- `(ds *DataSource) Close()` - Close connection pool gracefully

**Required Environment Variables**:
- `DB_HOST` - PostgreSQL server hostname
- `DB_PORT` - PostgreSQL server port
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name

**Usage Pattern**:
```go
if err := pgx.Init(ctx); err != nil {
    log.Fatal("Failed to initialize database")
}

ds := pgx.GetDS()
pool := ds.GetPool()
// Use pool for queries
```

---

### 5. server/

**Purpose**: HTTP and gRPC server implementations with gateway support.

#### 5.1 server/http.go - HTTP Gateway

**Key Struct**: `Gateway`
```go
type Gateway struct {
    server  *http.Server
    service interface{}
    logger  logger.Logger
    mux     *runtime.ServeMux
    ctx     context.Context
}
```

**Key Functions**:
- `NewGateway(ctx) (*Gateway, error)` - Create HTTP gateway with gRPC-gateway runtime
  - Sets up custom error handler
  - Configures header matchers for cookies
  - Returns gateway ready for service registration

- `(g *Gateway) WithServiceHandler(ctx, svc) (*Gateway, error)` - Register gRPC service
  - Expects gRPC server on localhost:GRPC_PORT (default 9090)
  - Connects to gRPC server with insecure credentials
  - Calls `svc.RegisterWithHandler()` if service implements `ServiceRegistrar` interface
  - Wraps with CORS middleware
  - Creates HTTP server on SERVER_PORT (default 8080)
  - Sets timeouts: Read=15s, Write=15s, Idle=60s

- `(g *Gateway) ListenAndServe() error` - Start HTTP server (blocks until stopped)
- `(g *Gateway) Shutdown(ctx) error` - Gracefully shutdown server
- `(g *Gateway) WrapHandler(wrapper func(http.Handler) http.Handler)` - Add middleware

**Features**:
- CORS support for: localhost:8000, localhost:3000, localhost:5173 (Vite), localhost:8083 (Admin)
- Custom error handler that removes `@type` from gRPC error details
- Cookie and header forwarding between HTTP and gRPC
- HTTP status codes properly mapped from gRPC codes

**ServiceRegistrar Interface**:
```go
type ServiceRegistrar interface {
    RegisterWithHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
}
```

---

#### 5.2 server/grpc.go - gRPC Server

**Key Struct**: `GRPCServer`
```go
type GRPCServer struct {
    server       *grpc.Server
    service      interface{}
    interceptors []grpc.UnaryServerInterceptor
    logger       logger.Logger
    registerFunc func(*grpc.Server, interface{})
}
```

**Key Functions**:
- `NewGRPC() *GRPCServer` - Create new gRPC server with logger
- `(s *GRPCServer) WithServiceInterceptors(...interface{}) *GRPCServer` - Add interceptors (TODO)
- `(s *GRPCServer) WithServiceServer(svc) *GRPCServer` - Set service implementation
- `(s *GRPCServer) WithGRPCRegistrar(fn) *GRPCServer` - Set service registration function
- `(s *GRPCServer) ListenAndServe() error` - Start server
  - Listens on GRPC_PORT (default 50051)
  - Registers service using `registerFunc`
  - Blocks until stopped or error

- `(s *GRPCServer) Shutdown()` - Graceful shutdown (calls GracefulStop)

**GRPCServiceRegistrar Type**:
```go
type GRPCServiceRegistrar func(*grpc.Server, interface{})
```

---

#### 5.3 server/profiler.go - Performance Profiler

**Key Function**:
- `RunProfiler()` - Start profiler server on PROFILER_PORT (default 6060)
  - Imports `net/http/pprof` to enable Go profiling endpoints
  - Available endpoints: /debug/pprof/*, /debug/pprof/profile, /debug/pprof/trace, etc.
  - Can be run as goroutine in background

**Usage**:
```go
go server.RunProfiler()  // Enable profiler in background

// Later, access profiling data:
// http://localhost:6060/debug/pprof/
// http://localhost:6060/debug/pprof/heap
// http://localhost:6060/debug/pprof/goroutine
```

---

### 6. response/

**Purpose**: Standardized response and error handling for gRPC services.

#### 6.1 response/constant.go

**HTTP Status Constants**:
```
statusSuccess, statusCreated, statusAccepted, statusError
```

**HTTP Header Keys**:
```
headerTrailer, headerTransferEncoding, headerContentType
headerWWWAuthenticate, headerCacheControl
```

**HTTP Header Values**:
```
ctApplicationJSON = "application/json"
ccNoStore = "no-store, max-age=0"
```

**gRPC Metadata**:
```
mdHTTPStatusCode = "x-http-statuscode"
```

**Error Code Series** (4-digit codes starting at 100):
```go
ErrInvalidRequest        = 100
ErrDBOperationFailed     = 101
ErrSomethingWentWrong    = 102
ErrInvalidPathParam      = 103
ErrInvalidQueryParam     = 104
ErrResourceNotFound      = 105
ErrInvalidToken          = 106
ErrTokenExpired          = 107
ErrInvalidBasicAuth      = 108
ErrEmptyBasicAuth        = 109
ErrTooManyRequests       = 110
ErrInvalidAPIKey         = 111
ErrUnsupportedFileType   = 112
```

---

#### 6.2 response/error_code_msg.go

**Global Error Messages**:
```go
var ErrMsg = map[ErrCode]string{
    ErrInvalidRequest: "Failure on decoding the json request...",
    ErrDBOperationFailed: "Something went wrong with this API",
    // ... more mappings
}
```

---

#### 6.3 response/grpc.go - Response Builders

**Type Definitions**:
```go
type ErrCode int32        // Error code
type Remarks string       // Additional error remarks
type ErrType string       // Validation error type
```

**Error Code Management**:
- `LoadErrCode(em map[ErrCode]string, vec map[ErrType]map[string]ErrCode)` - Initialize error codes and validation errors
- `RegisterErrMsg(em map[ErrCode]string)` - Register additional error messages
- `RegisterFieldErrCode(vec map[ErrType]map[string]ErrCode)` - Register validation field error codes
- `GetValidationErrCode(errType, jsonTag) ErrCode` - Get error code for field
- `GetErrMsg(errCode) string` - Get message for error code

**Success Response Functions** (return error = nil):
- `Success[T any](ctx, res) (T, error)` - HTTP 200
- `Created[T any](ctx, res) (T, error)` - HTTP 201
- `Accepted[T any](ctx, res) (T, error)` - HTTP 202

**Error Response Functions**:
All error functions follow the pattern: `Func[T any](ctx, res T, args ...any) (T, error)`

They return the zero value of type T and a gRPC status error.

**Status Code Mappings**:
```
Canceled            -> codes.Canceled (499)
Unknown             -> codes.Internal (500)
InvalidArgument     -> codes.InvalidArgument (400)
DeadlineExceeded    -> codes.DeadlineExceeded (504)
NotFound            -> codes.NotFound (404)
AlreadyExists       -> codes.AlreadyExists (409)
PermissionDenied    -> codes.PermissionDenied (403)
ResourceExhausted   -> codes.ResourceExhausted (429)
FailedPrecondition  -> codes.FailedPrecondition (400)
Aborted             -> codes.Aborted (409)
OutOfRange          -> codes.OutOfRange (400)
Unimplemented       -> codes.Unimplemented (501)
InternalError       -> codes.Internal (500)
Unavailable         -> codes.Unavailable (503)
DataLoss            -> codes.DataLoss (500)
Unauthenticated     -> codes.Unauthenticated (401)
```

**Error Detail Functions**:
- `FormatErr(errCode, args...)` - Format error message with args
- `FormatErrWithRemarks(errCode, remarks, args...)` - Format with remarks
- `StrictError[T](ctx, res, code, msg, strictErr)` - Use strict error object
- `GRPCError[T](ctx, res, err)` - Pass through existing gRPC errors
- `ReadGRPCError(err) *GRPCError` - Parse gRPC error details

**Remarks Type**:
```go
type Remarks string  // Additional context for errors
```

**Usage Pattern**:
```go
// Success response
return response.Success(ctx, &User{ID: 123})

// Error with code
return response.InvalidArgument(ctx, nil, response.ErrInvalidRequest)

// Error with details
return response.NotFound(ctx, nil, response.ErrResourceNotFound,
    response.Remarks("user_id not found in database"))

// Custom error code
return response.InternalError(ctx, nil, ErrCode(5001))
```

---

### 7. validate/

**Purpose**: Standardized request validation using go-playground/validator.

#### 7.1 validate/validate.go

**Global Validator**: `_v *validator.Validate` (thread-safe singleton)

**Key Functions**:
- `Request(ctx, req, errType) error` - Validate struct fields
  - Uses JSON tags for field names
  - Returns nil if validation passes
  - Returns gRPC InvalidArgument error with field details on validation failure
  - Includes detailed error codes and remarks showing field path (root.element.key)

- `RegisterCustomValidators(map[string]validator.Func) error` - Register custom validation functions
  - `validator.Func` has signature: `func(fl validator.FieldLevel) bool`

- `GetValidator() *validator.Validate` - Access validator for custom usage

**Error Type**:
```go
type ErrType string  // e.g., "CreateUserRequest"
```

**Validation Flow**:
1. Service defines validation error mapping: `ErrType -> Field -> ErrCode`
2. Framework calls `request.Validate(..., errType)`
3. If validation fails, framework:
   - Looks up custom error code for field
   - Gets error message for code
   - Returns field path as remarks
   - Returns gRPC InvalidArgument status

#### 7.2 validate/constant.go

**Country Constants**:
```go
CountryUS = "US"
CountryCA = "CA"
CountryIN = "IN"
CountryBR = "BR"
```

**Size Constants**:
```go
int2 = 2, int4 = 4, int24 = 24, int25 = 25
```

---

### 8. utils/generic/generic.go

**Purpose**: Generic utility functions using Go generics.

**Functions**:
- `ReturnZero[T any](T) T` - Return zero value of type T
- `Contains[T comparable]([]T, T) bool` - Check if value in slice
- `Remove[T comparable]([]T, T) []T` - Remove all occurrences of value
- `RemoveDuplicates[T comparable]([]T) []T` - Remove duplicate elements
- `IsZero[T any](T) bool` - Check if value is zero value for its type

---

## Shared Components

### Components All Services Depend On

1. **Logger** - Standard logging interface used across all services
   - Services get logger from Application or context
   - Consistent log formatting and output

2. **Environment Configuration** - Centralized env management
   - All services use same env variable keys
   - Support for .env files and YAML configs
   - Same environment type constants (dev, staging, prod)

3. **Response Handling** - Standardized error responses
   - All services use same error codes and mappings
   - Consistent HTTP status code handling
   - Same gRPC error detail format

4. **Server Management** - gRPC and HTTP servers
   - gRPC server with interceptor support
   - HTTP gateway with gRPC-gateway for REST endpoints
   - Consistent port configuration

5. **Validation** - Request validation framework
   - Same validator implementation
   - Custom error code registration per service
   - Consistent error detail format

6. **Database** - PostgreSQL driver
   - Reusable connection pool management
   - Single global datasource per service instance

7. **Build System** - Makefile targets
   - Proto generation and compilation
   - Testing, coverage, linting
   - Dependency management

---

## Configuration Management

### Environment Variables Used by Framework

**Service Configuration**:
```
SERVICE_NAME         - Service identifier
ENVIRONMENT          - dev, staging, prod, unittest
SERVER_PORT          - HTTP server port (default: 8080)
GRPC_PORT           - gRPC server port (default: 50051)
PROFILER_PORT       - Profiler port (default: 6060)
```

**Database Configuration**:
```
DB_HOST             - PostgreSQL hostname
DB_PORT             - PostgreSQL port
DB_USER             - Database user
DB_PASSWORD         - Database password
DB_NAME             - Database name
DB_DRIVERS          - Driver types
```

**Redis Configuration** (Optional):
```
REDIS_HOST          - Redis hostname (default: localhost)
REDIS_PORT          - Redis port (default: 6379)
REDIS_PASSWORD      - Redis password
```

**AWS Configuration** (Optional):
```
AWS_REGION          - AWS region (default: us-east-1)
SQS_QUEUE_URL       - SQS queue URL
S3_BUCKET           - S3 bucket name
```

**JWT Configuration** (For auth services):
```
JWT_SECRET_KEY      - Secret key for signing
JWT_ACCESS_TOKEN_TTL - Access token expiration
JWT_REFRESH_TOKEN_TTL - Refresh token expiration
JWT_ISSUER          - Token issuer claim
```

### Configuration Loading Priority

1. Environment variables (OS-level)
2. .env file (KEY=VALUE format)
3. YAML configuration file
4. Default values in code

Example .env file:
```
SERVICE_NAME=account
ENVIRONMENT=dev
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=secret
DB_NAME=smart_kart
```

---

## Server Management

### Server Architecture

```
Client
  ↓
[HTTP Server on :8080]
  ↓
[gRPC-Gateway]
  ↓
[gRPC Server on :50051]
  ↓
[Service Implementation]
```

### HTTP/gRPC Gateway Flow

1. HTTP request comes to Gateway on :8080
2. Gateway converts HTTP request to gRPC format
3. Gateway connects to gRPC server on :50051
4. Service processes gRPC request
5. Response converted back to HTTP JSON
6. HTTP response sent to client

### CORS Support

The HTTP gateway allows requests from:
- `http://localhost:8000` - Main application
- `http://localhost:3000` - Frontend dev server
- `http://localhost:5173` - Vite dev server
- `http://localhost:8083` - Admin dashboard

Headers allowed: `Content-Type, Authorization, X-Requested-With`
Methods allowed: `GET, POST, PUT, PATCH, DELETE, OPTIONS`
Cookies supported: Yes (Set-Cookie header forwarding)

### Graceful Shutdown Pattern

```go
// Start servers in goroutines
go func() {
    if err := grpcServer.ListenAndServe(); err != nil {
        log.Error("gRPC server error: %v", err)
    }
}()

go func() {
    if err := httpGateway.ListenAndServe(); err != nil {
        log.Error("HTTP server error: %v", err)
    }
}()

// Wait for shutdown signal
<-sigChan

// Graceful shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

grpcServer.Shutdown()
httpGateway.Shutdown(ctx)
```

---

## Response Handling

### Error Response Structure

All errors follow this structure:
```json
{
  "code": 13,
  "message": "Internal server error",
  "details": [
    {
      "code": 1001,
      "message": "Custom error message",
      "remarks": "Additional context"
    }
  ]
}
```

### Error Code Ranges

- **100-199**: Framework/General errors
- **1000-9999**: Service-specific custom errors

### Error Handling Flow

1. Service encounters error condition
2. Service calls appropriate response function (e.g., `response.InvalidArgument()`)
3. Function builds gRPC status with error details
4. gRPC returns status to HTTP gateway
5. Gateway converts gRPC status to HTTP JSON
6. Client receives standardized error response

### Custom Error Registration

Services must register custom error codes:
```go
var customErrors = map[response.ErrCode]string{
    5001: "User not found",
    5002: "Email already exists",
}

response.LoadErrCode(customErrors, nil)
```

---

## Validation System

### Validation Workflow

1. Service receives request
2. Service calls `validate.Request(ctx, req, errType)`
3. Validator checks struct tags (required, email, min, max, etc.)
4. If valid, returns nil
5. If invalid, returns detailed gRPC error with field names and custom codes

### Request Struct Example

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"gte=0,lte=120"`
}

// Service registers error codes
var validationErrors = map[response.ErrType]map[string]response.ErrCode{
    "CreateUserRequest": {
        "name": 5001,
        "email": 5002,
        "age": 5003,
    },
}

response.LoadErrCode(nil, validationErrors)

// In handler
err := validate.Request(ctx, req, "CreateUserRequest")
if err != nil {
    return nil, err  // Returns detailed validation error
}
```

### Custom Validators

Register custom validation functions:
```go
validate.RegisterCustomValidators(map[string]validator.Func{
    "phone": func(fl validator.FieldLevel) bool {
        phone := fl.Field().String()
        // Custom phone validation
        return isValidPhone(phone)
    },
})
```

---

## Database Integration

### PostgreSQL Connection Pool

The framework manages a global PostgreSQL connection pool using pgx.

**Initialization**:
```go
if err := pgx.Init(ctx); err != nil {
    log.Fatal("DB init failed")
}
```

**Usage**:
```go
ds := pgx.GetDS()
pool := ds.GetPool()

row := pool.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", userID)
var user User
if err := row.Scan(&user.ID, &user.Name); err != nil {
    // handle error
}
```

**Connection String Format**:
```
host={DB_HOST} port={DB_PORT} user={DB_USER} password={DB_PASSWORD} dbname={DB_NAME} sslmode=disable
```

**Features**:
- Connection pooling for efficiency
- Automatic connection testing on initialization
- Graceful pool closing

---

## Build System and Make Targets

### Root Makefile
Location: `/Users/mrpsycho/smart-kart/framework/Makefile`

Simply includes the common Makefile:
```makefile
include make/proto.mk
```

### Proto Makefile
Location: `/Users/mrpsycho/smart-kart/framework/make/proto.mk`

This file provides all build, test, and deployment targets.

#### Proto Generation Targets

- `make proto-gen` - Full generation pipeline
  - Cleans old generated files
  - Runs buf to generate from .proto files
  - Adds Go struct tags to generated code
  - Generates mocks for interfaces
  - Tidies go.mod

- `make proto-gen-clean` - Remove all generated files
- `make buf` - Run buf code generation
- `make go-tags` - Inject Go struct tags
- `make mock` - Generate mocks using mockery

#### Testing Targets

- `make unit-test` - Run all unit tests with coverage
  - Sets `-mod=readonly` for deterministic builds
  - Generates coverage.txt file
  - Outputs coverage percentage

- `make cover` - Generate HTML coverage report
  - Converts coverage.txt to coverage.html
  - Human-readable coverage visualization

- `make cover-hint` - Prioritize test coverage
  - Shows functions ordered by impact
  - Recommends which sections to test first

#### Code Quality Targets

- `make lint` - Run golangci-lint
  - Auto-installs if missing
  - Runs with 5-minute timeout
  - Formats code with gofmt

- `make mod` - Tidy and verify dependencies
  - `go mod tidy` - Remove unused, add required
  - `go mod verify` - Check integrity

#### Maintenance Targets

- `make update-dep` - Update all dependencies
  - Runs `go get -u all`
  - Updates test dependencies
  - Cleans and tidies

- `make clean` - Clean all caches
  - Golangci-lint cache
  - Go build cache
  - Go test cache
  - Clears go.sum

#### Documentation

- `make help` - Display all available targets

### Configuration Variables

Customize in service Makefile:
```makefile
include ../../framework/make/proto.mk

SERVICE_NAME := account
PROTO_DIR := ../proto           # Proto file location
PROTO_GEN_DIR := proto/gen      # Generated code output
COVER_MODE := atomic            # Coverage mode
COVER_PROFILE := coverage.txt   # Coverage file
GO_TEST_PKG := ./...           # Test packages
GO_COVER_PKG := ./...          # Coverage packages
```

### Tool Requirements

The Makefile auto-installs, but can install manually:
```bash
# Proto generation
go install github.com/bufbuild/buf/cmd/buf@latest

# Mock generation
go install github.com/vektra/mockery/v2@latest

# Linting
brew install golangci-lint  # macOS
# or
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh

# Coverage
go install github.com/gojekfarm/go-coverage@latest
go install github.com/axw/gocov/gocov@latest
go install github.com/matm/gocov-html/cmd/gocov-html@latest
```

---

## Dependencies and External Packages

### Direct Dependencies

```
github.com/go-playground/validator/v10 v10.24.0
  - Request validation with struct tags
  - Custom validators support
  - 40+ built-in validation rules

github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2
  - HTTP/gRPC reverse proxy
  - Converts HTTP to gRPC calls
  - Maps HTTP status to gRPC codes

github.com/jackc/pgx/v5 v5.7.6
  - PostgreSQL driver
  - Connection pooling
  - Prepared statement support

github.com/smart-kart/proto v0.0.0
  - Local proto package (replaced with ../proto)
  - Contains .proto definitions
  - Generated code: *pb.go files

github.com/spf13/cast v1.10.0
  - Type casting utilities
  - Safe conversions between types

github.com/stretchr/testify v1.8.4
  - Testing assertions
  - Mocking utilities
  - Test fixtures

google.golang.org/grpc v1.75.1
  - gRPC framework
  - Interceptor support
  - Service definitions

google.golang.org/protobuf v1.36.9
  - Protocol Buffers v3 support
  - Code generation
  - Message serialization

gopkg.in/yaml.v3 v3.0.1
  - YAML parsing and marshaling
  - Config file support
```

### Indirect Dependencies

All indirect dependencies handled by Go modules, including:
- Cryptography (golang.org/x/crypto)
- Networking (golang.org/x/net)
- Text processing (golang.org/x/text)
- Synchronization (golang.org/x/sync)
- Proto-generated code (google.golang.org/genproto/*)

### Go Version

```
go 1.24.0
```

### Dependency Management

```bash
# View dependencies
go list -m all

# Update framework only
go get github.com/smart-kart/framework

# Update all dependencies
make update-dep

# Verify dependency integrity
go mod verify

# Tidy unused dependencies
make mod
```

---

## Service Integration Guidelines

### Step 1: Import Framework

```go
import (
    "github.com/smart-kart/framework/env"
    "github.com/smart-kart/framework/logger"
    "github.com/smart-kart/framework/pgx"
    "github.com/smart-kart/framework/response"
    "github.com/smart-kart/framework/server"
    "github.com/smart-kart/framework/validate"
)
```

### Step 2: Setup Build System

Create `Makefile` in service root:
```makefile
include ../framework/make/proto.mk

SERVICE_NAME := account

run:
	go run cmd/main.go

build:
	go build -o bin/$(SERVICE_NAME) cmd/main.go

.DEFAULT_GOAL := help
```

### Step 3: Initialize Environment

```go
func main() {
    // Load environment variables
    if err := env.LoadFromEnv(".env"); err != nil {
        log.Fatal(err)
    }

    // Validate required vars
    if err := env.Validate(
        env.DBHost,
        env.DBPort,
        env.DBUser,
        env.DBPassword,
        env.DBName,
    ); err != nil {
        log.Fatal(err)
    }
}
```

### Step 4: Setup Logger and Application

```go
func main() {
    log := logger.New()
    ctx := logger.WithContext(context.Background(), log)

    app := application.New().
        WithPgx(pgx.Init).
        WithErrorCode(errorMessages, validationErrors).
        WithCustomValidator(customValidators)

    if err := app.Run(ctx); err != nil {
        log.Fatal("Application initialization failed: %v", err)
    }
}
```

### Step 5: Create gRPC Server

```go
grpcServer := server.NewGRPC().
    WithServiceServer(service).
    WithGRPCRegistrar(proto.RegisterAccountServiceServer)

go func() {
    if err := grpcServer.ListenAndServe(); err != nil {
        log.Error("gRPC server error: %v", err)
    }
}()
```

### Step 6: Create HTTP Gateway

```go
gateway, err := server.NewGateway(ctx)
if err != nil {
    log.Fatal("Gateway creation failed: %v", err)
}

gateway, err = gateway.WithServiceHandler(ctx, service)
if err != nil {
    log.Fatal("Handler registration failed: %v", err)
}

go func() {
    if err := gateway.ListenAndServe(); err != nil {
        log.Error("HTTP server error: %v", err)
    }
}()
```

### Step 7: Register Error Codes

```go
var errorMessages = map[response.ErrCode]string{
    5001: "User not found",
    5002: "Email already exists",
    5003: "Invalid password",
}

var validationErrors = map[response.ErrType]map[string]response.ErrCode{
    "CreateUserRequest": {
        "name": 5001,
        "email": 5002,
        "password": 5003,
    },
}
```

### Step 8: Implement Service Handler

```go
type AccountService struct {
    proto.UnimplementedAccountServiceServer
}

func (s *AccountService) CreateAccount(ctx context.Context, req *proto.CreateAccountRequest) (*proto.Account, error) {
    // Validate request
    if err := validate.Request(ctx, req, "CreateUserRequest"); err != nil {
        return nil, err
    }

    // Service logic
    user := &Account{Name: req.GetName()}

    // Success response
    return response.Created(ctx, user)
}

func (s *AccountService) RegisterWithHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
    return proto.RegisterAccountServiceHandler(ctx, mux, conn)
}
```

### Step 9: Graceful Shutdown

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

<-sigChan
log.Info("Shutdown signal received")

// Shutdown with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

grpcServer.Shutdown()
if err := gateway.Shutdown(ctx); err != nil {
    log.Error("Gateway shutdown error: %v", err)
}
```

### Service Integration Checklist

- [ ] Add framework to go.mod dependencies
- [ ] Create service-specific Makefile
- [ ] Load and validate environment variables
- [ ] Initialize logger with context
- [ ] Setup application with required integrations
- [ ] Create gRPC server with service
- [ ] Create HTTP gateway
- [ ] Register error codes and validators
- [ ] Implement service handlers
- [ ] Setup graceful shutdown
- [ ] Test with `make proto-gen`
- [ ] Run tests with `make unit-test`
- [ ] Run linter with `make lint`
- [ ] Generate coverage with `make cover`

---

## Development Guidelines

### When Extending the Framework

#### 1. Adding New Utility Functions

**Location**: `utils/generic/generic.go` for generic utilities

**Pattern**:
```go
// Descriptive comment explaining purpose
func NewFunction[T constraint](param T) result {
    // implementation
}

// Include unit tests
func TestNewFunction(t *testing.T) {
    // test cases
}
```

#### 2. Adding New Environment Variables

**Steps**:
1. Add constant to `env/env.go`:
   ```go
   const MyVariable = "MY_VARIABLE"
   ```

2. Update documentation in `claude.md`

3. Update `.env.example` in services using it

4. Update validation in services:
   ```go
   env.Validate(env.MyVariable)
   ```

#### 3. Adding New Response Error Codes

**Steps**:
1. Add constant to `response/constant.go`:
   ```go
   ErrMyCustomError ErrCode = 113
   ```

2. Add message to `response/error_code_msg.go`:
   ```go
   ErrMyCustomError: "Custom error description",
   ```

3. Document in services using it

#### 4. Adding New Server Features

**Pattern**:
1. Add method to `GRPCServer` or `Gateway`
2. Follow builder pattern (return *Server)
3. Test with integration tests
4. Update documentation

Example:
```go
func (s *GRPCServer) WithMetrics(metricsCollector MetricsCollector) *GRPCServer {
    // implementation
    return s
}
```

#### 5. Modifying Validation

**Pattern**:
```go
// In validate/validate.go
func ValidateCustom(ctx context.Context, req any) error {
    // Custom validation logic
    return nil
}

// Services call it
if err := validate.ValidateCustom(ctx, request); err != nil {
    return nil, err
}
```

### Testing Guidelines

#### Unit Tests

- Place in `*_test.go` files next to implementation
- Use table-driven tests for multiple cases
- Mock external dependencies
- Aim for >80% coverage

Example:
```go
func TestNewFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   interface{}
        want    interface{}
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   validInput,
            want:    expectedOutput,
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewFunction() error = %v", err)
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("NewFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

#### Running Tests

```bash
# Run all tests with coverage
make unit-test

# Generate HTML report
make cover

# Get coverage hints
make cover-hint
```

### Code Quality Standards

#### Linting

Run before committing:
```bash
make lint
```

Requirements:
- No undefined variables
- No unused variables (except `_`)
- Proper error handling
- No goroutine leaks
- Consistent formatting

#### Code Style

- Follow Go style guide (gofmt -s)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small

#### Documentation

- Add godoc comments to all exported functions
- Update `claude.md` when adding features
- Include usage examples in comments
- Document breaking changes

### Common Patterns

#### Builder Pattern (Used Throughout)

```go
result := New().
    With A(value1).
    WithB(value2).
    WithC(value3)
```

Allows optional configuration and clear API.

#### Error Handling

```go
result, err := response.Success(ctx, data)
if err != nil {
    // Already formatted as gRPC error
    return nil, err
}

result, err := response.InvalidArgument(ctx, nil, errCode)
if err != nil {
    // err is gRPC status error
    return nil, err
}
```

#### Context Usage

```go
// Add logger to context
ctx = logger.WithContext(ctx, log)

// Extract logger later
log := logger.FromContext(ctx)

// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

#### Validation Pattern

```go
// Define error mapping once
var validationErrors = map[response.ErrType]map[string]response.ErrCode{
    "RequestType": {
        "field_name": errorCode,
    },
}

// Register once
response.LoadErrCode(customErrors, validationErrors)

// Use in every handler
if err := validate.Request(ctx, request, "RequestType"); err != nil {
    return nil, err
}
```

---

## File Locations Reference

### Core Framework Files
- `/Users/mrpsycho/smart-kart/framework/application/application.go` - Application lifecycle
- `/Users/mrpsycho/smart-kart/framework/env/env.go` - Environment management
- `/Users/mrpsycho/smart-kart/framework/logger/logger.go` - Logging
- `/Users/mrpsycho/smart-kart/framework/pgx/driver.go` - PostgreSQL driver
- `/Users/mrpsycho/smart-kart/framework/server/grpc.go` - gRPC server
- `/Users/mrpsycho/smart-kart/framework/server/http.go` - HTTP gateway
- `/Users/mrpsycho/smart-kart/framework/server/profiler.go` - Performance profiler
- `/Users/mrpsycho/smart-kart/framework/response/constant.go` - Response constants
- `/Users/mrpsycho/smart-kart/framework/response/error_code_msg.go` - Error messages
- `/Users/mrpsycho/smart-kart/framework/response/grpc.go` - Response builders
- `/Users/mrpsycho/smart-kart/framework/validate/validate.go` - Request validation
- `/Users/mrpsycho/smart-kart/framework/validate/constant.go` - Validation constants
- `/Users/mrpsycho/smart-kart/framework/utils/generic/generic.go` - Generic utilities

### Build and Configuration
- `/Users/mrpsycho/smart-kart/framework/Makefile` - Root Makefile
- `/Users/mrpsycho/smart-kart/framework/make/proto.mk` - Proto build targets
- `/Users/mrpsycho/smart-kart/framework/go.mod` - Module dependencies
- `/Users/mrpsycho/smart-kart/framework/go.sum` - Locked versions
- `/Users/mrpsycho/smart-kart/framework/README.md` - Original README
- `/Users/mrpsycho/smart-kart/framework/claude.md` - This file

### Dependent Services
- `/Users/mrpsycho/smart-kart/account/` - Account service (imports framework)
- `/Users/mrpsycho/smart-kart/products/` - Products service (imports framework)
- `/Users/mrpsycho/smart-kart/proto/` - Proto definitions (used by framework)

---

## Quick Reference

### Importing Framework

```go
import (
    "github.com/smart-kart/framework/application"
    "github.com/smart-kart/framework/env"
    "github.com/smart-kart/framework/logger"
    "github.com/smart-kart/framework/pgx"
    "github.com/smart-kart/framework/response"
    "github.com/smart-kart/framework/server"
    "github.com/smart-kart/framework/validate"
)
```

### Most Common Functions

```go
// Environment
env.Get(key)
env.GetOrDefault(key, default)
env.LoadFromEnv(".env")
env.Validate(requiredKeys...)

// Logger
logger := logger.New()
logger.Info(msg)
logger.Error(msg)
ctx = logger.WithContext(ctx, logger)
log := logger.FromContext(ctx)

// Database
pgx.Init(ctx)
ds := pgx.GetDS()
pool := ds.GetPool()

// Server
grpc := server.NewGRPC().WithServiceServer(svc).WithGRPCRegistrar(register)
gateway, _ := server.NewGateway(ctx)
gateway, _ = gateway.WithServiceHandler(ctx, svc)

// Response
response.Success(ctx, data)
response.Created(ctx, data)
response.InvalidArgument(ctx, nil, errCode)
response.NotFound(ctx, nil, errCode)
response.InternalError(ctx, nil, errCode)

// Validation
validate.Request(ctx, req, errType)
validate.RegisterCustomValidators(customValidators)
```

---

## Troubleshooting

### "Cannot find module" errors

**Solution**: Update go.mod
```bash
go get github.com/smart-kart/framework
cd framework && make mod
```

### Build proto errors

**Solution**: Install buf and regenerate
```bash
go install github.com/bufbuild/buf/cmd/buf@latest
make proto-gen
```

### Validation errors not showing up

**Solution**: Ensure error codes are registered
```go
response.LoadErrCode(customErrors, validationErrors)
```

### Database connection failed

**Solution**: Check environment variables
```bash
echo $DB_HOST $DB_PORT $DB_USER $DB_NAME
# Verify PostgreSQL is running and credentials are correct
```

### HTTP server not accepting requests

**Solution**: Check CORS and gateway configuration
```bash
# Verify HTTP port is not in use
lsof -i :8080
# Verify gRPC server is running
lsof -i :50051
```

### Test coverage incomplete

**Solution**: Run coverage analysis
```bash
make cover
make cover-hint  # Shows which functions need tests
```

---

## Version History

- **v1.0.0** - Initial framework release
  - gRPC and HTTP server support
  - PostgreSQL driver with pooling
  - Environment and configuration management
  - Error handling and response builders
  - Request validation framework
  - Build system with Makefile targets
  - Logging interface
  - Performance profiler

---

## Support and Contribution

### Adding Features to Framework

1. Ensure changes are used by at least 2 services
2. Add comprehensive documentation
3. Include unit tests
4. Run `make lint` and `make cover`
5. Update this `claude.md` file

### Reporting Issues

Include:
- Go version
- OS and architecture
- Environment variables used
- Relevant code snippets
- Expected vs actual behavior

---

**This documentation is comprehensive and should answer any question about the Framework. Refer to the file paths and code examples provided throughout this document for specific implementations.**
