# Shamus Backend - Integration & Data Flow

## Initialization Flow (main.go)

### Phase 1: Configuration & Infrastructure
```
Config.LoadConfig() → YAML file
    ├─ Server: host, port, public_url
    ├─ OIDC: issuer, client_id, secret, scopes
    ├─ Redis: host, port, password, db
    ├─ Logger: level, pretty
    └─ Debug: bool (enables /docs routes)

Logger initialization → Zerolog
Gin router + Melody WebSocket manager
OIDC provider init
Redis client → Ping test
```

### Phase 2: Repositories
```
Redis client
    ├─→ RedisGameRepo
    ├─→ RedisPlayerRepo
    └─→ RedisVoteRepo
```

### Phase 3: Core Services
```
Repositories
    ├─→ GameService(gameRepo, playerRepo)
    ├─→ PlayerService(playerRepo, gameRepo)  // connChecker set later
    ├─→ VisibilityService()
    └─→ ChatService()
```

### Phase 4: WebSocket Components
```
SessionManager(melody)

NotificationService(sessionManager)

CommandHandler(gameService, playerService, chatService, notifier)
    ├─ disconnecter = nil  [SET LATER]
    ├─ gameEngine = nil    [SET LATER]
    └─ playerService set

WebSocketHandler(melody, sessions, promptService, commandHandler, notifier, gameService)
    └─ playerService = nil [SET LATER]
```

### Phase 5: Complete Wiring
```
wsHandler.SetPlayerService(playerService)
playerService.SetConnectionChecker(wsHandler)
commandHandler.SetDisconnecter(wsHandler)
```

### Phase 6: Game Flow Services
```
TimerService(broadcaster)
VoteService(voteRepo, broadcaster, playerSender)
NightService(broadcaster, voteService, playerRepo)

GameEngineV2(gameRepo, playerRepo, timerService, voteService, nightService, promptService, notifier, playerSender)

commandHandler.SetGameEngine(gameEngine)
```

### Phase 7: HTTP Server
```
Routes.InitRoutes(router, AppContext, redis)
    ├─ Health: /health, /ready, /live
    ├─ Docs (debug only): /docs/rest, /docs/ws, /docs/api/*
    ├─ Protected (/app with OIDC):
    │   ├─ Static files
    │   ├─ WebSocket: /ws/:gameID
    │   └─ API: /api/v1/game
    └─ Server.Run()
```

## WebSocket Message Flow

### Channel Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    WebSocket Connection                      │
├─────────────────────────────────────────────────────────────┤
│  Server → Client                                             │
│  ├─ notification: Informational updates (no response)       │
│  │   └─ Types: game_state, player_joined, chat_message...   │
│  └─ prompt: Interactive requests (response expected)         │
│      └─ Types: vote, select_player, select_option, confirm  │
├─────────────────────────────────────────────────────────────┤
│  Client → Server                                             │
│  ├─ command: Player-initiated actions                        │
│  │   └─ Types: send_chat, update_settings, start_game...    │
│  └─ response: Answers to prompts                             │
│      └─ Contains: promptId, response payload                 │
└─────────────────────────────────────────────────────────────┘
```

### Message Routing (Handler.onMessage)

```
Client WebSocket Message (JSON)
    │
    ▼
Parse channel field
    │
    ├─── channel: "command" ──────────────────────┐
    │                                              │
    │    CommandHandler.Handle(cmdCtx, cmd)        │
    │        │                                     │
    │        ├─ Validate player belongs to game    │
    │        │   (player.GameID == cmdCtx.GameID)  │
    │        │                                     │
    │        ├─ Route by cmd.Type:                 │
    │        │   ├─ send_chat → handleSendChat()   │
    │        │   │   └─ Sanitize message           │
    │        │   │   └─ Check channel permissions  │
    │        │   │   └─ NotifyChatMessage()        │
    │        │   ├─ update_settings → handleUpdateSettings()
    │        │   ├─ start_game → handleStartGame() │
    │        │   ├─ leave_game → handleLeaveGame() │
    │        │   └─ kick_player → handleKickPlayer()
    │        │                                     │
    │        └─ Return ack or error notification   │
    │                                              │
    └─── channel: "response" ─────────────────────┐
                                                   │
         PromptService.RespondToPrompt()           │
             │                                     │
             ├─ Validate prompt exists & pending   │
             ├─ Validate player matches            │
             ├─ Validate not expired               │
             ├─ Process response                   │
             │   └─ For group votes: update state  │
             └─ Return ack or error                │
```

## Data Flow: Create Game

```
Client → HTTP POST /app/api/v1/game
    │
    ▼
controllers.PostGameHandler
    ├─ Extract userID from OIDC context
    └─ gameService.CreateNewGame(ctx, playerID)
        ├─ Generate gameID (UUID)
        ├─ Create Game{
        │   Status: waiting,
        │   Phase: start,
        │   Settings: {4V, 2W, 1S, 1W}
        │ }
        └─ gameRepo.SaveGame() → Redis
    │
    ▼
HTTP 200 {gameID: "..."}
```

## Data Flow: Player Connect (WebSocket)

```
Client → WebSocket /app/ws/:gameID
    │
    ▼
Handler.onConnect(session)
    │
    ├─ ExtractSessionData(session)  // Safe type assertions
    │   └─ Returns: gameID, playerID, username
    │
    ├─ playerService.HandleConnect(ctx, gameID, playerID, username)
    │   ├─ Get game from repo
    │   ├─ If new player (game.status == waiting):
    │   │   ├─ Create Player
    │   │   ├─ playerRepo.SavePlayer()
    │   │   └─ playerRepo.AddPlayerToGame()
    │   └─ If reconnecting (within 2-min timeout):
    │       ├─ Cancel timeout timer
    │       └─ Mark player connected
    │
    ├─ sessions.JoinRoom(gameID, playerID, session)
    │
    ├─ notifier.NotifyPlayerJoined() or NotifyAll("player_reconnected")
    │
    └─ sendGameStateToPlayer(ctx, gameID, playerID)
        └─ NotifyPlayer(NotifGameState, {game, players})
```

## Data Flow: Start Game

```
Client → WebSocket {channel: "command", type: "start_game"}
    │
    ▼
CommandHandler.handleStartGame(cmdCtx)
    │
    ├─ Validate player is host
    │
    ├─ gameService.StartGame(ctx, gameID, playerID)
    │   ├─ game.CanStart() validation
    │   ├─ helpers.AssignRoles(players, settings)
    │   ├─ playerRepo.SavePlayers()
    │   ├─ game.Status = active, game.Phase = night
    │   └─ gameRepo.SaveGame()
    │
    ├─ gameEngine.StartGameFlow(ctx, gameID)
    │   ├─ NotifyAll(NotifGameStarted)
    │   ├─ For each player: NotifyPlayer(NotifRoleReveal)
    │   ├─ nightService.StartNight()
    │   └─ Start first role's turn (Seer)
    │
    └─ Return ack
```

## Data Flow: Night Phase Actions

### Seer Vision

```
GameEngine starts seer turn:
    │
    ├─ Find alive seer
    ├─ Get valid targets (alive, not self)
    └─ promptService.CreatePrompt({
        Type: select_player,
        Context: "seer_vision",
        Timeout: 30s
       })
        │
        ▼
    PromptService.CreatePrompt()
        ├─ Store prompt
        ├─ Start timer
        └─ sendPromptToPlayer() → WebSocket prompt message

Client receives prompt, selects target

Client → WebSocket {channel: "response", promptId, response: {playerId}}
    │
    ▼
PromptService.RespondToPrompt()
    ├─ Validate & mark answered
    └─ Trigger callback → GameEngine.handleSeerVisionResponse()
        ├─ Get target's role
        ├─ NotifyPlayer(NotifSeerResult, {targetId, isWerewolf})
        └─ nightService.AdvancePhase() → Werewolf turn
```

### Werewolf Vote

```
GameEngine starts werewolf turn:
    │
    └─ promptService.CreateGroupVote({
        Context: "werewolf_vote",
        Voters: [werewolf IDs],
        Targets: [non-werewolf alive IDs],
        Timeout: 60s
       })
        │
        ▼
    For each werewolf:
        └─ Send vote prompt

Werewolf votes → PromptService
    ├─ Update GroupVoteState
    ├─ NotifyAll werewolves (NotifVoteUpdate) with current votes
    └─ If all voted or timeout:
        └─ Resolve vote → GameEngine.handleWerewolfVoteResult()
            ├─ nightService.RecordWerewolfVictim()
            └─ Advance to Witch turn
```

### Witch Action

```
GameEngine starts witch turn:
    │
    ├─ Get werewolf victim (if any)
    ├─ Check available potions (heal/poison)
    └─ promptService.CreatePrompt({
        Type: select_option,
        Context: "witch_potion",
        Payload: WitchPotionPayload
       })

Witch chooses → PromptService → GameEngine.handleWitchResponse()
    ├─ If "heal": nightService.RecordWitchAction(heal=victim)
    ├─ If "poison": Prompt for target → RecordWitchAction(poison=target)
    └─ nightService.AdvancePhase() → End night
```

### Night Resolution

```
nightService.IsNightComplete() == true
    │
    ▼
GameEngine.processNightEnd()
    ├─ nightService.GetPendingDeaths()
    │   └─ Returns: werewolf victim (if not healed) + poison victim
    │
    ├─ For each death:
    │   ├─ player.Kill()
    │   ├─ playerRepo.SavePlayer()
    │   └─ NotifyAll(NotifPlayerDied, {playerId, cause, role})
    │
    ├─ Check win condition
    │
    ├─ If game continues:
    │   ├─ game.Phase = day
    │   ├─ gameRepo.SaveGame()
    │   ├─ NotifyAll(NotifPhaseChanged, {phase: "day"})
    │   └─ timerService.StartPhaseTimer(day, 3min)
    │
    └─ nightService.ClearNight()
```

## Data Flow: Day Phase → Vote

```
Timer expires (3 min day discussion)
    │
    ▼
timerService callback → GameEngine.handleTimerExpiry()
    │
    ├─ game.Phase = vote
    ├─ gameRepo.SaveGame()
    │
    ├─ voteService.StartVillageVote(gameID, alivePlayers)
    │   └─ Creates Vote entity
    │
    ├─ promptService.CreateGroupVote({
    │   Context: "village_vote",
    │   Voters: alive players,
    │   Targets: alive players,
    │   CanAbstain: true
    │ })
    │
    └─ NotifyAll(NotifVoteStarted)
```

## Data Flow: Village Vote Resolution

```
All players vote (or timeout)
    │
    ▼
PromptService resolves group vote
    │
    └─ Callback → GameEngine.handleVillageVoteResult()
        │
        ├─ voteService.ResolveVote()
        │   └─ vote.Resolve() → VoteResult{target, counts, isTie}
        │
        ├─ If target (no tie):
        │   ├─ player.Kill()
        │   ├─ playerRepo.SavePlayer()
        │   └─ NotifyAll(NotifPlayerDied)
        │
        ├─ NotifyAll(NotifVoteResult, {result, counts})
        │
        ├─ Check win condition
        │
        └─ If game continues:
            ├─ game.Phase = night, game.Day++
            ├─ gameRepo.SaveGame()
            ├─ nightService.StartNight()
            └─ Start seer turn
```

## Data Flow: Chat Message

```
Client → {channel: "command", type: "send_chat", payload: {message, channel}}
    │
    ▼
CommandHandler.handleSendChat()
    │
    ├─ Validate message length
    │
    ├─ helpers.SanitizeChatMessage(message)
    │   └─ Remove control characters
    │
    ├─ Get game and sender
    │
    ├─ chatService.CanSendToChannel(sender, channel, phase)
    │   └─ Rules: village=day+alive, werewolf=night+werewolf, dead=dead
    │
    ├─ getChannelRecipients(channel, game)
    │   └─ Filter players by channel rules
    │
    └─ notifier.NotifyChatMessage(recipients, senderID, username, message, channel, timestamp)
        └─ For each recipient: SendToPlayer(NotifChatMessage)
```

## Data Flow: Player Disconnect

```
WebSocket disconnect event
    │
    ▼
Handler.onDisconnect(session)
    │
    ├─ ExtractGameID, ExtractPlayerID (safe)
    │
    ├─ sessions.LeaveRoom(gameID, playerID, session)
    │
    ├─ playerService.HandleDisconnect(ctx, gameID, playerID)
    │   ├─ player.Disconnect()
    │   ├─ playerRepo.SavePlayer()
    │   └─ Start 2-min timer:
    │       └─ After timeout: player.SetInactive()
    │
    └─ notifier.NotifyPlayerLeft(gameID, playerID, username, "disconnected")
```

## Data Flow: Kick Player

```
Client → {channel: "command", type: "kick_player", payload: {playerId, reason}}
    │
    ▼
CommandHandler.handleKickPlayer()
    │
    ├─ Validate sender is host
    ├─ Validate target != self
    ├─ Validate game.Status == waiting
    │
    ├─ Get target player
    │
    ├─ playerRepo.RemovePlayerFromGame()
    ├─ Remove from game.Players
    ├─ gameRepo.SaveGame()
    │
    ├─ notifier.NotifyPlayerLeft(gameID, targetID, username, "kicked")
    │
    └─ disconnecter.DisconnectPlayer(gameID, targetID, reason)
        └─ sessions.DisconnectPlayer() → Close WebSocket
```

## Win Condition Check

```
GameEngine.checkWinCondition(gameID)
    │
    ├─ Get all players
    ├─ Count alive by clan
    │
    ├─ If no werewolves alive → Villagers win
    ├─ If werewolves >= villagers → Werewolves win
    ├─ If no one alive → Draw
    │
    └─ If winner:
        ├─ game.Status = ended
        ├─ gameRepo.SaveGame()
        ├─ NotifyAll(NotifGameEnded, {winner, survivors})
        ├─ timerService.CancelTimer()
        └─ playerService.CleanupGamePlayers()
```

## Broadcasting Patterns

### NotifyPlayer (single recipient)
```go
notifier.NotifyPlayer(playerID, NotifRoleReveal, {role: "seer"})
    │
    └─ sessionManager.SendToPlayer(playerID, payload)
        └─ session.Write(json)
```

### NotifyAll (room broadcast)
```go
notifier.NotifyAll(gameID, NotifPhaseChanged, {phase: "day"})
    │
    └─ sessionManager.BroadcastToGame(gameID, payload)
        └─ For each session in room: session.Write(json)
```

### NotifyExcept (broadcast minus one)
```go
notifier.NotifyExcept(gameID, excludeID, NotifPlayerJoined, {...})
    │
    └─ For each player in game except excludeID:
        └─ SendToPlayer(playerID, payload)
```

## Service Dependency Matrix

| Service | Uses | Used By |
|---------|------|---------|
| GameService | GameRepo, PlayerRepo | CommandHandler, GameEngine |
| PlayerService | PlayerRepo, GameRepo, ConnectionChecker | Handler, CommandHandler |
| GameEngine | All services, Repos | CommandHandler (via callbacks) |
| PromptService | PlayerSender, NotificationService | GameEngine, Handler |
| NotificationService | SessionManager | All services |
| VoteService | VoteRepo, Broadcaster | GameEngine |
| NightService | Broadcaster, VoteService, PlayerRepo | GameEngine |
| TimerService | Broadcaster | GameEngine |
| VisibilityService | - | Handler |
| ChatService | - | CommandHandler |
| CommandHandler | GameService, PlayerService, ChatService, Notifier | Handler |
| SessionManager | Melody | NotificationService, Handler |
