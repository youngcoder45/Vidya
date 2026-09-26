.PHONY: dev build run test lint vet migrate tidy clean

# Local dev with hot reload (requires air: go install github.com/air-verse/air@latest)
dev:
	@which air >/dev/null 2>&1 && air || go run ./cmd/server

build:
	go build -o bin/vidya-api ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./... -count=1

vet:
	go vet ./...

# Apply versioned SQL migrations to the configured database.
migrate:
	go run ./cmd/migrate

# golangci-lint (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
lint:
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed — run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

tidy:
	go mod tidy

clean:
	rm -rf bin
