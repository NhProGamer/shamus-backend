# Suggested Commands

## Build & Run
```bash
# Build the server
go build -o server ./cmd/server

# Run the server (requires config.yaml and Redis)
./server

# Build and run in one step
go run ./cmd/server
```

## Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/domain/entities/...
go test ./internal/domain/errors/...

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

# Run linter (if golangci-lint installed)
golangci-lint run

# Tidy dependencies
go mod tidy
```

## Dependencies
```bash
# Download dependencies
go mod download

# Update dependencies
go get -u ./...

# Verify dependencies
go mod verify
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

## System Utilities (Linux)
```bash
ls -la
find . -name "*.go"
grep -r "pattern" --include="*.go"
```
