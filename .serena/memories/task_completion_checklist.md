# Task Completion Checklist

When completing a coding task, verify the following:

## 1. Code Quality
- [ ] Code follows project naming conventions (PascalCase types, camelCase JSON)
- [ ] Errors use `apperrors` package when appropriate
- [ ] New exported types/functions have doc comments
- [ ] Dependencies injected via constructors (no globals)

## 2. Testing
```bash
# Run tests to verify no regressions
go test ./...
```

## 3. Formatting
```bash
# Format all Go files
go fmt ./...
```

## 4. Static Analysis
```bash
# Check for common issues
go vet ./...
```

## 5. Dependencies
```bash
# Ensure go.mod is clean
go mod tidy
```

## 6. Build Check
```bash
# Verify project compiles
go build ./...
```

## Quick One-Liner
```bash
go fmt ./... && go vet ./... && go test ./... && go build ./...
```

## Before Committing
1. Run the quick one-liner above
2. Review `git diff` for unintended changes
3. Write a clear commit message describing the change
