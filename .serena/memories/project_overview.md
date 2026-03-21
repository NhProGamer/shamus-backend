# Shamus Backend - Project Overview

## Purpose

**Shamus** is a real-time multiplayer **Werewolf/Mafia game** backend written in Go. It manages game rooms, player connections, role assignments, voting phases, and night actions via WebSocket communication.

## Tech Stack

- **Language**: Go 1.24
- **Web Framework**: Gin (HTTP router + middleware)
- **WebSocket**: Melody (based on Gorilla WebSocket)
- **Database**: Redis (game state, player sessions, votes)
- **Authentication**: OIDC (OpenID Connect)
- **Logging**: Zerolog
- **Config**: YAML
- **API Docs**: OpenAPI 3.1 (REST) + AsyncAPI 2.6 (WebSocket)

## Architecture

The project follows a **Hexagonal/Ports & Adapters** architecture:

```
shamus-backend/
├── cmd/server/main.go                 # Entry point, dependency injection
├── internal/
│   ├── domain/                        # Core business logic (NO external deps)
│   │   ├── entities/                  # Game, Player, Role, Vote, Prompt, Command, Notification
│   │   │   ├── commands/              # Command payloads
│   │   │   ├── prompts/               # Prompt payloads and responses
│   │   │   ├── roles/                 # Role implementations
│   │   │   └── abilities/             # Ability implementations
│   │   ├── ports/                     # Interfaces (services, repositories)
│   │   ├── errors/                    # Structured application errors
│   │   ├── helpers/                   # Game logic helpers, sanitization
│   │   └── constants/                 # Game constants
│   ├── application/                   # Use cases and orchestration
│   │   ├── services/                  # Service implementations
│   │   └── orchestration/             # GameEngineV2 (game flow orchestrator)
│   ├── adapters/                      # Interface implementations
│   │   ├── primary/                   # Driving adapters (incoming)
│   │   │   ├── http/                  # HTTP controllers, routes, middlewares
│   │   │   └── websocket/             # WebSocket handler, command handler, sessions
│   │   └── secondary/                 # Driven adapters (outgoing)
│   │       └── redis/                 # Redis repositories
│   └── infrastructure/
│       └── config/                    # YAML config loading
├── docs/
│   └── api/                           # API documentation (OpenAPI, AsyncAPI)
└── pkg/
    ├── logger/                        # Zerolog wrapper
    └── utils/                         # Shared utilities
```

## Hexagonal Architecture Layers

### Domain Layer (`internal/domain/`)
- **Entities**: Pure business objects (Game, Player, Role, Vote, Prompt, Command, Notification)
- **Ports**: Interface contracts for services, repositories, broadcasting
- **Errors**: Structured error types with error codes
- **Helpers**: Game logic utilities (role assignment, sanitization)
- **No external dependencies**

### Application Layer (`internal/application/`)
- **Services**: Use case implementations
- **Orchestration**: GameEngineV2 - orchestrates game flow and phase transitions
- **Depends only on domain ports**

### Adapters Layer (`internal/adapters/`)
- **Primary (Driving)**: Handle incoming requests
  - HTTP: REST API controllers, routes, OIDC middleware
  - WebSocket: Real-time communication, command processing, session management
- **Secondary (Driven)**: Handle outgoing dependencies
  - Redis: Game/Player/Vote persistence

### Infrastructure Layer (`internal/infrastructure/`)
- **Config**: Application configuration loading
- `Debug` flag enables API documentation routes

## Game Mechanics

- **Roles**: Villager, Werewolf, Seer, Witch
- **Phases**: Start → Night → Day → Vote → (loop)
- **Night Actions**: Seer vision, Werewolf vote, Witch heal/poison
- **Day Phase**: Discussion + Village vote to eliminate
- **Win Conditions**: Villagers win (no werewolves) or Werewolves win (≥ villagers)

## Key Services (Application Layer)

| Service | Responsibility |
|---------|----------------|
| **GameService** | Create/join games, update settings, start game |
| **PlayerService** | Connect/disconnect, player state, reconnection (2-min timeout) |
| **GameEngine** | Orchestrates phase transitions, actions, win conditions |
| **PromptService** | Interactive prompts with timeouts, group votes |
| **NotificationService** | Server-to-client notifications (20+ types) |
| **VoteService** | Village and werewolf voting mechanics |
| **NightService** | Night phase coordination (Seer → Werewolf → Witch) |
| **TimerService** | Phase and role timers |
| **VisibilityService** | Role-based information filtering |
| **ChatService** | Chat channel permissions |

## WebSocket Message Channels

| Channel | Direction | Purpose |
|---------|-----------|---------|
| `notification` | Server → Client | Informational updates (no response expected) |
| `prompt` | Server → Client | Interactive requests requiring player action |
| `command` | Client → Server | Player-initiated actions (chat, settings, etc.) |
| `response` | Client → Server | Answers to prompts |

## Key Files Reference

| Path | Purpose |
|------|---------|
| `cmd/server/main.go` | Entry point, DI wiring |
| `internal/domain/entities/` | Core entities |
| `internal/domain/ports/` | Interface contracts |
| `internal/application/services/` | Service implementations |
| `internal/application/orchestration/game_engine_v2.go` | Game flow orchestrator |
| `internal/adapters/primary/http/routes/routes.go` | HTTP + WebSocket routing |
| `internal/adapters/primary/websocket/handler.go` | WebSocket lifecycle |
| `internal/adapters/primary/websocket/command_handler.go` | Client command processing |
| `internal/adapters/secondary/redis/` | Redis repositories |
| `docs/api/openapi.yaml` | REST API specification |
| `docs/api/asyncapi.yaml` | WebSocket API specification |

## API Documentation

In debug mode (`config.Debug: true`):
- **REST API**: `/docs/rest` (Swagger UI)
- **WebSocket API**: `/docs/ws` (AsyncAPI UI)
- **Raw specs**: `/docs/api/openapi.yaml`, `/docs/api/asyncapi.yaml`
