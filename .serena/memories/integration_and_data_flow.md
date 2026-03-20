# Shamus Backend - Integration & Data Flow

## Initialization Flow (main.go)

### 1. Configuration Loading
```
Config.LoadConfig() → YAML file
  ├─ Server: host, port, OIDC config
  ├─ OIDC: issuer, client_id, secret, scopes
  ├─ Redis: host, port, password, db
  └─ Logger: level, pretty output
```

### 2. Infrastructure Setup
```
Logger initialization → Zerolog config
  ↓
Gin router + Melody WebSocket manager
  ↓
OIDC provider init (coreos/go-oidc)
  ↓
Redis client → Ping test
```

### 3. Dependency Injection Phase 1: Repositories
```
Redis client
  ├─→ RedisGameRepo (GameRepository)
  ├─→ RedisPlayerRepo (PlayerRepository)
  └─→ VoteRepository (in-memory)
```

### 4. Dependency Injection Phase 2: Services
```
Repositories
  ├─→ GameService(gameRepo, playerRepo)
  ├─→ PlayerService(playerRepo, gameRepo, nil) [connChecker set later]
  ├─→ VisibilityService()
  └─→ ChatService()
```

### 5. Dependency Injection Phase 3: WebSocket & Circular Dep Breaking
```
Create WebSocketHandler without circular deps:
  NewWebSocketHandler(melody, gameService, visibilityService, chatService)
    ├─ playerService = nil [SET LATER]
    └─ gameEngine = nil [SET LATER]
  ↓
Complete wiring:
  wsHandler.SetPlayerService(playerService)
  ↓
  playerService.connChecker now set to wsHandler
```

### 6. Dependency Injection Phase 4: Game Flow Services
```
TimerService(wsHandler) [uses Broadcaster]
VoteService(voteRepo, wsHandler, wsHandler) [uses Broadcaster & PlayerSender]
NightService(wsHandler, voteService) [uses Broadcaster]
  ↓
GameEngine(gameRepo, playerRepo, timerService, voteService, nightService, wsHandler, wsHandler)
  ├─ Sets up timer expiry callbacks
  └─ wsHandler.SetGameEngine(gameEngine)
```

### 7. HTTP Server Setup
```
Routes.InitRoutes(router, AppContext{config, gameService, wsHandler, oidcProvider}, redis)
  ├─ Health endpoints: /health, /ready, /live (no auth)
  ├─ Static files: /app/ (OIDC protected)
  ├─ WebSocket: /app/ws/{gameID} (OIDC protected)
  └─ API: /app/api/v1/game (OIDC protected)
```

## Data Flow: Create Game

```
Client → HTTP POST /app/api/v1/game
  ↓
Controllers.PostGameHandler
  ├─ Extract userID from context (from OIDC middleware)
  └─ gameService.CreateNewGame(playerID)
    ├─ Generate gameID (UUID)
    ├─ Create Game{
    │   ID: gameID,
    │   Status: waiting,
    │   Phase: start,
    │   Players: [],
    │   HostID: playerID,
    │   Settings: {4V, 2W, 1S, 1W}
    │ }
    └─ gameRepo.SaveGame(context, game)
      └─ Redis SET "game:{gameID}" with 24h TTL
  ↓
HTTP 200 {gameID: "..."}
  ↓
Client connects via WebSocket to /app/ws/{gameID}
```

## Data Flow: Join Game

```
Client → WebSocket Connect /app/ws/{gameID}
  ↓
WebSocketHandler.HandleWS (Melody connect event)
  ├─ Extract gameID and playerID from context
  ├─ playerService.HandleConnect(gameID, playerID, username)
  │  ├─ Check not already connected
  │  ├─ Get game from gameRepo
  │  ├─ Check game status != ended
  │  ├─ Check if player exists
  │  │  ├─ If NOT: Create new Player (only if game.status == waiting)
  │  │  │   ├─ playerRepo.SavePlayer()
  │  │  │   ├─ playerRepo.AddPlayerToGame()
  │  │  │   └─ return player, false (not reconnect)
  │  │  └─ If YES: Mark as connected (clear reconnect timer)
  │  │     └─ return player, true (reconnect)
  │  └─ If 2-min disconnected timeout: mark inactive
  │
  ├─ Register player session: wsHandler.playerSessions[playerID] = session
  ├─ Register in room: wsHandler.rooms[gameID] = append(..., session)
  │
  └─ Broadcast connection event to room
    └─ BroadcastToGame(gameID, Event{
         channel: "conn_event",
         type: "connection"
         data: {playerID, username, state: "connected"}
       })
```

## Data Flow: Start Game

```
Client → WebSocket message {
  channel: "game_event",
  type: "start_game",
  data: {gameID, playerID}
}
  ↓
WebSocketHandler.setupEvents() routes to StartGameHandler
  ├─ Validate playerID is host
  ├─ gameService.StartGame(gameID, playerID)
  │  ├─ Get game from gameRepo
  │  ├─ Validate game.CanStart() [players, roles, balance]
  │  ├─ Get all players: playerRepo.GetPlayersByGame()
  │  ├─ Shuffle and assign roles: helpers.AssignRoles(players, settings)
  │  │  └─ Each player gets Role implementation (Werewolf, Seer, etc)
  │  ├─ Save all players: playerRepo.SavePlayers()
  │  ├─ Update game: status=active, phase=night, day=1
  │  ├─ gameRepo.SaveGame()
  │  └─ return game, players
  │
  ├─ gameEngine.StartGameFlow(gameID)
  │  ├─ Start first night phase
  │  ├─ Determine seer player
  │  ├─ Start seer timer
  │  └─ BroadcastToGame(Event{type: "night"})
  │
  └─ Send visibility data to each player
    └─ For each player:
      └─ visibilityService.BuildPlayersDetailsForPlayer(player, allPlayers, phase)
        └─ Returns player list filtered by visibility rules
      └─ SendToPlayer(playerID, Event{type: "game_data", data: {...}})
```

## Data Flow: Seer Vision Action (Night Phase)

```
Client → WebSocket message {
  channel: "game_event",
  type: "seer_action",
  data: {gameID, seerID, targetID}
}
  ↓
WebSocketHandler.HandleSeerAction()
  ├─ Validation:
  │  ├─ gameEngine.HandleSeerAction(gameID, seerID, targetID)
  │  │  ├─ Get game, check phase == night
  │  │  ├─ Get seer and target players
  │  │  ├─ Check seer alive, not used ability, target alive
  │  │  ├─ nightService.RecordSeerAction(gameID, targetID, targetRole)
  │  │  ├─ Check if night complete: nightService.IsNightComplete()
  │  │  │  └─ If yes: ProcessNightEnd(gameID)
  │  │  │    ├─ Get pending deaths from witch actions
  │  │  │    ├─ Kill those players
  │  │  │    ├─ Update game: phase = day
  │  │  │    ├─ BroadcastToGame(Event{type: "day", data: {deaths}})
  │  │  │    ├─ Check win condition
  │  │  │    └─ If no winner: StartPhaseTimer(day) → 3 minutes
  │  │  │
  │  │  ├─ timerService.SkipTimer() if all acted
  │  │  └─ return nil
  │  │
  │  └─ Send ACK to seer
  │
  └─ BroadcastToGame(Event{
       channel: "game_event",
       type: "game_action",
       data: {action: "seer_action", status: "success"}
     })
```

## Data Flow: Village Vote (Day Phase)

```
Client → WebSocket message {
  channel: "game_event",
  type: "village_vote",
  data: {gameID, voterID, targetID}
}
  ↓
WebSocketHandler.HandleVillageVote()
  ├─ gameEngine.HandleVillageVote(gameID, voterID, targetID)
  │  ├─ Get game, check phase == vote
  │  ├─ Get vote: voteService.GetVote(gameID)
  │  ├─ voteService.CastVote(gameID, voterID, targetID)
  │  │  └─ vote.CastBallot(voterID, targetID)
  │  │    ├─ Validate voter eligible
  │  │    ├─ Validate target eligible
  │  │    ├─ Store ballot
  │  │    └─ return success
  │  │
  │  ├─ If voteService.HasEveryoneVoted():
  │  │  └─ ProcessVoteResult(gameID)
  │  │    ├─ voteService.ResolveVote(gameID)
  │  │    │  └─ vote.Resolve() → VoteResult{target, counts, isTie}
  │  │    ├─ If target: kill player
  │  │    ├─ Update game: phase = night, day++
  │  │    ├─ gameRepo.SaveGame()
  │  │    ├─ nightService.StartNight()
  │  │    ├─ Check win condition
  │  │    ├─ BroadcastToGame(Event{type: "vote_resolved"})
  │  │    └─ StartPhaseTimer(night)
  │  │
  │  └─ return nil
  │
  └─ BroadcastToGame(Event{
       channel: "game_event",
       type: "vote",
       data: {voterID, targetID}
     })
```

## Data Flow: Player Disconnect

```
Client → WebSocket Disconnect
  ↓
WebSocketHandler.HandleDisconnect (Melody event)
  ├─ playerService.HandleDisconnect(gameID, playerID)
  │  ├─ Get player
  │  ├─ Mark player.ConnectionState = disconnected
  │  ├─ playerRepo.SavePlayer()
  │  ├─ Start reconnection timer (2 minutes)
  │  │  └─ After 2 min: mark player.ConnectionState = inactive
  │  └─ return nil
  │
  ├─ Clean up sessions: delete wsHandler.playerSessions[playerID]
  │
  └─ BroadcastToGame(Event{
       channel: "conn_event",
       type: "disconnection",
       data: {playerID, connectionState: "disconnected"}
     })
```

## Data Flow: Timer Expiry

```
Timer expired (e.g., day phase 3 minutes up)
  ↓
timerService callback: engine.handleTimerExpiry(gameID, phase, roleType)
  ├─ Match phase:
  │  ├─ PhaseDay: gameEngine.TransitionToVote(gameID)
  │  │  ├─ Create vote: voteService.StartVillageVote(gameID, alivePlayers)
  │  │  ├─ Update game: phase = vote
  │  │  ├─ gameRepo.SaveGame()
  │  │  ├─ BroadcastToGame(Event{type: "vote_started"})
  │  │  └─ StartPhaseTimer(vote) → 2 minutes
  │  │
  │  ├─ PhaseVote: gameEngine.ProcessVoteResult(gameID)
  │  │  └─ [See Vote Result data flow above]
  │  │
  │  └─ PhaseNight (roleType): gameEngine.handleNightRoleTimeout()
  │     ├─ If seer timeout: mark seer as acted
  │     ├─ If werewolf timeout: resolve vote anyway
  │     └─ If witch timeout: proceed
  │
  └─ Possibly transition to next phase
```

## Data Flow: Broadcasting Event to Room

```
Service calls: broadcaster.BroadcastToGame(gameID, payload)
  ↓
WebSocketHandler.BroadcastToGame()
  ├─ Acquire read lock
  ├─ Get all sessions in room: wsHandler.rooms[gameID]
  ├─ For each session:
  │  └─ melody.Session.Write(payload)
  └─ Release lock
  ↓
Client receives WebSocket message with event
  ├─ Parse Event{channel, type, data}
  └─ Handle per channel/type
```

## Data Flow: Sending to Specific Player

```
Service calls: playerSender.SendToPlayer(playerID, payload)
  ↓
WebSocketHandler.SendToPlayer()
  ├─ Acquire read lock
  ├─ Get session: wsHandler.playerSessions[playerID]
  ├─ If exists: melody.Session.Write(payload)
  └─ Release lock
  ↓
Player client receives message
```

## Data Flow: Win Condition Check

```
After phase transitions, check: gameEngine.checkWinCondition(gameID)
  ├─ Get all players
  ├─ Separate by clan (alive players)
  ├─ If no werewolves alive: villagers win
  ├─ If werewolves >= villagers: werewolves win
  ├─ If no one alive: no winner (draw)
  ├─ If someone won:
  │  ├─ Update game: status = ended
  │  ├─ gameRepo.SaveGame()
  │  ├─ BroadcastToGame(Event{type: "win", data: {winners, clan}})
  │  ├─ playerService.CleanupGamePlayers(gameID)
  │  └─ Clean up timers
  └─ If no win yet: continue game
```

## Data Flow: Chat Message

```
Client → WebSocket message {
  channel: "game_event",
  type: "chat_message",
  data: {gameID, playerID, channel: "town"|"werewolf"|"dead", message}
}
  ↓
WebSocketHandler.HandleChatMessage()
  ├─ Validation:
  │  ├─ Get player and game
  │  ├─ chatService.CanSendToChannel(player, channel, gamePhase)
  │  │  └─ Rules:
  │  │     ├─ town: day/vote phase, alive players
  │  │     ├─ werewolf: night phase, alive werewolves
  │  │     └─ dead: only dead players
  │  └─ Validate message length
  │
  ├─ Get recipients: chatService.GetChannelRecipients(channel, allPlayers, phase)
  │
  └─ For each recipient:
    └─ SendToPlayer(recipientID, Event{
         type: "chat_message",
         data: {playerID, username, message}
       })
```

## Data Flow: Player Visibility

```
When game state changes or player joins:
  ├─ For each player in game:
  │  ├─ Get all game players
  │  ├─ visibilityService.BuildPlayersDetailsForPlayer(player, allPlayers, gamePhase)
  │  │  ├─ Player always sees own role
  │  │  ├─ Werewolves see each other's roles
  │  │  ├─ Dead players' roles visible at day/vote
  │  │  ├─ Others: username + alive status only
  │  │  └─ Return filtered []PlayersDetailsData
  │  │
  │  └─ SendToPlayer(playerID, Event{
  │       type: "game_data",
  │       data: {game, players: visiblePlayers}
  │     })
```

## Service Dependencies & Responsibility Matrix

| Service | Uses | Used By | Responsibility |
|---------|------|---------|-----------------|
| GameService | GameRepo, PlayerRepo | GameEngine, Controllers | Create, join, configure games |
| PlayerService | PlayerRepo, GameRepo, ConnChecker | WebSocketHandler, GameEngine | Player lifecycle, reconnection |
| GameEngine | All services | WebSocketHandler | Game flow orchestration, timers |
| TimerService | Broadcaster | GameEngine | Phase and role timers |
| VoteService | VoteRepo, Broadcaster, PlayerSender | GameEngine | Voting mechanics |
| NightService | Broadcaster, VoteService | GameEngine | Night phase coordination |
| VisibilityService | - | WebSocketHandler | Role visibility filtering |
| ChatService | - | WebSocketHandler | Chat permissions |
| WebSocketHandler | GameService, PlayerService, GameEngine | Routes | WebSocket lifecycle, event routing |

## Circular Dependency Resolution

The circular dependency between PlayerService and WebSocketHandler is resolved through:

1. **Deferred Wiring**:
   - WebSocketHandler created without PlayerService
   - PlayerService created separately
   - Later: wsHandler.SetPlayerService(playerService)
   - Later: playerService.connChecker = wsHandler

2. **Interface Segregation**:
   - PlayerService depends on ConnectionChecker interface (not WebSocketHandler)
   - WebSocketHandler implements ConnectionChecker
   - Allows inversion of control

3. **Callback Patterns**:
   - GameEngine.SetExpiryCallback(engine.handleTimerExpiry)
   - Events flow out via Broadcaster/PlayerSender
   - No inbound dependencies from services to handlers
