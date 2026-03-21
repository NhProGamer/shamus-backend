# Code Style & Conventions

## Naming Conventions

- **Packages**: lowercase, single word (e.g., `entities`, `services`, `redis`)
- **Types**: PascalCase (e.g., `GameService`, `PlayerID`, `GamePhase`)
- **Type aliases**: Custom types for IDs (e.g., `type GameID string`, `type PlayerID string`)
- **Interfaces**: Named by purpose, no "I" prefix (e.g., `GameService`, not `IGameService`)
- **Errors**: Prefixed with `Err` (e.g., `ErrGameNotFound`, `ErrPlayerDead`)
- **Constants**: PascalCase for exported, camelCase for internal

## Directory Structure

```
internal/
├── domain/           # Core business logic (NO external deps)
│   ├── entities/     # Domain objects
│   ├── ports/        # Interface contracts
│   ├── errors/       # Structured errors
│   ├── helpers/      # Domain utilities
│   └── constants/    # Game configuration
├── application/      # Use cases
│   ├── services/     # Service implementations
│   └── orchestration/ # GameEngine
├── adapters/
│   ├── primary/      # Driving (HTTP, WebSocket)
│   └── secondary/    # Driven (Redis)
└── infrastructure/   # Cross-cutting (config)
```

## File Organization

- One file per major concept (e.g., `game.go`, `player.go`, `vote.go`)
- Related implementations in subdirectories (e.g., `roles/seer.go`, `abilities/seer.go`)
- Tests alongside code with `_test.go` suffix
- Payloads in dedicated files (e.g., `commands/payloads.go`, `prompts/payloads.go`)

## Error Handling

- Use structured errors from `internal/domain/errors` (apperrors)
- Errors have `Code` (machine-readable) and `Message` (human-readable)
- Pre-defined errors as package-level vars (e.g., `ErrGameNotFound`)
- Constructor validation returns errors (e.g., `NewVote() (*Vote, error)`)

## Interfaces (Ports)

- Defined in `internal/domain/ports/` package
- Grouped by concern: `services.go`, `repositories.go`
- Each interface has doc comments explaining purpose
- Methods have context.Context as first param for repos/services

## Dependency Injection

- Constructor functions: `NewXxxService(deps...) *XxxService`
- Dependencies passed explicitly (no global state)
- Circular dependencies broken with setter injection:
  ```go
  wsHandler.SetPlayerService(playerService)
  commandHandler.SetGameEngine(gameEngine)
  ```

## JSON Serialization

- Use `json:"camelCase"` for API serialization
- Use `json:"-"` for internal-only fields (e.g., `Role` interface)
- Use `json:",omitempty"` for optional fields

## Logging

- Use `pkg/logger` wrapper around zerolog
- Context-enriched loggers: `WithGameID()`, `WithPlayerID()`
- Levels: Debug, Info, Warn, Error, Fatal

## Security Patterns

- **Input sanitization**: `helpers.SanitizeChatMessage()` for user input
- **Safe type assertions**: Use ok-pattern in session helpers
- **Validation first**: Validate at entry points (CommandHandler, PromptService)
- **Atomic operations**: `TryConsume() bool` instead of separate check+consume

## Comments

- Doc comments on exported types/functions
- Inline comments for complex logic
- Method comments describe behavior and rules
- Context comments for non-obvious decisions

## Testing

- Table-driven tests with `tests := []struct{...}`
- Subtests with `t.Run(tt.name, func(t *testing.T) {...})`
- Test files next to source: `game.go` → `game_test.go`
