.PHONY: help build build-optimized clean test run docker-build docker-run docker-stop docker-clean install lint fmt vet

# Variables
BINARY_NAME=pinger
DOCKER_IMAGE=pinger
DOCKER_TAG=latest
GO=go
GOFLAGS=-ldflags="-s -w" -trimpath

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary (standard build)
	$(GO) build -o $(BINARY_NAME) ./cmd/pinger/

build-optimized: ## Build the optimized binary (smaller size)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/pinger/
	@echo "Binary built: $(BINARY_NAME)"
	@ls -lh $(BINARY_NAME)

clean: ## Remove built binaries and temporary files
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(GO) clean

test: ## Run tests
	$(GO) test -v ./...

test-cover: ## Run tests with coverage
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

run: build-optimized ## Build and run the application locally
	./$(BINARY_NAME)

docker-build: ## Build Docker image (optimized)
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo ""
	@echo "Image built successfully!"
	@docker images $(DOCKER_IMAGE):$(DOCKER_TAG) --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

docker-run: ## Run Docker container
	docker run --rm --name $(BINARY_NAME) $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-run-detached: ## Run Docker container in detached mode
	docker run -d --name $(BINARY_NAME) $(DOCKER_IMAGE):$(DOCKER_TAG)
	@echo "Container started in background. Use 'make docker-logs' to view logs."

docker-logs: ## Show logs from running container
	docker logs -f $(BINARY_NAME)

docker-stop: ## Stop running Docker container
	docker stop $(BINARY_NAME) 2>/dev/null || true

docker-clean: docker-stop ## Remove Docker container and images
	docker rm $(BINARY_NAME) 2>/dev/null || true
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG) 2>/dev/null || true

docker-shell: ## Get a shell in the builder stage (for debugging)
	docker run --rm -it --entrypoint /bin/sh golang:1.26-alpine

install: build-optimized ## Install the binary to $GOPATH/bin
	$(GO) install $(GOFLAGS) ./cmd/pinger/

lint: ## Run linter (requires golangci-lint)
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Install with: brew install golangci-lint"; exit 1; }
	golangci-lint run ./...

fmt: ## Format Go code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

size: build-optimized ## Show binary size
	@echo "Binary size:"
	@ls -lh $(BINARY_NAME)
	@echo ""
	@echo "Detailed size breakdown:"
	@size $(BINARY_NAME) 2>/dev/null || echo "size command not available on this system"

docker-size: ## Compare Docker image sizes
	@echo "Docker Image Sizes:"
	@docker images --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}" | grep -E "(REPOSITORY|$(DOCKER_IMAGE))" || echo "No images found"

all: clean deps fmt vet test build-optimized ## Run all checks and build

release: clean deps test build-optimized docker-build ## Prepare a release (test, build binary and docker image)
	@echo ""
	@echo "Release build complete!"
	@echo "Binary: ./$(BINARY_NAME)"
	@echo "Docker: $(DOCKER_IMAGE):$(DOCKER_TAG)"


