# Suggested Commands

## Build & Run

```bash
# Build the server
go build -o server ./cmd/server

# Run the server (requires config.yaml and Redis)
./server

# Build and run in one step
go run ./cmd/server

# Build all packages (verify compilation)
go build ./...
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/domain/entities/...
go test ./internal/application/services/...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Code Quality

```bash
# Format code
go fmt ./...
gofmt -w .

# Vet code (static analysis)
go vet ./...

# Run linter (golangci-lint)
golangci-lint run

# Run linter with auto-fix
golangci-lint run --fix
```

## Dependencies

```bash
# Download dependencies
go mod download

# Update dependencies
go get -u ./...

# Tidy dependencies (remove unused)
go mod tidy

# Verify dependencies
go mod verify
```

## Quick One-Liner (Full Check)

```bash
go fmt ./... && go vet ./... && go test ./... && go build ./...
```

## Full Check with Lint

```bash
go fmt ./... && go vet ./... && golangci-lint run && go test ./... && go build ./...
```

## Git

```bash
git status
git diff
git add .
git commit -m "message"
git push
git pull
```

## API Documentation (Debug Mode)

When running with `debug: true` in config:
- REST docs: http://localhost:8080/docs/rest
- WebSocket docs: http://localhost:8080/docs/ws

## Redis

```bash
# Connect to Redis CLI
redis-cli

# Check keys
redis-cli KEYS "game:*"
redis-cli KEYS "player:*"

# Get a game
redis-cli GET "game:<gameID>"

# Clear all (DANGER)
redis-cli FLUSHDB
```

## Docker (if applicable)

```bash
# Build image
docker build -t shamus-backend .

# Run container
docker run -p 8080:8080 shamus-backend

# Docker compose
docker-compose up -d
```

## Task Completion Checklist

Before committing:
```bash
# 1. Format
go fmt ./...

# 2. Vet
go vet ./...

# 3. Lint (if installed)
golangci-lint run

# 4. Test
go test ./...

# 5. Build
go build ./...

# 6. Review changes
git diff

# 7. Commit
git add .
git commit -m "feat/fix/docs: description"
```
