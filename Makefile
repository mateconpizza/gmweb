# gomarks - simple bookmark manager
# See LICENSE file for copyright and license details.

BINARY_NAME 	:= gmweb
FN		?= .

all: lint check test build

# Run tests
test:
	@echo '>> Testing $(BINARY_NAME)'
	@go test ./...
	@echo

# Run tests with gotestsum
testsum:
	@echo '>> Testing gmweb with gotestsum'
	@gotestsum --format pkgname --hide-summary=skipped --format-icons codicons

# Lint code with 'golangci-lint'
lint:
	@echo '>> Linting code'
	@go vet ./...
	golangci-lint run ./...

# Run tests for a specific function
testfn:
	@echo '>> Testing function $(FN)'
	@go test -run $(FN) ./...

tidy:
	go mod tidy

run:
	go run . -a :8083 --dev -vvvvvvv

.PHONY: all build test clean full lint testfn
