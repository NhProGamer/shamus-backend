# Shamus Backend - Architecture Deep Dive

## Project Overview

**Shamus Backend** is a Werewolf game backend built in Go with a clean, layered architecture using Hexagonal (Ports & Adapters) pattern. The project implements a real-time multiplayer game engine with WebSocket support, role-based visibility, voting mechanics, and night phase actions.

## Architecture Layers

### 1. **Domain Layer** (`internal/domain/`)
The core business logic and rules of the game - completely independent of frameworks and infrastructure.

#### Key Components:

**Entities** (`entities/`):
- `Game`: Represents a game instance with ID, status (waiting/active/ended), phase, day counter, players list, host ID, and role settings
- `Player`: Individual player with ID, username, role assignment, alive status, vote, and connection state
- `Role`: Interface implemented by RoleType (Seer, Villager, Werewolf, Witch) with abilities
- `Vote`: Voting system with eligible voters, targets, ballots, and resolution logic
- `Event`: Generic event structure for WebSocket messages (channels: game_event, conn_event, settings_event, timer_event)
- `NightPhase`: Enum tracking night sub-phases (seer -> werewolf -> witch -> end)

**Roles** (`entities/roles/`):
- Werewolf, Villager, Seer, Witch implementations
- Each role has clan affiliation, abilities, priority, and name/description

**Abilities** (`entities/abilities/`):
- Seer: Vision ability to see target's role
- Werewolf: Kill ability
- Witch: Heal/Poison abilities (only once each per game)

**Events** (`entities/events/`):
- Structured domain events for: Day start, Night start, Votes, Deaths, Chat messages, Connection/Disconnection, Role reveals, Timers, Game actions, etc.

**Ports** (`ports/`):
- Interfaces defining contracts between domain and adapters
- Repositories: GameRepository, PlayerRepository
- Services: GameService, PlayerService, VisibilityService, ChatService, TimerService, VoteService, NightService, GameEngine
- Broadcasting: Broadcaster, PlayerSender, GameMessenger

**Constants** (`constants/`):
- Game rules: MinPlayers=4, MaxPlayers=24
- Timeouts: Reconnection=2min, Server read/write=15s
- Phase durations: Day=3min, Vote=2min, Seer=30s, Werewolf=1min, Witch=45s
- Redis TTLs: 24 hours for games/players/votes/night state

**Errors** (`errors/`):
- Structured AppError type with code and message
- Comprehensive error definitions for game, player, auth, validation, settings, actions, and voting errors

### 2. **Application Layer** (`internal/adapters/app/`)
Implements business logic/use cases using domain entities and ports.

#### Services:

**GameService**:
- CreateNewGame: Creates new game with default 4V 2W 1S 1W configuration
- JoinGame: Adds player to existing game
- GetGame: Retrieves game state
- UpdateSettings: Only host can update roles (requires waiting status)
- StartGame: Validates game state and transitions to active

**PlayerService**:
- HandleConnect: Manages player connection with reconnection logic (2-min timeout timer)
- HandleDisconnect: Marks player as disconnected
- IsPlayerConnected: Checks active WebSocket session
- CleanupGamePlayers: Removes all players when game ends

**GameEngine** (Orchestrator):
- StartGameFlow: Initiates first night phase
- Handles all action processing: Seer actions, Werewolf votes, Witch actions, Village votes
- Manages phase transitions and win condition checks
- Integrated timer expiry callback

**VisibilityService**:
- Filters player information based on viewer's role and game phase
- Rules: Own role always visible, werewolves see each other, dead roles reveal at day

**ChatService**:
- Validates message permissions per channel/phase/role
- Determines channel recipients

**TimerService**:
- Manages phase and role timers
- Supports timer expiry callbacks for phase transitions

**VoteService**:
- Creates village and werewolf votes
- Records ballots and resolves votes
- Handles tie detection and elimination logic

**NightService**:
- Manages night phase sub-phases (seer -> werewolf -> witch)
- Records role actions and pending deaths
- Determines deaths from werewolf victim + witch actions

### 3. **Infrastructure Adapters** (`internal/adapters/`)

#### API Adapters (`api/ws/`):

**WebSocketHandler**:
- Implements Broadcaster and PlayerSender interfaces
- Manages player-to-session mapping and room broadcasts
- Handles WebSocket lifecycle: connect, disconnect, message routing
- Error code translation
- Event routing to appropriate handlers
- Circular dependency resolution: Set PlayerService and GameEngine after creation

**Event Service** (WebSocket):
- Provides SendToPlayer and BroadcastToGame implementations

#### Infrastructure Adapters (`infra/`):

**RedisGameRepo**:
- Persists games to Redis with 24h TTL
- Key format: "game:{gameID}"

**RedisPlayerRepo**:
- Persists players to Redis with 24h TTL
- Key format: "player:{playerID}"
- Maintains game player sets: "game:{gameID}:player_ids"
- Batch save support via Redis pipeline

**VoteRepository**:
- Stores active votes in memory (concurrent-safe)
- Keyed by game ID

### 4. **Infrastructure Layer** (`internal/infrastructure/`)

**Config**:
- YAML-based configuration for Server, OIDC, Redis, Logger settings
- Validation and URL parsing for OIDC issuer and public URLs

**Controllers**:
- REST endpoints for game creation and static files
- OIDC authentication middleware integration

**Routes**:
- Health check endpoints: /health, /ready, /live
- Protected routes under /app with OIDC middleware
- WebSocket endpoint: /app/ws/{gameID}
- API routes: /app/api/v1/game

**Middlewares**:
- OIDC authentication handler

## Component Integration & Data Flow

### Dependency Injection (main.go)

```
Redis → Repositories (GameRepo, PlayerRepo, VoteRepo)
         ↓
    → GameService (uses repos)
    → PlayerService (uses repos)
    → VisibilityService, ChatService
    ↓
WebSocketHandler (created first without circular deps)
    ↓ SetPlayerService() → completes wiring
    ↓
TimerService, VoteService, NightService (use broadcaster)
    ↓
GameEngine (uses all above)
    ↓ SetGameEngine() → completes wiring
    ↓
HTTP Server with Routes and Controllers
```

### Game Flow

1. **Create Game**: Player creates game → GameService creates with waiting status
2. **Join Game**: Players join before game starts
3. **Update Settings**: Host configures roles (validation ensures balance)
4. **Start Game**: Host starts → Role assignment via shuffle → GameEngine.StartGameFlow
5. **Night Phase**:
   - Seer acts (vision)
   - Werewolves vote (kill selection)
   - Witch acts (heal/poison)
   - Deaths resolved
6. **Day Phase**: All players discuss
7. **Vote Phase**: All alive players vote to eliminate
8. **Win Condition**: Check after each phase

### Broadcasting Pattern

Services emit events through Broadcaster/PlayerSender interfaces (raw bytes):
- Broadcaster: BroadcastToGame(gameID, payload) → WebSocketHandler → rooms
- PlayerSender: SendToPlayer(playerID, payload) → WebSocketHandler → player session
- All services use JSON serialization

### Message Routing (WebSocket)

1. Client sends JSON event with channel and type
2. WebSocketHandler routes based on channel:
   - conn_event → HandleConnect/Disconnect
   - game_event → GameService/GameEngine
   - settings_event → SettingsHandlers
   - timer_event → TimerService
3. Responses sent back to player/room

## Key Design Patterns

1. **Hexagonal Architecture**: Domain isolated from infrastructure via ports/adapters
2. **Dependency Injection**: Explicit wiring in main.go with circular dependency breaking
3. **Interface Segregation**: Small focused interfaces (Broadcaster, PlayerSender)
4. **Error Handling**: Structured AppError with codes for consistent client handling
5. **Event-Driven**: Services emit events to broadcaster, WebSocket handler sends to clients
6. **State Management**: Redis as single source of truth for game/player state
7. **Role Interface**: Polymorphic role behavior with consistent interface

## Critical Integration Points

1. **GameEngine ↔ GameService**: Game creation/state management
2. **GameEngine ↔ VoteService**: Vote creation and resolution
3. **GameEngine ↔ NightService**: Night phase orchestration
4. **GameEngine ↔ TimerService**: Phase transitions on timer expiry
5. **PlayerService ↔ WebSocketHandler**: Connection state tracking
6. **WebSocketHandler ↔ All Services**: Event distribution and response sending
7. **Repositories ↔ Services**: Persistent state storage and retrieval
