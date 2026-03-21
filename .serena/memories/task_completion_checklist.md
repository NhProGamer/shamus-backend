# Task Completion Checklist

## Before Committing

### 1. Code Quality
- [ ] Code follows naming conventions (PascalCase types, camelCase JSON)
- [ ] Errors use `apperrors` package when appropriate
- [ ] New exported types/functions have doc comments
- [ ] Dependencies injected via constructors (no globals)
- [ ] Input validation at entry points
- [ ] User input sanitized where needed

### 2. Security Checks
- [ ] No unsafe type assertions (use ok-pattern)
- [ ] Player-game validation for WebSocket commands
- [ ] Chat messages sanitized
- [ ] Atomic operations for consumable resources

### 3. Build & Test
```bash
# Format
go fmt ./...

# Static analysis
go vet ./...

# Tests
go test ./...

# Build
go build ./...
```

### 4. Lint (Optional but Recommended)
```bash
golangci-lint run
```

### 5. Quick One-Liner
```bash
go fmt ./... && go vet ./... && go test ./... && go build ./...
```

## Adding New Features

### New Domain Entity
- [ ] Create entity in `internal/domain/entities/`
- [ ] Add any related types (ID alias, enum, etc.)
- [ ] Add validation methods if needed
- [ ] Export in package

### New Service
- [ ] Define port interface in `internal/domain/ports/services.go`
- [ ] Implement in `internal/application/services/`
- [ ] Add to DI in `cmd/server/main.go`
- [ ] Wire to handlers if needed

### New Command (WebSocket)
- [ ] Add CommandType constant in `entities/command.go`
- [ ] Add payload struct in `entities/commands/payloads.go`
- [ ] Add handler in `websocket/command_handler.go`
- [ ] Add error mapping in `websocket/handler.go`
- [ ] Update `docs/api/asyncapi.yaml`

### New Notification
- [ ] Add NotificationType constant in `entities/notification.go`
- [ ] Add notify method in `services/notification_service.go`
- [ ] Update `docs/api/asyncapi.yaml`

### New Prompt Type
- [ ] Add PromptType constant if new
- [ ] Add payload struct in `entities/prompts/payloads.go`
- [ ] Add response struct in `entities/prompts/responses.go`
- [ ] Update `docs/api/asyncapi.yaml`

### New REST Endpoint
- [ ] Add handler in `http/controllers/`
- [ ] Add route in `http/routes/routes.go`
- [ ] Update `docs/api/openapi.yaml`

## Commit Message Format

```
<type>(<scope>): <description>

[optional body]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `refactor`: Code refactoring
- `test`: Tests
- `chore`: Maintenance

Examples:
```
feat(websocket): add kick_player command
fix(vote): validate voters in NewVote constructor
docs(api): add AsyncAPI specification
refactor(abilities): replace Consume with TryConsume
```

## PR Description Template

```markdown
## Summary
Brief description of changes

## Changes
- Change 1
- Change 2

## Testing
- [ ] Unit tests pass
- [ ] Manual testing done

## Documentation
- [ ] Code comments added
- [ ] API docs updated (if applicable)
```
