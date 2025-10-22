# Proto generation Makefile
# Include this file in your service Makefile
# Example: include ../../framework/make/proto.mk

# Variables
PROTO_FILES := $(shell find proto -name "*.proto" 2>/dev/null)
GOOGLEAPIS_DIR := ./googleapis
PROTO_GEN_DIR := proto/gen
COVER_MODE := atomic
COVER_PROFILE := coverage.txt
COVER_REPORT := coverage.html
GO_TEST_PKG := ./...
GO_COVER_PKG := ./...
GO_FMT := gofmt -s -w .
GO_LINT := golangci-lint run --timeout 5m

.PHONY: proto-gen proto-gen-clean buf go-tags mock mod unit-test cover-hint cover lint update-dep clean

## Generate proto files, clean, run buf, add go tags, generate mocks, and tidy modules
proto-gen: proto-gen-clean buf go-tags mock mod

## Clean generated proto files
proto-gen-clean:
	@echo "Cleaning generated proto files..."
	@rm -rf $(PROTO_GEN_DIR)

## Generate code from proto files using buf
buf:
	@echo "Generating code from proto files using buf..."
	@command -v buf >/dev/null 2>&1 || { \
		echo "buf not found, installing..."; \
		go install github.com/bufbuild/buf/cmd/buf@latest; \
	}
	@cd proto && buf generate

## Alternative: Generate using protoc (uncomment if you prefer protoc over buf)
# proto-gen-protoc:
# 	@echo "Generating code from proto files using protoc..."
# 	protoc -Iproto -I$(GOOGLEAPIS_DIR) \
# 		--go_out=$(PROTO_GEN_DIR) --go_opt=paths=source_relative \
# 		--go-grpc_out=$(PROTO_GEN_DIR) --go-grpc_opt=paths=source_relative \
# 		--grpc-gateway_out=$(PROTO_GEN_DIR) --grpc-gateway_opt=paths=source_relative \
# 		$(PROTO_FILES)

## Add go tags to generated files using protoc-go-inject-tag
go-tags:
	@echo "Adding go tags..."
	@command -v protoc-go-inject-tag >/dev/null 2>&1 || { \
		echo "protoc-go-inject-tag not found, installing..."; \
		go install github.com/favadi/protoc-go-inject-tag@latest; \
	}
	@cd proto && for file in $$(find gen/go -name "*.pb.go" -type f 2>/dev/null); do \
		protoc-go-inject-tag -input=$$file -remove_tag_comment; \
	done

## Generate mocks for all the interfaces
mock:
	@echo "Generating mocks..."
	@command -v mockery >/dev/null 2>&1 || { \
		echo "mockery not found, installing..."; \
		go install github.com/vektra/mockery/v2@latest; \
	}
	@rm -rf mocks
	@mockery --all --keeptree --case underscore --with-expecter --exported

## Organize and clean up go.mod and go.sum
mod:
	@echo "Tidying go.mod..."
	@go mod tidy
	@go mod verify

# -race: Detect if the code being tested has any race conditions, where multiple goroutines access
# and modify shared memory concurrently without proper synchronization.
# -mod=readonly: Prohibits go build from modifying go.mod
## Run all the tests by excluding auto-generated/conditional packages
unit-test:
	@echo "Running unit tests..."
	@go clean -testcache
	@go test -ldflags="-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=ignore" ${GO_TEST_PKG} -mod=readonly -cover -covermode=${COVER_MODE} -coverprofile=${COVER_PROFILE} -coverpkg=${GO_COVER_PKG}

# provides the sorted list of functions to cover and the impact associated with covering it.
## Prioritize which sections of code to unit test first
cover-hint: unit-test
	@echo "Generating coverage hints..."
	@go-coverage -h >/dev/null 2>&1 || (go get github.com/gojekfarm/go-coverage && \
		go install github.com/gojekfarm/go-coverage && make mod)
	@go-coverage -f ${COVER_PROFILE} --line-filter 3

## Run the tests & prepare a coverage report in .html file
cover: unit-test
	@echo "Generating coverage report..."
	@gocov-html -lt >/dev/null 2>&1 || (go install github.com/axw/gocov/gocov@latest && \
		go install github.com/matm/gocov-html/cmd/gocov-html@latest && make mod)
	@gocov convert ${COVER_PROFILE} | gocov-html > ${COVER_REPORT}
	@echo "Coverage report generated: ${COVER_REPORT}"

## Run aggregated Go linters
lint:
ifeq ($(shell which golangci-lint),)
ifeq ($(shell uname),Darwin)
	@echo "Installing golangci-lint via brew..."
	@brew install golangci-lint
else
	@echo "Installing golangci-lint..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ~/.local/bin
endif
endif
	@echo "Running linters..."
	@${GO_FMT}
ifeq ($(GIT_HOOK), 1)
	@${GO_LINT}
else
	@fx -v >/dev/null 2>&1 || (go install github.com/antonmedv/fx@24.1.0 && make mod)
	@${GO_LINT} --output-format code-climate > "${PWD}/lint.json" || true; \
		cat "${PWD}/lint.json" | fx . || true; \
		rm -f "${PWD}/lint.json"
endif

## Recursively update packages including test dependencies
update-dep: clean
	@echo "Updating dependencies..."
	@go get -u all
	@make mod

## Clean up caches
clean:
	@echo "Cleaning caches..."
	@golangci-lint cache clean 2>/dev/null || true
	@go clean -cache
	@go clean -testcache
	@go clean -modcache
	@cat /dev/null > go.sum
	@make mod

## Display help
help:
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
