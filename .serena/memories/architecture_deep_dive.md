# Shamus Backend - Architecture Deep Dive

## Project Overview

**Shamus Backend** is a Werewolf game backend built in Go with a clean Hexagonal (Ports & Adapters) architecture. The project implements a real-time multiplayer game engine with WebSocket support, role-based visibility, voting mechanics, and night phase actions.

## Directory Structure

```
shamus-backend/
├── cmd/server/main.go                    # Entry point, DI wiring
├── internal/
│   ├── domain/                           # Core business logic (NO external deps)
│   │   ├── entities/                     # Game, Player, Role, Vote, Prompt, Command, Notification
│   │   │   ├── commands/                 # Command payloads (SendChat, KickPlayer, etc.)
│   │   │   ├── prompts/                  # Prompt payloads and responses
│   │   │   ├── roles/                    # Role implementations (Seer, Werewolf, etc.)
│   │   │   └── abilities/                # Ability implementations (Heal, Poison, etc.)
│   │   ├── ports/                        # Interface contracts (services, repositories)
│   │   ├── errors/                       # Structured AppError types
│   │   ├── helpers/                      # Game logic helpers, sanitization
│   │   └── constants/                    # Game constants (timers, limits)
│   ├── application/                      # Use cases and orchestration
│   │   ├── services/                     # Service implementations
│   │   │   ├── game_service.go
│   │   │   ├── player_service.go
│   │   │   ├── vote_service.go
│   │   │   ├── night_service.go
│   │   │   ├── prompt_service.go         # Interactive prompts with timeouts
│   │   │   ├── notification_service.go   # Server-to-client notifications
│   │   │   ├── timer_service.go
│   │   │   ├── visibility_service.go
│   │   │   └── chat_service.go
│   │   └── orchestration/                # Game flow orchestrator
│   │       └── game_engine_v2.go
│   ├── adapters/                         # Interface implementations
│   │   ├── primary/                      # Driving adapters (incoming)
│   │   │   ├── http/
│   │   │   │   ├── controllers/          # HTTP handlers
│   │   │   │   ├── routes/               # URL routing
│   │   │   │   └── middlewares/          # OIDC auth middleware
│   │   │   └── websocket/
│   │   │       ├── handler.go            # WebSocket lifecycle
│   │   │       ├── command_handler.go    # Client command processing
│   │   │       ├── session_manager.go    # Player-session mapping
│   │   │       └── session_helpers.go    # Safe type assertions
│   │   └── secondary/                    # Driven adapters (outgoing)
│   │       └── redis/
│   │           ├── game_repository.go
│   │           ├── player_repository.go
│   │           └── vote_repository.go
│   └── infrastructure/
│       └── config/                       # YAML config loading
├── docs/
│   └── api/                              # API documentation
│       ├── openapi.yaml                  # REST API spec (OpenAPI 3.1)
│       ├── asyncapi.yaml                 # WebSocket API spec (AsyncAPI 2.6)
│       ├── swagger-ui.html               # Swagger UI viewer
│       └── asyncapi-ui.html              # AsyncAPI viewer
└── pkg/
    ├── logger/                           # Zerolog wrapper
    └── utils/                            # Shared utilities
```

## Architecture Layers

### 1. Domain Layer (`internal/domain/`)

The core business logic - completely independent of frameworks and infrastructure.

#### Entities (`entities/`)

**Game** - Game state container:
```go
type Game struct {
    ID       GameID       // UUID
    Status   GameStatus   // waiting | active | ended
    Phase    GamePhase    // start | day | night | vote
    Day      int          // Current day number
    Players  []PlayerID   // Player IDs in game
    HostID   PlayerID     // Game creator
    Settings GameSettings // Role configuration
}
```

**Player** - Individual player state:
```go
type Player struct {
    ID              PlayerID
    Username        string
    Role            Role            // Interface (not serialized)
    RoleType        *RoleType       // For JSON serialization
    IsAlive         bool
    VotedFor        *PlayerID
    ConnectionState ConnectionState // connected | disconnected | inactive
    GameID          *GameID
}
```

**Vote** - Voting system:
```go
type Vote struct {
    ID              string
    Type            VoteType        // village | werewolf
    Status          VoteStatus      // pending | active | resolved
    EligibleVoters  []PlayerID
    EligibleTargets []PlayerID
    Ballots         map[PlayerID]*PlayerID
    AllowAbstain    bool
    Result          *VoteResult
}

// Constructor validates voters
func NewVote(...) (*Vote, error)  // Returns error if no eligible voters
```

**Prompt** - Interactive requests with timeout:
```go
type Prompt struct {
    ID          PromptID
    Type        PromptType      // select_player | select_option | vote | confirm
    Context     string          // e.g., "seer_vision", "werewolf_vote"
    GameID      GameID
    PlayerID    PlayerID
    Payload     json.RawMessage
    Status      PromptStatus    // pending | answered | expired | cancelled
    ExpiresAt   time.Time
    AllowChange bool            // Can re-submit response
    CanSkip     bool
    GroupID     *GroupID        // For group votes
}
```

**Command** - Client-to-server commands:
```go
type Command struct {
    Channel Channel         // "command"
    Type    CommandType     // send_chat | update_settings | start_game | leave_game | kick_player
    Payload json.RawMessage
}
```

**Notification** - Server-to-client notifications:
```go
type Notification struct {
    Channel Channel          // "notification"
    Type    NotificationType // 20+ types: game_state, player_joined, chat_message, etc.
    Payload json.RawMessage
}
```

**Channel** - WebSocket message channels:
```go
const (
    ChannelNotification Channel = "notification"  // Server → Client (info)
    ChannelPrompt       Channel = "prompt"        // Server → Client (action required)
    ChannelResponse     Channel = "response"      // Client → Server (prompt answer)
    ChannelCommand      Channel = "command"       // Client → Server (player action)
)
```

**Role Interface**:
```go
type Role interface {
    GetType() RoleType
    GetName() string
    GetDescription() string
    GetClans() []Clan
    GetPriority() Priority
    GetAbilities() *[]Ability
}
```

**Ability Interface**:
```go
type Ability interface {
    GetName() string
    GetDescription() string
    CanUse(game *Game, player *Player) bool
    GetConsumptions() *uint8
    TryConsume() bool  // Returns false if no consumptions remaining
}
```

#### Ports (`ports/`)

Interface contracts between domain and adapters:

**Repository Ports**:
- `GameRepository`: SaveGame, GetGame, DeleteGame
- `PlayerRepository`: SavePlayer, GetPlayer, GetPlayersByGame, etc.
- `VoteRepository`: SaveVote, GetVote, DeleteVote

**Service Ports**:
- `GameService`: CreateNewGame, JoinGame, GetGame, UpdateSettings, StartGame
- `PlayerService`: HandleConnect, HandleDisconnect, GetPlayer, GetGamePlayers
- `GameEngine`: StartGameFlow, HandleSeerAction, HandleWerewolfVote, HandleWitchAction, HandleVillageVote
- `VoteService`: StartVillageVote, StartWerewolfVote, CastVote, ResolveVote
- `NightService`: StartNight, RecordSeerAction, RecordWerewolfVictim, RecordWitchAction
- `PromptService`: CreatePrompt, RespondToPrompt, CreateGroupVote
- `NotificationService`: NotifyPlayer, NotifyAll, NotifyPlayerJoined, NotifyChatMessage, etc.
- `TimerService`: StartPhaseTimer, CancelTimer, GetRemainingTime
- `VisibilityService`: BuildPlayersDetailsForPlayer
- `ChatService`: CanSendToChannel, GetChannelRecipients

**Broadcasting Ports**:
- `PlayerSender`: SendToPlayer(playerID, payload)
- `Broadcaster`: BroadcastToGame(gameID, payload)
- `NotificationService`: High-level notification methods
- `PlayerDisconnecter`: DisconnectPlayer (for kicks)

### 2. Application Layer (`internal/application/`)

Implements business logic using domain entities and ports.

#### Services (`services/`)

**GameService**: Game lifecycle management
- CreateNewGame: Default config (4V, 2W, 1S, 1W)
- JoinGame: Add player to waiting game
- UpdateSettings: Host-only, validates role configuration
- StartGame: Validates and assigns roles

**PlayerService**: Player connection management
- HandleConnect: Join game or reconnect (2-min timeout)
- HandleDisconnect: Mark disconnected, start timeout timer
- Tracks connection state

**PromptService**: Interactive prompts with timeouts
- CreatePrompt: Send prompt to player
- CreateGroupVote: Werewolf/village group votes
- RespondToPrompt: Process player response
- Manages timers and expiry

**NotificationService**: Server-to-client notifications
- High-level methods: NotifyPlayerJoined, NotifyChatMessage, NotifyVoteResult, etc.
- Uses PlayerSender and Broadcaster interfaces

**VoteService**: Voting mechanics
- StartVillageVote, StartWerewolfVote
- CastVote, ResolveVote
- Tie detection

**NightService**: Night phase coordination
- Sequential sub-phases: Seer → Werewolf → Witch
- Tracks actions and pending deaths

#### Orchestration (`orchestration/`)

**GameEngineV2**: Game flow orchestrator
- StartGameFlow: Initiates first night
- Phase transitions
- Win condition checks
- Timer expiry callbacks
- Coordinates all services

### 3. Adapters Layer (`internal/adapters/`)

#### Primary Adapters (Driving)

**HTTP** (`primary/http/`):
- `controllers/`: REST API handlers (PostGameHandler, HealthHandler)
- `routes/`: URL routing, documentation routes in debug mode
- `middlewares/`: OIDC authentication

**WebSocket** (`primary/websocket/`):
- `handler.go`: WebSocket lifecycle (connect, disconnect, message routing)
- `command_handler.go`: Processes client commands (chat, settings, kick, etc.)
- `session_manager.go`: Player-session mapping, room broadcasts
- `session_helpers.go`: Safe type assertions for session data

**WebSocket Message Flow**:
```
Client → WebSocket Message
    ↓
Handler.onMessage()
    ├─ Parse channel from JSON
    ├─ If "command" → CommandHandler.Handle()
    │   ├─ Validate player belongs to game
    │   ├─ Route by command type
    │   └─ Return ack/error
    └─ If "response" → PromptService.RespondToPrompt()
        └─ Process prompt response

Server → Client
    ├─ NotificationService.NotifyX() → Notification message
    └─ PromptService.CreatePrompt() → Prompt message
```

#### Secondary Adapters (Driven)

**Redis** (`secondary/redis/`):
- `game_repository.go`: Game persistence (24h TTL)
- `player_repository.go`: Player persistence, game player sets
- `vote_repository.go`: Vote persistence

### 4. Infrastructure Layer (`internal/infrastructure/`)

**Config** (`config/`):
- YAML configuration loading
- Server, OIDC, Redis, Logger settings
- `Debug` flag for development features

## Key Design Patterns

1. **Hexagonal Architecture**: Domain isolated from infrastructure via ports
2. **Dependency Injection**: Explicit wiring in main.go
3. **Interface Segregation**: Small interfaces (PlayerSender, Broadcaster)
4. **Command Pattern**: Commands with typed payloads
5. **Observer Pattern**: Notifications broadcast to subscribers
6. **Repository Pattern**: Persistence abstracted behind interfaces
7. **Orchestrator Pattern**: GameEngine coordinates services

## Circular Dependency Resolution

```go
// 1. Create WebSocketHandler without circular deps
wsHandler := NewHandler(melody, sessions, promptService, commandHandler, notifier, gameService)

// 2. Create PlayerService
playerService := NewPlayerService(playerRepo, gameRepo)

// 3. Complete wiring
wsHandler.SetPlayerService(playerService)
commandHandler.SetPlayerService(playerService)
commandHandler.SetGameEngine(gameEngine)
commandHandler.SetDisconnecter(wsHandler)
```

## Component Integration

```
HTTP Request → Controller → Service → Repository → Redis
                              ↓
WebSocket ← NotificationService ← GameEngine
    ↓
CommandHandler → Service → Repository
    ↓
PromptService → Player (prompt) → Response → Service
```

## API Documentation

In debug mode (`config.Debug: true`):
- `/docs/rest` - Swagger UI (OpenAPI 3.1)
- `/docs/ws` - AsyncAPI UI (AsyncAPI 2.6)
- `/docs/api/*` - Raw spec files
