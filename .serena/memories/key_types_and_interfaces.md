# Shamus Backend - Key Types & Interfaces

## Critical Domain Types

### Game State Types

**GameID** (string alias)
- Unique identifier for a game instance
- Assigned as UUID on creation
- Used as Redis key: "game:{gameID}"

**Game struct**
```
ID       GameID
Status   GameStatus (waiting|active|ended)
Phase    GamePhase (start|day|night|vote)
Day      int (0+ incrementing through cycles)
Players  []PlayerID (list of player IDs in game)
HostID   PlayerID (creator/controller)
Settings GameSettings {Roles map[RoleType]int}
```

**GameStatus** enum:
- waiting: Game created, waiting for players
- active: Game in progress
- ended: Game finished

**GamePhase** enum:
- start: Initial phase before first night
- night: Night action phase
- day: Day discussion phase
- vote: Village voting phase

**GameSettings**
- Roles: Map from RoleType to count (e.g., 4 villagers, 2 werewolves)
- Methods: TotalRoles(), ValidateSettings(), CanStart()

### Player Types

**PlayerID** (string alias)
- Unique identifier for a player (UUID)
- Used as Redis key: "player:{playerID}"
- Associated with WebSocket sessions

**Player struct**
```
ID              PlayerID
Username        string
Role            Role (interface, not serialized)
RoleType        *RoleType (for JSON serialization)
IsAlive         bool
VotedFor        *PlayerID (nil if no vote)
ConnectionState ConnectionState
GameID          *GameID (nil if not in game)
```

**ConnectionState** enum:
- connected: Active WebSocket session
- disconnected: No session (temporary)
- inactive: Marked inactive after 2-min timeout

Methods:
- AssignRole(r Role)
- Kill(), Revive()
- Vote(target *PlayerID), ClearVote()
- Connect(), Disconnect(), SetInactive()

### Role Types

**RoleType** enum:
- seer: Sees one player's role each night
- villager: Regular player (no night action)
- werewolf: Kills one player each night
- witch: Can heal one and poison one per game

**Role interface**
```
GetType() RoleType
GetName() string (localized name)
GetDescription() string
GetClans() []Clan
AddClan(clan Clan)
GetPriority() Priority (seer=10, werewolf=9, witch=8)
GetAbilities() *[]Ability
```

**Clan** enum:
- villager: Seer, Villager, Witch (win condition)
- werewolf: Werewolf (separate win condition)
- rogue: Others (future roles)
- none: Draw condition
- lovers: Added by Cupid (future)

**Ability interface** (not fully detailed in current code):
- Seer ability: Vision
- Werewolf ability: Kill
- Witch abilities: Heal, Poison

### Voting Types

**VoteType** enum:
- village: Day vote by all alive players
- werewolf: Night vote by werewolves

**VoteStatus** enum:
- pending: Vote created but not active
- active: Voting in progress
- resolved: Vote concluded

**Vote struct**
```
ID              string (UUID)
Type            VoteType
Status          VoteStatus
EligibleVoters  []PlayerID
EligibleTargets []PlayerID
Ballots         map[PlayerID]*PlayerID (nil = abstain)
AllowAbstain    bool
Result          *VoteResult
```

Methods:
- NewVote(id, type, voters, targets, allowAbstain)
- CastBallot(voterID, targetID *PlayerID) bool
- HasEveryoneVoted() bool
- Resolve() *VoteResult

**VoteResult struct**
```
Target      *PlayerID (nil if tie)
Counts      map[PlayerID]int
IsTie       bool
TiedPlayers []PlayerID
```

### Night Phase Types

**NightPhase** enum:
- seer: Seer action phase
- werewolf: Werewolf voting phase
- witch: Witch action phase
- end: Night complete, transitioning to day

**NightPhaseOrder** [3]NightPhase:
- Seer (priority 10)
- Werewolf (priority 9)
- Witch (priority 8)

### Event Types

**EventChannel** enum:
- game_event: Game state changes
- conn_event: Connection/disconnection
- settings_event: Game settings updates
- timer_event: Timer notifications

**Event[T]** generic struct
```
Channel EventChannel
Type    EventType (string)
Data    T
```

**RawEvent** = Event[json.RawMessage]

Key event types in events package:
- StartGame, GameData
- Night, NightPhase
- Day
- Vote, VoteResolved
- Death
- Win
- ChatMessage
- Timer
- Connection, Disconnection, Reconnection, Inactive
- RoleReveal
- HostChange
- GameAction
- Error, Ack

## Port Interfaces

### Repository Ports

**GameRepository interface**
```
SaveGame(ctx context.Context, game *Game) error
GetGame(ctx context.Context, id GameID) (*Game, error)
DeleteGame(ctx context.Context, id GameID) error
```

**PlayerRepository interface**
```
SavePlayer(ctx context.Context, player *Player) error
SavePlayers(ctx context.Context, players []*Player) error
GetPlayer(ctx context.Context, id PlayerID) (*Player, error)
DeletePlayer(ctx context.Context, id PlayerID) error
GetPlayersByGame(ctx context.Context, gameID GameID) ([]*Player, error)
AddPlayerToGame(ctx context.Context, gameID GameID, playerID PlayerID) error
RemovePlayerFromGame(ctx context.Context, gameID GameID, playerID PlayerID) error
DeleteGamePlayers(ctx context.Context, gameID GameID) error
```

### Service Ports

**GameService interface**
```
CreateNewGame(hostID PlayerID) (*Game, error)
JoinGame(gameID GameID, playerID PlayerID) (*Game, error)
GetGame(gameID GameID) (*Game, error)
UpdateSettings(gameID GameID, playerID PlayerID, settings GameSettings) (*Game, error)
StartGame(gameID GameID, playerID PlayerID) (*Game, []*Player, error)
```

**PlayerService interface**
```
HandleConnect(gameID GameID, playerID PlayerID, username string) (*Player, bool, error)
HandleDisconnect(gameID GameID, playerID PlayerID) error
GetPlayer(id PlayerID) (*Player, error)
GetGamePlayers(gameID GameID) ([]*Player, error)
CleanupGamePlayers(gameID GameID) error
IsPlayerConnected(playerID PlayerID) bool
```

**VisibilityService interface**
```
BuildPlayersDetailsForPlayer(
    viewer *Player,
    allPlayers []*Player,
    gamePhase GamePhase,
) []events.PlayersDetailsData
```

**ChatService interface**
```
CanSendToChannel(
    sender *Player,
    channel events.ChatChannel,
    gamePhase GamePhase,
) bool

GetChannelRecipients(
    channel events.ChatChannel,
    allPlayers []*Player,
    gamePhase GamePhase,
) []*Player
```

**TimerService interface**
```
StartPhaseTimer(gameID GameID, phase GamePhase)
StartRoleTimer(gameID GameID, roleType RoleType)
CancelTimer(gameID GameID)
SkipTimer(gameID GameID)
GetRemainingTime(gameID GameID) time.Duration
```

**VoteService interface**
```
StartVillageVote(gameID GameID, alivePlayers []*Player) (*Vote, error)
StartWerewolfVote(gameID GameID, werewolves []*Player, potentialVictims []*Player) (*Vote, error)
CastVote(gameID GameID, voterID PlayerID, targetID *PlayerID) error
HasEveryoneVoted(gameID GameID) bool
ResolveVote(gameID GameID) (*VoteResult, error)
GetVote(gameID GameID) (*Vote, bool)
ClearVote(gameID GameID)
```

**NightService interface**
```
StartNight(gameID GameID, players []*Player)
GetCurrentPhase(gameID GameID) string
RecordSeerAction(gameID GameID, targetID PlayerID, revealedRole RoleType)
RecordWerewolfVictim(gameID GameID, victimID *PlayerID)
RecordWitchAction(gameID GameID, healTargetID, poisonTargetID *PlayerID)
GetPendingDeaths(gameID GameID) []PlayerID
IsNightComplete(gameID GameID) bool
ClearNight(gameID GameID)
```

**GameEngine interface**
```
StartGameFlow(gameID GameID) error
HandleSeerAction(gameID GameID, seerID PlayerID, targetID PlayerID) error
HandleWerewolfVote(gameID GameID, werewolfID PlayerID, targetID *PlayerID) error
HandleWitchAction(gameID GameID, witchID PlayerID, healTargetID, poisonTargetID *PlayerID) error
HandleVillageVote(gameID GameID, voterID PlayerID, targetID *PlayerID) error
```

### Broadcasting Ports

**Broadcaster interface**
```
BroadcastToGame(gameID GameID, payload []byte) error
```

**PlayerSender interface**
```
SendToPlayer(playerID PlayerID, payload []byte) error
```

**GameMessenger interface** (composition)
```
Broadcaster
PlayerSender
```

## Implementation Classes

### App Layer

**GameService** (implements GameService interface)
```
Fields:
  gameRepo   GameRepository
  playerRepo PlayerRepository
```

**PlayerService** (implements PlayerService interface)
```
Fields:
  playerRepo      PlayerRepository
  gameRepo        GameRepository
  connChecker     ConnectionChecker (WebSocketHandler)
  reconnTimers    map[PlayerID]*time.Timer
  timerLock       sync.Mutex
```

**GameEngine** (implements GameEngine interface)
```
Fields:
  gameRepo       GameRepository
  playerRepo     PlayerRepository
  timerService   *TimerService
  voteService    *VoteService
  nightService   *NightService
  broadcaster    Broadcaster
  playerSender   PlayerSender
```

**VisibilityService, ChatService, TimerService, VoteService, NightService**
- Specialized service implementations with focused responsibilities

### Infrastructure Layer

**RedisGameRepo** (implements GameRepository)
```
Fields:
  rdb *redis.Client
```

**RedisPlayerRepo** (implements PlayerRepository)
```
Fields:
  rdb *redis.Client
```

**VoteRepository** (in-memory, concurrent-safe)
```
Stores active votes during game
Keyed by game ID
```

**WebSocketHandler** (implements Broadcaster, PlayerSender, ConnectionChecker)
```
Fields:
  melody            *melody.Melody
  gameService       GameService
  playerService     PlayerService
  visibilityService VisibilityService
  chatService       ChatService
  gameEngine        GameEngine
  rooms             map[GameID][]*melody.Session
  playerSessions    map[PlayerID]*melody.Session
  lock              sync.RWMutex
```

## Error Code Enumeration

Structured error codes for client handling:
- GAME_NOT_FOUND, GAME_ALREADY_EXISTS, GAME_ENDED, GAME_FULL
- PLAYER_NOT_FOUND, PLAYER_ALREADY_EXISTS, PLAYER_NOT_IN_GAME
- UNAUTHORIZED, INVALID_TOKEN, MISSING_USER_ID
- INVALID_INPUT, MISSING_FIELD
- NOT_HOST, INVALID_ROLE, TOO_MANY_ROLES, ROLE_LIMIT_EXCEEDED
- NOT_ENOUGH_PLAYERS, TOO_MANY_PLAYERS, ROLE_COUNT_MISMATCH
- NOT_YOUR_TURN, WRONG_PHASE, ALREADY_ACTED, GAME_NOT_ACTIVE
- PLAYER_DEAD, WRONG_ROLE
- INVALID_TARGET, TARGET_DEAD, CANNOT_TARGET_SELF
- ABILITY_USED, CAN_ONLY_HEAL_VICTIM, TARGET_ALREADY_DYING
- VOTE_NOT_FOUND, VOTE_ALREADY_EXISTS, INVALID_VOTER, VOTE_NOT_ACTIVE
- ErrorCodeUnknown, ErrorCodeWrongPhase, ErrorCodeNotYourTurn, etc.

## Constants & Configuration

**Game Configuration**
- MinPlayers: 4
- MaxPlayers: 24
- RoleLimits: Seer max 1, Witch max 1

**Timeouts**
- ReconnectionTimeout: 2 minutes
- ServerReadTimeout: 15 seconds
- ServerWriteTimeout: 15 seconds

**Chat Validation**
- MaxChatMessageLength: 500
- MinChatMessageLength: 1

**Phase Durations**
- DayPhaseDuration: 3 minutes
- VotePhaseDuration: 2 minutes
- SeerActionDuration: 30 seconds
- WerewolfVoteDuration: 1 minute
- WitchActionDuration: 45 seconds

**Redis TTLs**
- GameTTL: 24 hours
- PlayerTTL: 24 hours
- VoteTTL: 24 hours
- NightStateTTL: 24 hours
