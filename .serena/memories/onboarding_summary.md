# Shamus Backend - Comprehensive Onboarding Summary

## Project at a Glance

**Shamus Backend** is a production-grade Werewolf game backend implementing a real-time multiplayer game engine with:
- Clean hexagonal (ports & adapters) architecture
- WebSocket-based real-time communication (Command/Response/Prompt/Notification pattern)
- Redis for persistent state management
- OIDC-based authentication
- Comprehensive game logic and role mechanics
- OpenAPI + AsyncAPI documentation

**Key Stats**:
- Go 1.24
- 4-24 players per game
- 4 roles: Villager, Werewolf, Seer, Witch
- 3+ game phases per cycle: Night → Day → Vote
- 24-hour data retention
- 20+ notification types, 5 command types, 4 prompt types

## Architecture Overview

### Four-Layer Design

```
┌─────────────────────────────────────┐
│  Primary Adapters                    │
│  (HTTP Controllers, WebSocket)      │
├─────────────────────────────────────┤
│  Application Layer                   │
│  (Services, GameEngine Orchestrator) │
├─────────────────────────────────────┤
│  Domain Layer                        │
│  (Entities, Ports, Rules)           │
├─────────────────────────────────────┤
│  Secondary Adapters                  │
│  (Redis Repositories)               │
└─────────────────────────────────────┘
```

### Key Components

1. **Domain Entities** (Business Logic)
   - Game: State container with phases, players, roles
   - Player: Individual with role, connection state, vote
   - Role: Polymorphic interface (Seer, Werewolf, Witch, Villager)
   - Vote: Voting system with resolution logic
   - Prompt: Interactive requests with timeout
   - Command: Client-to-server actions
   - Notification: Server-to-client updates

2. **Application Services** (Business Rules)
   - GameService: Game creation and lifecycle
   - PlayerService: Player management and reconnection
   - GameEngine: Game flow orchestration (orchestrator pattern)
   - PromptService: Interactive prompts with timeouts
   - NotificationService: Server-to-client notifications
   - VoteService: Voting mechanics
   - NightService: Night phase coordination
   - TimerService: Phase and role timing
   - VisibilityService: Role-based information filtering
   - ChatService: Channel permissions and routing

3. **Primary Adapters**
   - HTTP: REST API (game creation, health checks)
   - WebSocket Handler: Connection lifecycle
   - Command Handler: Process client commands
   - Session Manager: Player-session mapping

4. **Secondary Adapters**
   - Redis Repositories: Game, Player, Vote persistence

## Critical Data Flows

### Game Lifecycle

```
1. CREATE    → HTTP POST /app/api/v1/game       → Game(status:waiting)
2. JOIN      → WebSocket /app/ws/{gameID}       → Player joins
3. CONFIGURE → Command: update_settings          → Host updates roles
4. START     → Command: start_game               → Roles assigned, night begins
5. PLAY      → Prompts/Responses                 → Actions, votes, phases
6. END       → Win condition met                 → Game status:ended
```

### WebSocket Channel Architecture

```
Server → Client:
├─ notification: Informational (game_state, player_joined, chat_message, etc.)
└─ prompt: Action required (vote, select_player, select_option, confirm)

Client → Server:
├─ command: Player actions (send_chat, update_settings, start_game, etc.)
└─ response: Prompt answers (vote result, selection)
```

### Phase Cycle (Active Game)

```
NIGHT (sequential sub-phases)
├─ Seer (30s): Choose target to see
├─ Werewolf (1 min): Group vote to kill
└─ Witch (45s): Heal victim or poison another

DAY (3 min)
└─ Discussion: All alive players chat

VOTE (2 min)
└─ Village vote: Eliminate by majority

[Back to NIGHT if game continues]
```

## Important Concepts

### Dependency Injection Resolution

Circular dependencies handled via:
1. Create components without circular deps
2. Complete wiring with setter methods
3. Interface segregation (ConnectionChecker, PlayerDisconnecter)

```go
wsHandler.SetPlayerService(playerService)
commandHandler.SetGameEngine(gameEngine)
commandHandler.SetDisconnecter(wsHandler)
```

### Security Measures

- **Player-Game Validation**: CommandHandler validates player belongs to game
- **Chat Sanitization**: Control characters removed (helpers.SanitizeChatMessage)
- **Safe Type Assertions**: Session helpers with ok-pattern checks
- **Ability Consumption**: TryConsume() prevents underflow
- **Vote Validation**: NewVote() validates non-empty voters

### Visibility Rules

Role information visibility depends on:
- **Player Role**: Always see own role
- **Werewolves**: See other werewolves
- **Dead Players**: Roles revealed during day/vote phases
- **Others**: Only see username and alive status

### Win Conditions

Checked after each phase:
- **Villagers Win**: No werewolves alive
- **Werewolves Win**: Werewolves >= Villagers (alive)
- **Draw**: No players alive
- **Ongoing**: Neither condition met

## External Dependencies

| Library | Purpose | Critical? |
|---------|---------|-----------|
| Gin | HTTP routing | Yes |
| Melody | WebSocket management | Yes |
| Redis | State persistence | Yes |
| go-oidc | OIDC authentication | Yes |
| Zerolog | Logging | No |
| UUID | ID generation | Yes |

## Development Recommendations

### For Feature Development

1. **New Game Action**:
   - Add validation to Game/Player entity
   - Implement in GameEngine
   - Create Prompt if interactive
   - Emit notifications via NotificationService

2. **New Role**:
   - Create Role implementation (`entities/roles/`)
   - Add RoleType enum
   - Add to NightPhaseOrder if active at night
   - Add visibility rules

3. **New Command**:
   - Add CommandType constant
   - Add payload struct (`entities/commands/`)
   - Add handler in CommandHandler

### Commands

```bash
go build ./...             # Build
go test ./...              # Run all tests
go run cmd/server/main.go  # Start server
golangci-lint run          # Lint check
go fmt ./... && go vet ./... && go test ./... && go build ./...  # Full check
```

## Critical Files to Know

| Path | Purpose |
|------|---------|
| `cmd/server/main.go` | Entry point, DI wiring |
| `internal/domain/entities/` | Core entities |
| `internal/domain/ports/` | Interface contracts |
| `internal/application/services/` | Service implementations |
| `internal/application/orchestration/game_engine_v2.go` | Game orchestrator |
| `internal/adapters/primary/websocket/handler.go` | WebSocket lifecycle |
| `internal/adapters/primary/websocket/command_handler.go` | Command processing |
| `internal/adapters/secondary/redis/` | Redis repositories |
| `docs/api/openapi.yaml` | REST API spec |
| `docs/api/asyncapi.yaml` | WebSocket API spec |

## API Documentation

In debug mode (`config.Debug: true`):
- `/docs/rest` - Swagger UI (REST)
- `/docs/ws` - AsyncAPI UI (WebSocket)
- `/docs/api/*` - Raw spec files

## Memory Files Reference

| File | Content |
|------|---------|
| `architecture_deep_dive` | Layer details, directory structure |
| `key_types_and_interfaces` | Type definitions, port interfaces |
| `integration_and_data_flow` | Data flows, service interactions |
| `project_overview` | Quick overview |
| `code_style_conventions` | Naming, formatting |
| `suggested_commands` | Build, test, lint commands |
| `testing_and_validation` | Test patterns, validation layers |
| `api_documentation` | API docs reference |

## Key Insights

1. **Real-Time is Event-Driven**: Services emit notifications, handlers deliver
2. **State is in Redis**: 24-hour TTL, authoritative source
3. **Visibility is Role-Based**: Different rules for different roles/phases
4. **Night Phase is Sequential**: Seer → Werewolf → Witch (priority order)
5. **Circular Deps Resolved Elegantly**: Interface segregation + deferred wiring
6. **Role Assignment is Random**: Fisher-Yates shuffle before start
7. **Timeouts Matter**: 2-min reconnect, 3-min day, 2-min vote, 30-90sec roles

---

**Updated**: March 2026
**Version**: 2.0
**Status**: Complete Onboarding
