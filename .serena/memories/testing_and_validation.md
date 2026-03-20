# Shamus Backend - Testing & Validation

## Testing Patterns

### Unit Testing Approach

The project uses standard Go testing package with table-driven tests. Located in `*_test.go` files.

**Example: Game State Validation** (`internal/domain/entities/game_test.go`)

Test cases include:
- Valid game configuration (sufficient players, balanced roles)
- Not enough players error
- Too many players error
- Role count mismatch error
- Role limit exceeded (e.g., 2 seers when max is 1)
- Invalid role type error
- Invalid composition (missing villagers or werewolves)

**Test Structure Pattern**:
```go
func TestGame_CanStart(t *testing.T) {
  tests := []struct {
    name    string
    game    *entities.Game
    wantErr bool
    errMsg  string
  }{
    {
      name: "valid game configuration",
      game: &entities.Game{...},
      wantErr: false,
    },
    // ... more test cases
  }
  
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      err := tt.game.CanStart()
      if (err != nil) != tt.wantErr {
        t.Errorf("CanStart() error = %v, wantErr %v", err, tt.wantErr)
      }
      if err != nil && !strings.Contains(err.Error(), tt.errMsg) {
        t.Errorf("error message mismatch")
      }
    })
  }
}
```

### Validation Layers

#### 1. **Game Creation Validation**
```
Game.CanStart() validates:
├─ PlayerCount >= MinPlayers (4) and <= MaxPlayers (24)
├─ TotalRoles == PlayerCount
├─ Role limits (Seer max 1, Witch max 1)
├─ Required clans present (at least 1 villager, 1 werewolf/rogue)
└─ All role types are valid
```

#### 2. **GameSettings Validation**
```
Game.ValidateSettings() checks:
├─ All role types are valid
├─ Role counts non-negative
├─ Role-specific limits (Seer ≤ 1, Witch ≤ 1)
└─ TotalRoles <= MaxPlayers
```

#### 3. **Chat Message Validation**
```
Constants:
├─ MinChatMessageLength: 1 character (> 0)
└─ MaxChatMessageLength: 500 characters

Validation rules:
├─ len(message) >= MinChatMessageLength
├─ len(message) <= MaxChatMessageLength
└─ Not empty/whitespace
```

#### 4. **WebSocket Event Validation**
Events must contain:
```
├─ Channel: one of (game_event, conn_event, settings_event, timer_event)
├─ Type: string identifying event type
└─ Data: JSON payload for event
```

#### 5. **Player Connection Validation**
```
HandleConnect() validates:
├─ Player not already connected
├─ Game exists
├─ Game status != ended
├─ If new player: game.status == waiting
├─ If reconnecting: within 2-min timeout window
└─ Player has assigned role if game active
```

#### 6. **Game Action Validation**
```
Each game action (vote, seer action, etc.) checks:
├─ Game in correct phase
├─ Player alive (except for specific actions)
├─ Player has correct role
├─ Target is valid (alive, not self, etc.)
├─ Ability not used (for limited abilities)
└─ Vote eligibility (voter in eligible list, target in targets)
```

#### 7. **Role Assignment Validation**
```
helpers.AssignRoles() validates:
├─ Role count matches player count
├─ Shuffle maintains all role types
└─ Each player gets exactly one role
```

## Error Handling & Codes

### Structured Error Format

```go
type AppError struct {
  Code    string  // Machine-readable code
  Message string  // Human-readable message
  Err     error   // Wrapped error
}
```

### Error Code Categories

**Game Errors**:
- GAME_NOT_FOUND: Game ID not found in database
- GAME_ALREADY_EXISTS: Duplicate game creation attempt
- GAME_ENDED: Game has concluded
- GAME_FULL: Maximum player limit reached
- GAME_NOT_STARTED: Game still in waiting phase
- GAME_NOT_WAITING: Cannot join non-waiting game

**Player Errors**:
- PLAYER_NOT_FOUND: Player ID not found
- PLAYER_ALREADY_EXISTS: Duplicate player
- PLAYER_NOT_IN_GAME: Player not part of game
- PLAYER_ALREADY_IN_GAME: Player already joined
- PLAYER_ALREADY_CONNECTED: Active WebSocket session exists
- PLAYER_INACTIVE: Marked inactive after timeout

**Game Action Errors**:
- NOT_YOUR_TURN: Out of turn action
- WRONG_PHASE: Action not valid for current phase
- ALREADY_ACTED: Player already acted this phase
- GAME_NOT_ACTIVE: Game not in active state
- PLAYER_DEAD: Dead players cannot act
- WRONG_ROLE: Action requires specific role

**Target Errors**:
- INVALID_TARGET: Target ID not in eligible targets
- TARGET_DEAD: Target player is dead
- CANNOT_TARGET_SELF: Self-targeting not allowed

**Ability Errors**:
- ABILITY_USED: Limited-use ability already used
- CAN_ONLY_HEAL_VICTIM: Witch can only heal werewolf victim
- TARGET_ALREADY_DYING: Multiple death sources

**Vote Errors**:
- VOTE_NOT_FOUND: No active vote
- VOTE_ALREADY_EXISTS: Vote already exists
- INVALID_VOTER: Voter not eligible
- VOTE_NOT_ACTIVE: Vote not in active state

**Settings Errors**:
- NOT_HOST: Only host can modify
- INVALID_ROLE: Role type not recognized
- TOO_MANY_ROLES: Exceeds MaxPlayers
- ROLE_LIMIT_EXCEEDED: Role-specific limit violated
- NOT_ENOUGH_PLAYERS: Below MinPlayers
- TOO_MANY_PLAYERS: Exceeds MaxPlayers
- ROLE_COUNT_MISMATCH: Roles ≠ players
- INVALID_COMPOSITION: Missing villagers or werewolves

### Error Translation for WebSocket

WebSocketHandler.errorToCode() maps AppError to structured ErrorCode enums:
- ErrorCodeUnknown
- ErrorCodeWrongPhase
- ErrorCodeNotYourTurn
- ErrorCodeAlreadyActed
- ErrorCodeGameNotActive
- ErrorCodePlayerDead
- ErrorCodeWrongRole
- ErrorCodeInvalidTarget
- ErrorCodeTargetDead
- ErrorCodeCannotTargetSelf
- ErrorCodeAbilityUsed
- ErrorCodeCanOnlyHealVictim
- ErrorCodeVoteNotFound
- ErrorCodeVoteNotActive
- ErrorCodeInvalidVoter
- (And more...)

Clients receive JSON error events:
```json
{
  "channel": "game_event",
  "type": "error",
  "data": {
    "code": "WRONG_PHASE",
    "message": "wrong game phase"
  }
}
```

## State Validation Rules

### Vote Validation

**Vote Eligibility**:
- Eligible voters must be in vote.EligibleVoters list
- Targets must be in vote.EligibleTargets list
- Each voter can only vote once (ballot override allowed)
- Abstention only if AllowAbstain = true

**Vote Resolution**:
- Requires all eligible voters to have voted
- Tallies votes and finds maximum vote count
- Detects ties (multiple targets with max votes)
- Eliminates target only if no tie

### Night Phase Validation

**Sub-phase Ordering**:
1. Seer (priority 10): Can see one player's role
2. Werewolf (priority 9): Vote to kill one player
3. Witch (priority 8): Can heal werewolf victim OR poison another
4. End: Process deaths and transition to day

**Witch Ability Constraints**:
- Heal: Only valid target is werewolf victim
- Poison: Can only poison once per game (one-time use)
- Both heal and poison apply same night (logical OR)

**Death Resolution**:
- Werewolf victim dies if not healed
- Poison victim dies if not healed
- No double death (heal cancels both kill and poison)

### Connection State Machine

```
States: connected, disconnected, inactive

Transitions:
├─ connected → disconnected: WebSocket disconnect
│                 ↓ (2 min timeout)
│              inactive: Marked inactive
│
├─ disconnected → connected: Reconnect within 2 min
│
└─ inactive → cannot rejoin active game
```

### Game Status Validation

```
Status: waiting → active → ended

Rules:
├─ Waiting:
│  ├─ Players can join
│  ├─ Host can configure settings
│  ├─ Host can start (if valid)
│  └─ Cannot perform game actions
│
├─ Active:
│  ├─ New players cannot join
│  ├─ Cannot modify settings
│  ├─ Game actions allowed
│  └─ Timers active
│
└─ Ended:
   ├─ No joins allowed
   ├─ No actions allowed
   ├─ Winner determined
   └─ Cleanup eligible
```

## Redis Data Validation

### Key Formats & TTL

**Games**:
- Key: `game:{gameID}`
- TTL: 24 hours
- Value: Serialized Game struct (JSON)

**Players**:
- Key: `player:{playerID}`
- TTL: 24 hours
- Value: Serialized Player struct (JSON)

**Game Player Sets**:
- Key: `game:{gameID}:player_ids`
- Type: Redis set
- Members: Player IDs
- TTL: 24 hours

**Votes**:
- Storage: In-memory map (VoteRepository)
- Key: GameID
- Value: Vote struct
- Note: Not persisted to Redis

**Night State**:
- Storage: In-memory (NightService)
- Key: GameID
- Value: Night phase state
- Includes: Current phase, actions recorded, deaths pending

### JSON Serialization Validation

All Redis values are JSON-serialized:
```go
// Marshal
data, err := JSON.Marshal(entity)
rdb.Set(ctx, key, data, ttl).Err()

// Unmarshal
var entity Entity
JSON.Unmarshal([]byte(val), &entity)
```

**Role Serialization**:
- Role interface NOT serialized (not exported)
- RoleType string serialized as JSON field
- On deserialize: Must reconstruct Role object from RoleType

**Player Serialization**:
- Player.Role (interface): Skipped (tag: `-`)
- Player.RoleType (*RoleType): Serialized
- On load: Reconstruct Role from RoleType using factories.GetNewRole()

## Testing Considerations

### Gaps/Future Testing

The project would benefit from:

1. **Integration Tests**:
   - Full game flow from creation to finish
   - WebSocket message routing
   - Redis interaction
   - OIDC authentication flow

2. **Service Layer Tests**:
   - GameEngine orchestration logic
   - Timer callbacks and phase transitions
   - Vote resolution with tie scenarios
   - Night phase sub-phase progression

3. **Concurrency Tests**:
   - Multiple players voting simultaneously
   - Connection/disconnection during active game
   - Race conditions in shared state

4. **WebSocket Tests**:
   - Event serialization/deserialization
   - Error code translation
   - Broadcast to room
   - Player-specific messaging

5. **Redis Tests**:
   - Persistence and retrieval
   - TTL expiry behavior
   - Pipeline batch operations
   - Error handling (connection failures)

6. **Game Logic Tests**:
   - Win condition detection
   - Complex role interaction scenarios
   - Edge cases (all dead, single werewolf, etc.)

## Validation Entry Points

### From Client

1. **HTTP Request** → Controllers
   - Extract userID (OIDC middleware)
   - Validate request body
   - Call service

2. **WebSocket Message** → WebSocketHandler
   - Parse Event JSON
   - Validate channel and type
   - Route to appropriate handler
   - Validate payload structure

3. **Game Action** → GameEngine
   - Validate game state
   - Validate player state
   - Validate action parameters
   - Apply business rules

### From Repository

1. **GetGame/GetPlayer**
   - Check found (not redis.Nil)
   - Deserialize JSON
   - Reconstruct complex objects

2. **SaveGame/SavePlayer**
   - Serialize to JSON
   - Write to Redis
   - Verify success

## Configuration & Security

### YAML Validation

Config struct parsed with validation:
```yaml
server:
  host: "127.0.0.1"
  port: 8080
  public_url: "https://..."
  cookie_store_key: "..."

oidc:
  issuer: "https://..."
  client_id: "..."
  secret: "..."
  scopes:
    - openid
    - profile
    - email

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

logger:
  level: "info"
  pretty: true

debug: false
```

### OIDC Validation

- Provider initialization with issuer URL validation
- Token verification
- User ID extraction from claims
- Scope validation

### CORS Configuration

- Allowed origins: localhost:3000, localhost:5173
- Allowed methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
- Allowed headers: Origin, Content-Type, Accept, Authorization
- Credentials allowed
- Max age: 12 hours
