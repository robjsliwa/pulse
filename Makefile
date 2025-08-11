SHELL := /bin/bash

.PHONY: setup tidy lint vet test swagger swagger-ci run build tools fmt

setup: tidy tools

tidy:
	go mod tidy

tools:
	@which golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1
	@which swag >/dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@v1.16.3

lint:
	golangci-lint run

vet:
	go vet ./...

test:
	go test ./... -race -count=1

# Local swagger generation; deterministic output recommended for CI
swagger:
	swag init -g cmd/pulse/main.go --parseDependency --parseInternal --generatedTime=false

swagger-ci:
	swag init -g cmd/pulse/main.go --parseDependency --parseInternal --generatedTime=false
	git diff --exit-code docs/

run:
	go run ./cmd/pulse

build:
	go build ./...

fmt:
	gofmt -s -w .

