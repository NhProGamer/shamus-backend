# Shamus Backend - Testing & Validation

## Testing Patterns

### Unit Testing Approach

Standard Go testing with table-driven tests in `*_test.go` files.

**Example Pattern**:
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
        {
            name: "not enough players",
            game: &entities.Game{Players: []PlayerID{"1", "2"}},
            wantErr: true,
            errMsg: "not enough players",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.game.CanStart()
            if (err != nil) != tt.wantErr {
                t.Errorf("CanStart() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Current Test Coverage

| Package | Coverage | Notes |
|---------|----------|-------|
| `domain/entities` | Partial | Game.CanStart(), Vote.Resolve() |
| `domain/errors` | Good | Error code mapping |
| `application/services` | Limited | Need more coverage |
| `adapters` | Limited | Need integration tests |

## Validation Layers

### Layer 1: WebSocket Entry (Handler)
```
onMessage()
├─ ExtractSessionData() - Safe type assertions
├─ Parse channel from JSON
└─ Route to CommandHandler or PromptService
```

### Layer 2: Command Validation (CommandHandler)
```
Handle()
├─ Validate player belongs to game
├─ Route by command type
└─ Per-command validation:
    ├─ handleSendChat: length, sanitization, channel permissions
    ├─ handleUpdateSettings: host check, game status
    ├─ handleStartGame: host check, game.CanStart()
    ├─ handleKickPlayer: host check, not self
    └─ handleLeaveGame: player exists
```

### Layer 3: Service Validation
```
GameService.StartGame()
├─ game.Status == waiting
├─ playerID == game.HostID
├─ game.CanStart()
│   ├─ PlayerCount >= MinPlayers (4)
│   ├─ PlayerCount <= MaxPlayers (24)
│   ├─ TotalRoles == PlayerCount
│   ├─ Role limits respected
│   └─ HasRequiredClans()
└─ Role assignment
```

### Layer 4: Entity Validation
```
NewVote()
├─ len(voters) > 0 (returns ErrNoEligibleVoters)
└─ Creates Vote struct

Vote.CastBallot()
├─ Voter in EligibleVoters
├─ Target in EligibleTargets (or nil if AllowAbstain)
└─ Stores ballot

Ability.TryConsume()
├─ consumptions != nil
├─ *consumptions > 0
└─ Decrements and returns true
```

### Layer 5: Prompt Validation
```
PromptService.RespondToPrompt()
├─ Prompt exists
├─ Prompt.PlayerID matches
├─ Prompt not expired
├─ Prompt.CanRespond() (pending or allowChange)
└─ Process response
```

### Layer 6: Repository Validation
```
Redis operations
├─ Key exists check
├─ JSON unmarshal validation
└─ Error propagation
```

## Error Handling

### Structured Errors
```go
type AppError struct {
    Code    string  // Machine-readable
    Message string  // Human-readable
    Err     error   // Wrapped error
}
```

### Error Categories

| Category | Examples |
|----------|----------|
| Game | `ErrGameNotFound`, `ErrGameEnded`, `ErrGameFull` |
| Player | `ErrPlayerNotFound`, `ErrPlayerDead`, `ErrPlayerNotInGame` |
| Vote | `ErrVoteNotFound`, `ErrInvalidVoter`, `ErrNoEligibleVoters` |
| Action | `ErrWrongPhase`, `ErrNotYourTurn`, `ErrAlreadyActed` |
| Prompt | `ErrPromptNotFound`, `ErrPromptExpired`, `ErrPromptWrongPlayer` |
| Settings | `ErrNotHost`, `ErrInvalidRole`, `ErrRoleLimitExceeded` |

### WebSocket Error Codes
```go
// Mapped in handler.go
switch err {
case ErrUnknownCommand:     code = "UNKNOWN_COMMAND"
case ErrInvalidPayload:     code = "INVALID_PAYLOAD"
case ErrNotHost:            code = "NOT_HOST"
case ErrPlayerNotInGame:    code = "NOT_IN_GAME"
case ErrChatMessageEmpty:   code = "MESSAGE_EMPTY"
// etc.
}
```

## State Validation

### Game State Machine
```
waiting → active → ended
   │         │
   │         └─ Win condition met
   └─ host calls start_game (valid config)
```

### Player Connection State
```
connected → disconnected → inactive
    │            │              │
    │            │              └─ 2-min timeout
    │            └─ WebSocket close
    └─ Reconnect within 2 min → connected
```

### Vote State
```
pending → active → resolved
            │          │
            │          └─ All voted or timeout
            └─ Vote created
```

### Prompt State
```
pending → answered
    │         │
    │         └─ Player responded
    ├─ expired (timeout)
    └─ cancelled (manually)
```

## Security Validations

### Player-Game Validation
```go
// CommandHandler.Handle()
player, err := h.playerService.GetPlayer(ctx, cmdCtx.PlayerID)
if player.GameID == nil || *player.GameID != cmdCtx.GameID {
    return ErrPlayerNotInGame
}
```

### Chat Sanitization
```go
// helpers.SanitizeChatMessage()
for _, r := range message {
    if unicode.IsPrint(r) || r == '\t' || r == '\n' {
        sb.WriteRune(r)
    }
    // Control characters silently dropped
}
```

### Safe Type Assertions
```go
// session_helpers.go
func ExtractGameID(s *melody.Session) (entities.GameID, error) {
    val, exists := s.Get("gameId")
    if !exists || val == nil {
        return "", ErrMissingGameID
    }
    str, ok := val.(string)  // ok-pattern
    if !ok {
        return "", ErrInvalidGameID
    }
    return entities.GameID(str), nil
}
```

### Atomic Ability Consumption
```go
// TryConsume() instead of Consume()
func (h *HealAbility) TryConsume() bool {
    if h.consumptions == nil {
        return true  // Unlimited
    }
    if *h.consumptions == 0 {
        return false  // Cannot consume
    }
    *h.consumptions--
    return true
}
```

## Testing Gaps & Recommendations

### Current Gaps
- No integration tests (full game flow)
- No WebSocket routing tests
- No concurrent action tests
- No Redis persistence tests
- Limited service orchestration tests

### Recommended Tests

1. **Integration Tests**: Full game from creation to end
2. **WebSocket Tests**: Message routing, error handling
3. **Concurrency Tests**: Multiple players voting simultaneously
4. **Edge Cases**: All dead, single werewolf, ties
5. **Reconnection Tests**: Timeout behavior, state preservation

## Test Commands

```bash
# Run all tests
go test ./...

# Verbose
go test -v ./...

# Specific package
go test ./internal/domain/entities/...

# Coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
