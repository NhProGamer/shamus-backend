# Shamus Backend - Key Types & Interfaces

## Domain Types

### Game State Types

**GameID** (string alias) - UUID identifier for games

**Game struct**:
```go
type Game struct {
    ID       GameID       // UUID
    Status   GameStatus   // waiting | active | ended
    Phase    GamePhase    // start | day | night | vote
    Day      int          // Day counter (0+)
    Players  []PlayerID   // Player IDs in game
    HostID   PlayerID     // Game creator
    Settings GameSettings // Role configuration
}
```

**GameStatus** enum: `waiting`, `active`, `ended`

**GamePhase** enum: `start`, `day`, `night`, `vote`

**GameSettings**:
```go
type GameSettings struct {
    Roles map[RoleType]int  // e.g., {villager: 4, werewolf: 2, seer: 1}
}
```

### Player Types

**PlayerID** (string alias) - OIDC subject ID

**Player struct**:
```go
type Player struct {
    ID              PlayerID
    Username        string
    Role            Role            // Interface (json:"-")
    RoleType        *RoleType       // For JSON serialization
    IsAlive         bool
    VotedFor        *PlayerID
    ConnectionState ConnectionState
    GameID          *GameID
}
```

**ConnectionState** enum: `connected`, `disconnected`, `inactive`

### Role Types

**RoleType** enum: `villager`, `werewolf`, `seer`, `witch`

**Role interface**:
```go
type Role interface {
    GetType() RoleType
    GetName() string
    GetDescription() string
    GetClans() []Clan
    AddClan(clan Clan)
    GetPriority() Priority  // seer=10, werewolf=9, witch=8
    GetAbilities() *[]Ability
}
```

**Clan** enum: `villager`, `werewolf`, `rogue`, `none`, `lovers`

**Ability interface**:
```go
type Ability interface {
    GetName() string
    GetDescription() string
    CanUse(game *Game, player *Player) bool
    GetConsumptions() *uint8
    TryConsume() bool  // Returns false if no consumptions remaining (atomic check+decrement)
}
```

### Vote Types

**VoteType** enum: `village`, `werewolf`

**VoteStatus** enum: `pending`, `active`, `resolved`

**Vote struct**:
```go
type Vote struct {
    ID              string
    Type            VoteType
    Status          VoteStatus
    EligibleVoters  []PlayerID
    EligibleTargets []PlayerID
    Ballots         map[PlayerID]*PlayerID  // nil = abstain
    AllowAbstain    bool
    Result          *VoteResult
}

// Constructor with validation
func NewVote(id string, voteType VoteType, voters, targets []PlayerID, allowAbstain bool) (*Vote, error)
// Returns ErrNoEligibleVoters if voters is empty
```

**VoteResult struct**:
```go
type VoteResult struct {
    Target      *PlayerID         // nil if tie
    Counts      map[PlayerID]int
    IsTie       bool
    TiedPlayers []PlayerID
}
```

### WebSocket Message Types

**Channel** enum:
```go
const (
    ChannelNotification Channel = "notification"  // Server → Client (info)
    ChannelPrompt       Channel = "prompt"        // Server → Client (action required)
    ChannelResponse     Channel = "response"      // Client → Server (prompt answer)
    ChannelCommand      Channel = "command"       // Client → Server (player action)
)
```

**CommandType** enum:
```go
const (
    CmdSendChat       CommandType = "send_chat"
    CmdUpdateSettings CommandType = "update_settings"
    CmdStartGame      CommandType = "start_game"
    CmdLeaveGame      CommandType = "leave_game"
    CmdKickPlayer     CommandType = "kick_player"
)
```

**Command struct**:
```go
type Command struct {
    Channel Channel         // "command"
    Type    CommandType
    Payload json.RawMessage
}
```

**NotificationType** enum (20+ types):
```go
const (
    // Phase
    NotifPhaseChanged NotificationType = "phase_changed"
    
    // Players
    NotifPlayerJoined   = "player_joined"
    NotifPlayerLeft     = "player_left"
    NotifPlayerDied     = "player_died"
    NotifPlayerInactive = "player_inactive"
    NotifHostChanged    = "host_changed"
    
    // Game state
    NotifGameState   = "game_state"
    NotifGameStarted = "game_started"
    NotifGameEnded   = "game_ended"
    NotifRoleReveal  = "role_reveal"
    
    // Timers
    NotifTimerStarted = "timer_started"
    NotifTimerTick    = "timer_tick"
    NotifTimerExpired = "timer_expired"
    
    // Actions
    NotifSeerResult = "seer_result"
    
    // Votes
    NotifVoteStarted     = "vote_started"
    NotifVoteUpdate      = "vote_update"
    NotifVoteResult      = "vote_result"
    NotifMayorTiebreaker = "mayor_tiebreaker"
    
    // Chat
    NotifChatMessage = "chat_message"
    
    // System
    NotifError = "error"
    NotifAck   = "ack"
)
```

**Notification struct**:
```go
type Notification struct {
    Channel Channel          // "notification"
    Type    NotificationType
    Payload json.RawMessage
}
```

### Prompt Types

**PromptID** (string alias) - UUID

**GroupID** (string alias) - For group votes

**PromptType** enum:
```go
const (
    PromptSelectPlayer PromptType = "select_player"
    PromptSelectOption PromptType = "select_option"
    PromptVote         PromptType = "vote"
    PromptConfirm      PromptType = "confirm"
)
```

**PromptStatus** enum: `pending`, `answered`, `expired`, `cancelled`

**Prompt struct**:
```go
type Prompt struct {
    ID          PromptID
    Type        PromptType
    Context     string          // e.g., "seer_vision", "werewolf_vote"
    GameID      GameID
    PlayerID    PlayerID
    Payload     json.RawMessage
    Status      PromptStatus
    Timeout     time.Duration   // Internal
    ExpiresAt   time.Time
    TimeoutSecs int             // For client display
    AllowChange bool            // Can re-submit (group votes)
    CanSkip     bool
    GroupID     *GroupID
    Response    json.RawMessage // Internal
}
```

**PromptMessage** (sent to client):
```go
type PromptMessage struct {
    Channel     Channel  // "prompt"
    Type        PromptType
    ID          PromptID
    Context     string
    Payload     json.RawMessage
    ExpiresAt   time.Time
    TimeoutSecs int
    AllowChange bool
    CanSkip     bool
    GroupID     *GroupID
}
```

## Port Interfaces

### Repository Ports (`internal/domain/ports/repositories.go`)

**GameRepository**:
```go
type GameRepository interface {
    SaveGame(ctx context.Context, game *entities.Game) error
    GetGame(ctx context.Context, id entities.GameID) (*entities.Game, error)
    DeleteGame(ctx context.Context, id entities.GameID) error
}
```

**PlayerRepository**:
```go
type PlayerRepository interface {
    SavePlayer(ctx context.Context, player *entities.Player) error
    SavePlayers(ctx context.Context, players []*entities.Player) error
    GetPlayer(ctx context.Context, id entities.PlayerID) (*entities.Player, error)
    DeletePlayer(ctx context.Context, id entities.PlayerID) error
    GetPlayersByGame(ctx context.Context, gameID entities.GameID) ([]*entities.Player, error)
    AddPlayerToGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) error
    RemovePlayerFromGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) error
    DeleteGamePlayers(ctx context.Context, gameID entities.GameID) error
}
```

**VoteRepository**:
```go
type VoteRepository interface {
    SaveVote(ctx context.Context, gameID entities.GameID, vote *entities.Vote) error
    GetVote(ctx context.Context, gameID entities.GameID) (*entities.Vote, error)
    DeleteVote(ctx context.Context, gameID entities.GameID) error
}
```

### Service Ports (`internal/domain/ports/services.go`)

**GameService**:
```go
type GameService interface {
    CreateNewGame(ctx context.Context, hostID entities.PlayerID) (*entities.Game, error)
    JoinGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) (*entities.Game, error)
    GetGame(ctx context.Context, gameID entities.GameID) (*entities.Game, error)
    UpdateSettings(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID, settings entities.GameSettings) (*entities.Game, error)
    StartGame(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) (*entities.Game, []*entities.Player, error)
}
```

**PlayerService**:
```go
type PlayerService interface {
    HandleConnect(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID, username string) (*entities.Player, bool, error)
    HandleDisconnect(ctx context.Context, gameID entities.GameID, playerID entities.PlayerID) error
    GetPlayer(ctx context.Context, id entities.PlayerID) (*entities.Player, error)
    GetGamePlayers(ctx context.Context, gameID entities.GameID) ([]*entities.Player, error)
    CleanupGamePlayers(ctx context.Context, gameID entities.GameID) error
    IsPlayerConnected(playerID entities.PlayerID) bool
    LeaveGame(ctx context.Context, playerID entities.PlayerID) error
}
```

**PromptService**:
```go
type PromptService interface {
    CreatePrompt(prompt *entities.Prompt) error
    RespondToPrompt(ctx context.Context, promptID entities.PromptID, playerID entities.PlayerID, response json.RawMessage) error
    CreateGroupVote(gameID entities.GameID, context string, voters []entities.PlayerID, targets []entities.PlayerID, timeout time.Duration, canAbstain bool) (entities.GroupID, error)
    GetGroupState(groupID entities.GroupID) (*GroupVoteState, bool)
    CancelPrompt(promptID entities.PromptID) error
    CancelGroupVote(groupID entities.GroupID) error
}
```

**NotificationService**:
```go
type NotificationService interface {
    NotifyPlayer(playerID entities.PlayerID, notifType entities.NotificationType, payload interface{})
    NotifyAll(gameID entities.GameID, notifType entities.NotificationType, payload interface{})
    NotifyExcept(gameID entities.GameID, excludeID entities.PlayerID, notifType entities.NotificationType, payload interface{})
    NotifyPlayerJoined(gameID entities.GameID, playerID entities.PlayerID, username string)
    NotifyPlayerLeft(gameID entities.GameID, playerID entities.PlayerID, username string, reason string)
    NotifyChatMessage(recipients []entities.PlayerID, senderID entities.PlayerID, username string, message string, channel string, timestamp int64)
    // ... more notification methods
}
```

**GameEngine**:
```go
type GameEngine interface {
    StartGameFlow(ctx context.Context, gameID entities.GameID) error
    HandleSeerAction(ctx context.Context, gameID entities.GameID, seerID entities.PlayerID, targetID entities.PlayerID) error
    HandleWerewolfVote(ctx context.Context, gameID entities.GameID, werewolfID entities.PlayerID, targetID *entities.PlayerID) error
    HandleWitchAction(ctx context.Context, gameID entities.GameID, witchID entities.PlayerID, action string, targetID *entities.PlayerID) error
    HandleVillageVote(ctx context.Context, gameID entities.GameID, voterID entities.PlayerID, targetID *entities.PlayerID) error
}
```

### Broadcasting Ports

**PlayerSender**:
```go
type PlayerSender interface {
    SendToPlayer(playerID entities.PlayerID, payload []byte) error
}
```

**Broadcaster**:
```go
type Broadcaster interface {
    BroadcastToGame(gameID entities.GameID, payload []byte) error
}
```

**ConnectionChecker**:
```go
type ConnectionChecker interface {
    IsPlayerConnected(playerID entities.PlayerID) bool
}
```

**PlayerDisconnecter**:
```go
type PlayerDisconnecter interface {
    DisconnectPlayer(gameID entities.GameID, playerID entities.PlayerID, reason string)
}
```

## Error Codes (`internal/domain/errors/`)

**Game Errors**:
- `ErrGameNotFound`, `ErrGameAlreadyExists`, `ErrGameEnded`, `ErrGameFull`
- `ErrGameNotStarted`, `ErrGameNotWaiting`, `ErrGameNotActive`

**Player Errors**:
- `ErrPlayerNotFound`, `ErrPlayerAlreadyExists`, `ErrPlayerNotInGame`
- `ErrPlayerAlreadyInGame`, `ErrPlayerAlreadyConnected`, `ErrPlayerInactive`

**Vote Errors**:
- `ErrVoteNotFound`, `ErrVoteAlreadyExists`, `ErrInvalidVoter`
- `ErrVoteNotActive`, `ErrNoEligibleVoters`

**Action Errors**:
- `ErrNotYourTurn`, `ErrWrongPhase`, `ErrAlreadyActed`
- `ErrPlayerDead`, `ErrWrongRole`

**Target Errors**:
- `ErrInvalidTarget`, `ErrTargetDead`, `ErrCannotTargetSelf`

**Ability Errors**:
- `ErrAbilityUsed`, `ErrCanOnlyHealVictim`, `ErrTargetAlreadyDying`

**Prompt Errors**:
- `ErrPromptNotFound`, `ErrPromptWrongPlayer`, `ErrPromptExpired`, `ErrPromptAlreadyAnswered`

**Settings Errors**:
- `ErrNotHost`, `ErrInvalidRole`, `ErrTooManyRoles`, `ErrRoleLimitExceeded`
- `ErrNotEnoughPlayers`, `ErrTooManyPlayers`, `ErrRoleCountMismatch`

**WebSocket Command Errors** (in `command_handler.go`):
- `ErrUnknownCommand`, `ErrInvalidPayload`
- `ErrChatMessageEmpty`, `ErrChatMessageTooLong`
- `ErrInvalidChatChannel`, `ErrCannotSendToChannel`
- `ErrNotHost`, `ErrCannotKickSelf`, `ErrGameNotWaiting`
- `ErrPlayerNotInGame`

## Constants (`internal/domain/constants/`)

**Game Configuration**:
- `MinPlayers`: 4
- `MaxPlayers`: 24
- `RoleLimits`: Seer max 1, Witch max 1

**Timeouts**:
- `ReconnectionTimeout`: 2 minutes
- `ServerReadTimeout`: 15 seconds
- `ServerWriteTimeout`: 15 seconds

**Phase Durations**:
- `DayPhaseDuration`: 3 minutes
- `VotePhaseDuration`: 2 minutes
- `SeerActionDuration`: 30 seconds
- `WerewolfVoteDuration`: 1 minute
- `WitchActionDuration`: 45 seconds

**Chat Validation**:
- `MinChatMessageLength`: 1
- `MaxChatMessageLength`: 500

**Redis TTLs**:
- `GameTTL`: 24 hours
- `PlayerTTL`: 24 hours
- `VoteTTL`: 24 hours

## Implementation Classes

### Services (`internal/application/services/`)

| Service | File | Key Dependencies |
|---------|------|------------------|
| GameService | `game_service.go` | GameRepo, PlayerRepo |
| PlayerService | `player_service.go` | PlayerRepo, GameRepo, ConnectionChecker |
| PromptService | `prompt_service.go` | PlayerSender, NotificationService, PlayerRepo |
| NotificationService | `notification_service.go` | PlayerSender, Broadcaster, SessionManager |
| VoteService | `vote_service.go` | VoteRepo, Broadcaster, PlayerSender |
| NightService | `night_service.go` | Broadcaster, VoteService, PlayerRepo |
| TimerService | `timer_service.go` | Broadcaster |
| VisibilityService | `visibility_service.go` | - |
| ChatService | `chat_service.go` | - |

### Adapters

| Adapter | File | Implements |
|---------|------|------------|
| WebSocket Handler | `primary/websocket/handler.go` | Broadcaster, PlayerSender, ConnectionChecker |
| Command Handler | `primary/websocket/command_handler.go` | - |
| Session Manager | `primary/websocket/session_manager.go` | - |
| Redis Game Repo | `secondary/redis/game_repository.go` | GameRepository |
| Redis Player Repo | `secondary/redis/player_repository.go` | PlayerRepository |
| Redis Vote Repo | `secondary/redis/vote_repository.go` | VoteRepository |

## Helpers (`internal/domain/helpers/`)

**Game Helpers** (`game_helpers.go`):
- `AssignRoles()`: Fisher-Yates shuffle for role assignment
- `GetClansFromRoles()`: Extract clans from role list
- `HasRequiredClans()`: Validate game composition

**Sanitize Helpers** (`sanitize.go`):
- `SanitizeChatMessage()`: Remove control characters, preserve printable Unicode
