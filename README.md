# Cozy Hub Framework

Shared build configurations and utilities for Cozy Hub services.

## Structure

```
framework/
└── make/
    └── proto.mk    # Common Makefile targets for proto generation, testing, linting
```

## Usage

### In Proto Repository

Include the framework Makefile in your proto repository:

```makefile
# proto/Makefile
include ../framework/make/proto.mk
```

### In Service Repositories

Include the framework Makefile in each service:

```makefile
# account/Makefile
include ../framework/make/proto.mk

# Service-specific variables
SERVICE_NAME := account
PROTO_DIR := ../proto

# Add service-specific targets
run:
	go run cmd/main.go
```

## Available Targets

### Proto Generation

- `make proto-gen` - Generate proto files, clean, run buf, add go tags, generate mocks, and tidy modules
- `make proto-gen-clean` - Clean generated proto files
- `make buf` - Generate code from proto files using buf
- `make go-tags` - Add go tags to generated files
- `make mock` - Generate mocks for all interfaces

### Testing

- `make unit-test` - Run all unit tests with coverage
- `make cover` - Generate HTML coverage report
- `make cover-hint` - Get coverage hints for prioritizing tests

### Code Quality

- `make lint` - Run Go linters (golangci-lint)
- `make mod` - Tidy and verify go.mod

### Maintenance

- `make clean` - Clean all caches
- `make update-dep` - Update all dependencies
- `make help` - Display available targets

## Environment Variables

You can override these variables in your service Makefile:

```makefile
PROTO_DIR := ../proto          # Path to proto directory
PROTO_GEN_DIR := proto/gen     # Generated code output directory
COVER_MODE := atomic           # Coverage mode
COVER_PROFILE := coverage.txt  # Coverage profile file
COVER_REPORT := coverage.html  # Coverage HTML report
GO_TEST_PKG := ./...          # Packages to test
GO_COVER_PKG := ./...         # Packages to cover
```

## Example: Account Service

```makefile
# account/Makefile
include ../framework/make/proto.mk

SERVICE_NAME := account

run:
	go run cmd/main.go

build:
	go build -o bin/$(SERVICE_NAME) cmd/main.go

test: unit-test

.DEFAULT_GOAL := help
```

Then you can run:

```bash
cd account
make proto-gen    # Generate protos and mocks
make lint         # Run linters
make test         # Run tests
make cover        # Generate coverage report
make run          # Run the service
```

## Tools Required

The Makefile will auto-install most tools, but you may want to install them manually:

```bash
# Buf (proto generation)
go install github.com/bufbuild/buf/cmd/buf@latest

# Mockery (mock generation)
go install github.com/vektra/mockery/v2@latest

# Linting
brew install golangci-lint  # macOS
# or
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh

# Coverage tools
go install github.com/gojekfarm/go-coverage@latest
go install github.com/axw/gocov/gocov@latest
go install github.com/matm/gocov-html/cmd/gocov-html@latest
```

## Integration with CI/CD

Example GitHub Actions workflow:

```yaml
name: Test

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Generate protos
        run: make proto-gen
        working-directory: ./proto

      - name: Lint
        run: make lint
        working-directory: ./account

      - name: Test
        run: make cover
        working-directory: ./account

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./account/coverage.txt
```
