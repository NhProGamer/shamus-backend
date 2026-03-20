# Shamus Backend - Comprehensive Onboarding Summary

## Project at a Glance

**Shamus Backend** is a production-grade Werewolf game backend implementing a real-time multiplayer game engine with:
- Clean hexagonal (ports & adapters) architecture
- WebSocket-based real-time communication
- Redis for persistent state management
- OIDC-based authentication
- Comprehensive game logic and role mechanics

**Key Stats**:
- Go 1.24
- 4-24 players per game
- 4 roles: Villager, Werewolf, Seer, Witch
- 3+ game phases per cycle: Night → Day → Vote
- 24-hour data retention

## Architecture Overview

### Four-Layer Design

```
┌─────────────────────────────────────┐
│  Infrastructure Layer                │
│  (HTTP, WebSocket, Config, Routes)  │
├─────────────────────────────────────┤
│  Adapters Layer                      │
│  (Services, Repositories, WebSocket) │
├─────────────────────────────────────┤
│  Domain Layer                        │
│  (Entities, Ports, Rules)           │
├─────────────────────────────────────┤
│  External Dependencies               │
│  (Redis, OIDC, Gin, Melody)         │
└─────────────────────────────────────┘
```

### Key Components

1. **Domain Entities** (Business Logic)
   - Game: State container with phases, players, roles
   - Player: Individual with role, connection state, vote
   - Role: Polymorphic interface (Seer, Werewolf, Witch, Villager)
   - Vote: Voting system with resolution logic
   - Event: WebSocket message structure

2. **Application Services** (Business Rules)
   - GameService: Game creation and lifecycle
   - PlayerService: Player management and reconnection
   - GameEngine: Game flow orchestration (orchestrator pattern)
   - VisibilityService: Role-based information filtering
   - ChatService: Channel permissions and routing
   - TimerService: Phase and role timing
   - VoteService: Voting mechanics
   - NightService: Night phase coordination

3. **Infrastructure Adapters**
   - WebSocketHandler: Real-time communication (implements Broadcaster)
   - RedisGameRepo: Game persistence
   - RedisPlayerRepo: Player persistence
   - VoteRepository: In-memory vote storage
   - Controllers: HTTP endpoints
   - Routes: URL routing

4. **External Integration**
   - Redis: State persistence with 24h TTL
   - Gin: HTTP framework
   - Melody: WebSocket library
   - OIDC (coreos/go-oidc): Authentication
   - Zerolog: Logging

## Critical Data Flows

### Game Lifecycle

```
1. CREATE    → HTTP POST /api/v1/game              → Game(status:waiting)
2. JOIN      → WebSocket /ws/{gameID}              → Player joins, adds to game
3. CONFIGURE → WebSocket game_event:settings       → Host updates roles
4. START     → WebSocket game_event:start_game     → Roles assigned, night begins
5. PLAY      → WebSocket game_event:*              → Actions, votes, phases
6. END       → Win condition met                   → Game status:ended
```

### Phase Cycle (Active Game)

```
NIGHT (1 min total)
├─ Seer (30s): Choose target to see
├─ Werewolf (1 min): Vote to kill
└─ Witch (45s): Heal or poison

DAY (3 min)
├─ Discuss: All players chat
└─ Vote (2 min): Eliminate by majority

[Back to NIGHT if game continues]
```

### Key Integration Points

1. **Player Joins** → PlayerService → GameRepo/PlayerRepo → Redis
2. **Game Starts** → GameService → RoleAssignment → GameEngine → TimerService
3. **Action Occurs** → GameEngine → Validation → StateUpdate → Broadcast
4. **Timer Expires** → TimerService → GameEngine → Phase Transition
5. **Vote Resolves** → VoteService → PlayerUpdate → WinCheck → End/Continue
6. **Player Disconnects** → PlayerService → 2-min Timeout → Mark Inactive

## Important Concepts

### Dependency Injection Resolution

The codebase handles circular dependencies elegantly:
- WebSocketHandler created first without dependencies
- Services created independently
- Later: SetPlayerService() and SetGameEngine() complete the wiring
- ConnectionChecker interface breaks the cycle (interface segregation)

### Visibility Rules

Role information visibility depends on:
- **Player Role**: Always see own role
- **Werewolves**: See other werewolves
- **Dead Players**: Roles revealed during day/vote phases
- **Others**: Only see username and alive status

### Vote Mechanics

- Village votes: All alive players vote (majority elimination)
- Werewolf votes: Only werewolves vote (selection)
- Ties: No elimination if multiple targets tied
- Abstention: Allowed for village vote (counts as no vote)

### State Persistence

- Redis TTL: 24 hours for all data
- Key Format: game:{id}, player:{id}, game:{id}:player_ids
- Batch Operations: PlayerRepo uses Redis pipeline for efficiency
- In-Memory: Votes and night state stored temporarily, not persisted

### Win Conditions

Checked after each phase:
- **Villagers Win**: No werewolves alive
- **Werewolves Win**: Werewolves >= Villagers (alive)
- **Draw**: No players alive
- **Ongoing**: Neither condition met

## External Dependencies

### Framework & Libraries

| Library | Purpose | Critical? |
|---------|---------|-----------|
| Gin | HTTP routing | Yes |
| Melody | WebSocket management | Yes |
| Redis | State persistence | Yes |
| go-oidc | OIDC authentication | Yes |
| Zerolog | Logging | No (replaceable) |
| UUID | ID generation | Yes |

### Configuration Needs

- OIDC Provider: Issuer URL, Client ID, Secret
- Redis: Host, Port, Password, Database
- Server: Host, Port, Public URL
- CORS: Frontend origins (localhost:3000, :5173)

## Testing Insights

### Current Testing
- Table-driven unit tests in domain layer
- Game.CanStart() validation extensively tested
- Error code mapping tested

### Test Gaps
- No integration tests (end-to-end flow)
- No service orchestration tests
- No WebSocket routing tests
- No concurrent action tests
- No Redis persistence tests

### Validation Layers
1. Controller: Extract, validate request
2. Service: Business rule enforcement
3. Repository: Persistence guarantees
4. Client: Error code handling

## Development Recommendations

### For Feature Development

1. **New Game Action**:
   - Add validation to Game/Player entity
   - Implement in GameEngine
   - Add to WebSocketHandler routing
   - Emit events via Broadcaster
   - Handle visibility if role-specific

2. **New Role**:
   - Create Role implementation (entities/roles/)
   - Add RoleType enum
   - Create factory method (factories/role_factory.go)
   - Add to NightPhaseOrder if active at night
   - Add visibility rules (VisibilityService)

3. **New Service**:
   - Define Port interface (domain/ports/)
   - Implement in app layer
   - Inject in main.go
   - Wire to WebSocketHandler if needed
   - Ensure broadcasts through Broadcaster

### Code Quality Tools Available

- golangci-lint: Installed and available
- gopls: Language server for IDE support

### Key Commands

```bash
go test ./...              # Run all tests
go run cmd/server/main.go  # Start server
golangci-lint run          # Lint check
```

## Critical Files to Know

| Path | Purpose |
|------|---------|
| cmd/server/main.go | Entry point, DI wiring |
| internal/domain/entities/game.go | Game state |
| internal/domain/entities/player.go | Player state |
| internal/domain/ports/*.go | Interface contracts |
| internal/adapters/app/game_engine.go | Game orchestration |
| internal/adapters/api/ws/websocket_handler.go | Real-time communication |
| internal/adapters/infra/*_repo.go | Persistence |
| internal/infrastructure/routes/routes.go | HTTP routes |
| internal/domain/constants/constants.go | Game configuration |
| internal/domain/errors/errors.go | Error definitions |

## Quick Reference: Error Scenarios

```
New Player Join    → Game.status == waiting
Game Start         → Game.CanStart() validation
Seer Vision        → Game.phase == night, player alive, target alive
Werewolf Kill Vote → All werewolves voted
Witch Action       → Heal only if victim exists
Village Vote       → Game.phase == vote, player alive
Win Check          → After each phase transition
Disconnect         → 2-min timeout before marking inactive
```

## Memory Files Created

This onboarding includes four comprehensive memory files:

1. **architecture_deep_dive.md**: Detailed layer-by-layer breakdown with component relationships
2. **key_types_and_interfaces.md**: Complete type definitions and port interfaces
3. **integration_and_data_flow.md**: How components interact and data flows through system
4. **testing_and_validation.md**: Testing patterns and validation approach

Refer to these for:
- Architecture questions → architecture_deep_dive.md
- Type definitions → key_types_and_interfaces.md
- Data flow questions → integration_and_data_flow.md
- Testing/validation → testing_and_validation.md

## Next Steps for Development

1. **Understand the Flow**: Read integration_and_data_flow.md and trace a complete game cycle
2. **Set Up Dev Environment**: Ensure Redis and OIDC provider configured
3. **Run Tests**: `go test ./...` to verify setup
4. **Study a Service**: Start with GameService for simple pattern
5. **Read WebSocket Handler**: Understand message routing
6. **Explore GameEngine**: Learn game orchestration logic
7. **Check Domain Rules**: Review constants and validation logic

## Architecture Philosophy

The codebase embodies these principles:
- **Domain-Driven Design**: Business logic isolated in domain layer
- **Hexagonal Architecture**: Infrastructure as plugin adapters
- **Interface Segregation**: Small, focused interfaces (Broadcaster, PlayerSender)
- **Dependency Inversion**: Services depend on ports, not implementation
- **Event-Driven**: Services emit events, handlers decide delivery
- **Repository Pattern**: Abstract persistence behind interface
- **Single Responsibility**: Each service handles one concern

This architecture allows:
- Easy testing (mock implementations)
- Technology swapping (Redis → different DB)
- Parallel development (teams work on different layers)
- Clear separation of concerns
- Business logic reusability across interfaces

## Key Insights

1. **Real-Time is Event-Driven**: Services emit events to Broadcaster, which sends to clients
2. **State is Authoritative in Redis**: Memory caches, but Redis is truth
3. **Visibility is Complex**: Three different visibility rules depending on player type and phase
4. **Night Phase is Sequential**: Seer → Werewolf → Witch, each with timer
5. **Circular Dependencies Resolved Elegantly**: Through interface segregation and deferred wiring
6. **Role Assignment is Random**: Fisher-Yates shuffle before game start
7. **Timeouts Matter**: 2-min reconnect, then inactive; 3-min day, 2-min vote, 30-90sec roles

---

**Created**: January 25, 2026
**Version**: 1.0
**Status**: Complete Onboarding
