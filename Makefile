.PHONY: all proto build run test migrate clean help

# Variables
BINARY_NAME=product-catalog-service
PROTO_DIR=proto/product/v1
GO_FILES=$(shell find . -name '*.go' -type f)
SPANNER_PROJECT=test-project
SPANNER_INSTANCE=test-instance
SPANNER_DATABASE=product-catalog
SPANNER_EMULATOR_HOST=localhost:9010

# Default target
all: proto build

## help: Display this help message
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  /'

## proto: Generate Go code from proto files
proto:
	@echo "Generating protobuf code..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/product/v1/*.proto

## build: Build the server binary
build: proto
	@echo "Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/server

## run: Run the server
run: build
	@echo "Starting server..."
	@SPANNER_DATABASE=projects/$(SPANNER_PROJECT)/instances/$(SPANNER_INSTANCE)/databases/$(SPANNER_DATABASE) \
		SPANNER_EMULATOR_HOST=$(SPANNER_EMULATOR_HOST) \
		PORT=50051 \
		./bin/$(BINARY_NAME)

## test: Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

## test-e2e: Run E2E tests only
test-e2e:
	@echo "Running E2E tests..."
	@go test -v ./tests/e2e/...

## migrate: Apply database migrations
migrate:
	@echo "Applying migrations to Spanner emulator..."
	@gcloud config set auth/disable_credentials true
	@gcloud config set project $(SPANNER_PROJECT)
	@gcloud spanner instances create $(SPANNER_INSTANCE) --config=emulator-config --description="Test Instance" || true
	@gcloud spanner databases create $(SPANNER_DATABASE) --instance=$(SPANNER_INSTANCE) || true
	@gcloud spanner databases ddl update $(SPANNER_DATABASE) --instance=$(SPANNER_INSTANCE) --ddl="$(cat migrations/001_initial_schema.sql)"

## spanner-up: Start Spanner emulator
spanner-up:
	@echo "Starting Spanner emulator..."
	@docker-compose up -d spanner-emulator
	@echo "Waiting for emulator to be ready..."
	@sleep 5

## spanner-down: Stop Spanner emulator
spanner-down:
	@echo "Stopping Spanner emulator..."
	@docker-compose down

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f $(PROTO_DIR)/*.pb.go

## deps: Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

## lint: Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

## mod-verify: Verify dependencies
mod-verify:
	@echo "Verifying dependencies..."
	@go mod verify
