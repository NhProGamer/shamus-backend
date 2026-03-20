# SHAMUS BACKEND - START HERE

## Welcome to the Onboarding

This folder contains comprehensive documentation for the Shamus Backend project - a production-grade Werewolf game engine built in Go.

**Project:** Werewolf Game Backend  
**Language:** Go 1.24  
**Architecture:** Hexagonal (Ports & Adapters)  
**Status:** Complete Onboarding  
**Created:** January 25, 2026

## Read These In Order

### 1. **onboarding_summary.md** (START HERE)
Quick overview of the entire system. Read this first to understand:
- What the project does
- High-level architecture
- Key concepts
- External dependencies
- Development recommendations

**Reading Time:** 10-15 minutes

### 2. **architecture_deep_dive.md** (DEEP DIVE)
Detailed explanation of all four architecture layers with component relationships:
- Domain layer entities and business rules
- Application layer services
- Infrastructure adapters
- How components integrate
- Design patterns used

**Reading Time:** 20-30 minutes

### 3. **key_types_and_interfaces.md** (REFERENCE)
Complete catalog of critical types and interfaces:
- All domain types (Game, Player, Role, Vote, Event)
- Port interfaces (8 service interfaces)
- Repository contracts
- Error codes (30+)
- Constants and configuration

**Use this as a reference while reading code**

### 4. **integration_and_data_flow.md** (UNDERSTANDING FLOW)
How the system actually works:
- 7-phase initialization in main.go
- 8 detailed data flow scenarios
- Broadcasting patterns
- Service dependencies
- How data moves through the system

**Essential for understanding "how things work"**

### 5. **testing_and_validation.md** (QUALITY)
Testing patterns and validation strategies:
- Unit test approach
- 7 validation layers
- Error handling
- State machines
- Testing gaps and recommendations

**Reference for understanding testing and improving coverage**

## Quick Navigation

**I want to understand...**
- The overall system → Read **onboarding_summary.md**
- How layers work → Read **architecture_deep_dive.md**
- A specific type or interface → Check **key_types_and_interfaces.md**
- How data flows → Read **integration_and_data_flow.md**
- Testing approach → Check **testing_and_validation.md**
- A specific class → Look in **key_types_and_interfaces.md** for Implementation Classes

## Key Concepts At A Glance

### Architecture
```
Infrastructure (Routes, Controllers, Config)
    ↓
Adapters (Services, Repositories, WebSocket)
    ↓
Domain (Entities, Ports, Rules)
    ↓
External (Redis, Gin, Melody, OIDC)
```

### Game Lifecycle
```
CREATE → CONFIGURE → START → PLAY → END
        (waiting)   (active)  (phases)  (ended)
```

### Phase Cycle
```
NIGHT (Seer→Werewolf→Witch) → DAY (Discussion) → VOTE (Eliminate)
[Repeat until win condition]
```

### Core Service Components
- **GameService**: Create, join, configure games
- **PlayerService**: Player lifecycle, reconnection (2-min timeout)
- **GameEngine**: Orchestrates game flow (orchestrator pattern)
- **VoteService**: Voting mechanics and resolution
- **NightService**: Night phase coordination (sequential sub-phases)
- **TimerService**: Phase and role timers
- **VisibilityService**: Role-based information filtering
- **ChatService**: Channel permissions

### Key Insight: Event-Driven
Services don't call clients directly. They emit events to a Broadcaster interface, which WebSocketHandler implements and sends to WebSocket clients.

## Critical Data Model

### Game
- ID, Status (waiting|active|ended), Phase (start|day|night|vote)
- Players list, HostID, Settings (role counts)
- Day counter

### Player
- ID, Username, Role (interface), IsAlive, VotedFor
- ConnectionState (connected|disconnected|inactive)
- GameID

### Vote
- EligibleVoters, EligibleTargets, Ballots (map)
- Status (active|resolved), Result (with tie detection)

### Role
- Type (seer|villager|werewolf|witch)
- Clan (villager|werewolf)
- Priority (for night phase ordering)
- Abilities (vision, kill, heal, poison)

## Important Files To Know

| Path | Purpose |
|------|---------|
| cmd/server/main.go | Entry point, 7-phase DI |
| internal/domain/entities/ | Game, Player, Role, Vote |
| internal/domain/ports/ | Interface contracts (3 files) |
| internal/adapters/app/ | GameService, PlayerService, GameEngine |
| internal/adapters/api/ws/ | WebSocketHandler (real-time communication) |
| internal/adapters/infra/ | Repositories (Redis) |
| internal/infrastructure/routes/ | HTTP routes and endpoints |

## External Dependencies

| Library | Purpose | Critical |
|---------|---------|----------|
| Gin | HTTP routing | Yes |
| Melody | WebSocket | Yes |
| Redis | State persistence | Yes |
| go-oidc | Authentication | Yes |
| UUID | ID generation | Yes |

## Validation Layers (Front to Back)

1. **Controller**: Extract, validate request
2. **Service**: Business rule enforcement
3. **GameEngine**: Game action validation
4. **Repository**: Persistence layer
5. **Client**: Error code handling

## Error Handling

Structured error system with 30+ error codes:
- Game errors (GAME_NOT_FOUND, GAME_FULL, etc.)
- Player errors (PLAYER_NOT_FOUND, PLAYER_DEAD, etc.)
- Vote errors (VOTE_NOT_FOUND, INVALID_VOTER, etc.)
- Action errors (WRONG_PHASE, NOT_YOUR_TURN, etc.)

See **key_types_and_interfaces.md** for complete list.

## Development Workflow

### To Add a New Feature

**New Game Action:**
1. Add validation to Game/Player entity
2. Implement handler in GameEngine
3. Add routing in WebSocketHandler
4. Emit events via Broadcaster
5. Handle visibility if role-specific

**New Role:**
1. Create Role implementation in entities/roles/
2. Add RoleType enum
3. Create factory method
4. Add to NightPhaseOrder if active at night
5. Update VisibilityService rules

**New Service:**
1. Define Port interface in domain/ports/
2. Implement in adapters/app/
3. Inject in main.go DI phase
4. Wire to WebSocketHandler if needed
5. Emit through Broadcaster

### Commands

```bash
go test ./...              # Run all tests
go run cmd/server/main.go  # Start server
golangci-lint run          # Lint check
```

## Architecture Philosophy

The codebase follows these principles:
- **Domain-Driven**: Business logic isolated in domain layer
- **Hexagonal**: Infrastructure as pluggable adapters
- **Interface Segregation**: Small focused interfaces
- **Dependency Inversion**: Services depend on ports, not implementations
- **Event-Driven**: Services emit, handlers deliver
- **Repository Pattern**: Persistence abstracted behind interface

This allows:
- Easy testing (mock implementations)
- Technology swapping (Redis → different DB)
- Parallel development
- Clear separation of concerns

## Key Insights

1. **Real-Time is Event-Driven**
   Services emit to Broadcaster → WebSocketHandler broadcasts

2. **State is in Redis**
   24-hour TTL, not in memory

3. **Visibility is Role-Based**
   Different rules for different roles and phases

4. **Night Phase is Sequential**
   Seer → Werewolf → Witch (priority order)

5. **Circular Deps Resolved Elegantly**
   Interface segregation + deferred wiring

6. **Role Assignment is Random**
   Fisher-Yates shuffle before start

7. **Timeouts Matter**
   2-min reconnect, 3-min day, 2-min vote, 30-90sec roles

## Next Steps

1. ✓ You are here - reading START_HERE
2. → Read **onboarding_summary.md** (10 min)
3. → Read **architecture_deep_dive.md** (25 min)
4. → Check **key_types_and_interfaces.md** when reading code
5. → Study **integration_and_data_flow.md** (20 min)
6. → Review **testing_and_validation.md** for code quality
7. → Explore the actual code in your IDE
8. → Try adding a simple feature

## Finding Help

**Question About...** → **See This File**
- Project overview → onboarding_summary.md
- Architecture layers → architecture_deep_dive.md
- Types/Interfaces → key_types_and_interfaces.md
- Data flows → integration_and_data_flow.md
- Testing → testing_and_validation.md
- Specific error → key_types_and_interfaces.md (Error Code section)
- Component relationships → architecture_deep_dive.md (Integration Points)

## Statistics

- **Go Files**: 40+
- **Port Interfaces**: 8
- **Error Codes**: 30+
- **Game Phases**: 4
- **Roles**: 4
- **Player Limit**: 4-24
- **Data TTL**: 24 hours
- **Reconnect Window**: 2 minutes

---

**Welcome!** This is a well-architected production system. The layered architecture makes it easy to understand and extend.

Start with onboarding_summary.md → architecture_deep_dive.md → then explore the code!
