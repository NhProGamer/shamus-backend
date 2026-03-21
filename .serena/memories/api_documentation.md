# Shamus Backend - API Documentation

## Overview

The Shamus Backend exposes two API types:
- **REST API**: Game creation and health checks
- **WebSocket API**: Real-time game interactions

Both APIs are documented using industry-standard specifications and served with interactive UIs in debug mode.

## Documentation Files

| File | Format | Purpose |
|------|--------|---------|
| `docs/api/openapi.yaml` | OpenAPI 3.1 | REST API specification |
| `docs/api/asyncapi.yaml` | AsyncAPI 2.6 | WebSocket API specification |
| `docs/api/swagger-ui.html` | HTML | Swagger UI viewer (CDN) |
| `docs/api/asyncapi-ui.html` | HTML | AsyncAPI React viewer (CDN) |

## Accessing Documentation

### In Debug Mode

When `config.Debug: true`:

| Route | Description |
|-------|-------------|
| `/docs/rest` | Swagger UI for REST API |
| `/docs/ws` | AsyncAPI UI for WebSocket API |
| `/docs/api/openapi.yaml` | Raw OpenAPI spec |
| `/docs/api/asyncapi.yaml` | Raw AsyncAPI spec |
| `/docs` | Redirects to `/docs/rest` |

### Configuration

```yaml
# config.yaml
debug: true  # Enables /docs/* routes
```

### Route Setup (`routes.go`)

```go
if ctx.Config.Debug {
    r.Static("/docs/api", "./docs/api")
    r.GET("/docs/rest", func(c *gin.Context) { c.File("./docs/api/swagger-ui.html") })
    r.GET("/docs/ws", func(c *gin.Context) { c.File("./docs/api/asyncapi-ui.html") })
    r.GET("/docs", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/docs/rest") })
}
```

## REST API Summary (OpenAPI)

### Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/app/api/v1/game` | OIDC | Create a new game |
| `GET` | `/health` | None | Health check (Redis) |
| `GET` | `/ready` | None | Readiness probe |
| `GET` | `/live` | None | Liveness probe |

### Schemas

- `Game`: id, status, phase, day, players, hostId, settings
- `GameSettings`: roles map (roleType → count)
- `CreateGameResponse`: gameID
- `Error`: error message

## WebSocket API Summary (AsyncAPI)

### Connection

```
wss://{host}/app/ws/{gameId}
```

Requires OIDC authentication (cookie/token from HTTP).

### Channels (Message Types)

#### Server → Client

| Channel | Purpose | Examples |
|---------|---------|----------|
| `notification` | Info updates | game_state, player_joined, chat_message, vote_result |
| `prompt` | Action required | vote, select_player, select_option |

#### Client → Server

| Channel | Purpose | Examples |
|---------|---------|----------|
| `command` | Player actions | send_chat, update_settings, start_game, kick_player |
| `response` | Prompt answers | Vote response, player selection |

### Message Format

All messages are JSON with a `channel` field:

```json
// Notification (server → client)
{
  "channel": "notification",
  "type": "player_joined",
  "payload": { "playerId": "...", "username": "Alice" }
}

// Prompt (server → client)
{
  "channel": "prompt",
  "type": "vote",
  "id": "prompt-uuid",
  "context": "village_vote",
  "payload": { "message": "...", "eligibleTargets": [...] },
  "expiresAt": "2026-03-21T10:00:00Z",
  "timeout": 120
}

// Command (client → server)
{
  "channel": "command",
  "type": "send_chat",
  "payload": { "message": "Hello!", "channel": "village" }
}

// Response (client → server)
{
  "channel": "response",
  "promptId": "prompt-uuid",
  "response": { "targetId": "player-id" }
}
```

### Notification Types (20+)

| Category | Types |
|----------|-------|
| Phase | `phase_changed` |
| Players | `player_joined`, `player_left`, `player_died`, `player_inactive`, `host_changed` |
| Game | `game_state`, `game_started`, `game_ended`, `role_reveal` |
| Timers | `timer_started`, `timer_tick`, `timer_expired` |
| Actions | `seer_result` |
| Votes | `vote_started`, `vote_update`, `vote_result`, `mayor_tiebreaker` |
| Chat | `chat_message` |
| System | `error`, `ack` |

### Command Types (5)

| Type | Payload | Description |
|------|---------|-------------|
| `send_chat` | message, channel | Send chat message |
| `update_settings` | roles | Update game settings (host) |
| `start_game` | - | Start the game (host) |
| `leave_game` | - | Leave current game |
| `kick_player` | playerId, reason | Kick player (host) |

### Prompt Types (4)

| Type | Context Examples | Response |
|------|------------------|----------|
| `vote` | village_vote, werewolf_vote | targetId, abstained |
| `select_player` | seer_vision, witch_poison | playerId, skipped |
| `select_option` | witch_potion | optionId, targetId |
| `confirm` | - | confirmed |

### Error Codes

| Code | Description |
|------|-------------|
| `MISSING_CONTEXT` | Session data missing |
| `INVALID_JSON` | Malformed JSON |
| `UNKNOWN_COMMAND` | Unknown command type |
| `INVALID_PAYLOAD` | Invalid payload format |
| `NOT_HOST` | Action requires host |
| `GAME_NOT_WAITING` | Game already started |
| `MESSAGE_EMPTY` | Empty chat message |
| `MESSAGE_TOO_LONG` | Chat exceeds 500 chars |
| `CHANNEL_FORBIDDEN` | Cannot send to channel |
| `NOT_IN_GAME` | Player not in this game |
| `PROMPT_NOT_FOUND` | Prompt ID not found |
| `WRONG_PLAYER` | Not your prompt |
| `PROMPT_EXPIRED` | Prompt timed out |
| `ALREADY_ANSWERED` | Prompt already answered |

## Updating Documentation

When adding new features:

1. **New REST endpoint**: Update `docs/api/openapi.yaml`
2. **New notification type**: Update `docs/api/asyncapi.yaml` (messages + schemas)
3. **New command type**: Update `docs/api/asyncapi.yaml` (messages + schemas)
4. **New prompt type**: Update `docs/api/asyncapi.yaml` (messages + schemas)

### Validation

Use online validators:
- OpenAPI: https://editor.swagger.io/
- AsyncAPI: https://studio.asyncapi.com/
