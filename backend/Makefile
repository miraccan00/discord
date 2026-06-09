.PHONY: help lint lint-fix build run test tidy

GOLANGCI_LINT_VERSION := v1.62.2

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-12s %s\n", $$1, $$2}'

lint: ## Run golangci-lint
	golangci-lint run ./...

lint-fix: ## Run golangci-lint with autofix
	golangci-lint run --fix ./...

build: ## Build the server binary
	go build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server

run: build ## Build and run the server
	./bin/server

test: ## Run tests with the race detector
	go test -race -v ./...

tidy: ## Tidy go.mod / go.sum
	go mod tidy
