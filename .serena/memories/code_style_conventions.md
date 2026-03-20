# Code Style & Conventions

## Naming Conventions
- **Packages**: lowercase, single word (e.g., `entities`, `app`, `infra`)
- **Types**: PascalCase (e.g., `GameService`, `PlayerID`, `GamePhase`)
- **Type aliases**: Use custom types for IDs (e.g., `type GameID string`, `type PlayerID string`)
- **Interfaces**: Named by their purpose, not with "I" prefix (e.g., `GameService`, not `IGameService`)
- **Errors**: Prefixed with `Err` (e.g., `ErrGameNotFound`, `ErrPlayerDead`)
- **Constants**: PascalCase for exported, camelCase for internal

## File Organization
- One file per major concept (e.g., `game.go`, `player.go`, `role.go`)
- Related implementations in subdirectories (e.g., `roles/seer.go`, `abilities/seer.go`)
- Tests alongside code with `_test.go` suffix

## Error Handling
- Use structured errors from `internal/domain/errors` package (apperrors)
- Errors have `Code` (machine-readable) and `Message` (human-readable)
- Wrap errors with context using `apperrors.Wrap()`
- Pre-defined errors as package-level vars (e.g., `ErrGameNotFound`)

## Interfaces (Ports)
- Defined in `internal/domain/ports/` package
- Each interface has doc comments explaining its purpose
- Methods have doc comments describing behavior

## Dependency Injection
- Constructor functions: `NewXxxService(deps...) *XxxService`
- Dependencies passed explicitly (no global state)
- Circular dependencies broken with setter injection (e.g., `wsHandler.SetPlayerService()`)

## JSON Tags
- Use `json:"camelCase"` for API serialization
- Consistent with TypeScript frontend conventions

## Logging
- Use `pkg/logger` wrapper around zerolog
- Context-enriched loggers: `log.WithGameID()`, `log.WithPlayerID()`
- Levels: Debug, Info, Warn, Error, Fatal

## Comments
- Doc comments on exported types/functions
- Inline comments for complex logic
- Method comments describe behavior and rules
