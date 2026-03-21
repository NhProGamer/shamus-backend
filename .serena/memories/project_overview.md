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

## Architecture
The project follows a **Hexagonal/Ports & Adapters** architecture:

```
shamus-backend/
├── cmd/server/                    # Entry point (main.go)
├── internal/
│   ├── domain/                    # Core business logic
│   │   ├── entities/              # Game, Player, Role, Vote, Ability, Event
│   │   ├── ports/                 # Interfaces (services, repositories, broadcasting)
│   │   ├── errors/                # Structured application errors
│   │   ├── helpers/               # Game logic helpers
│   │   └── constants/             # Game constants
│   ├── application/               # Use cases and orchestration
│   │   ├── services/              # Application services (GameService, PlayerService, etc.)
│   │   └── orchestration/         # GameEngineV2 (game flow orchestrator)
│   ├── adapters/                  # Interface implementations
│   │   ├── primary/               # Driving adapters (incoming)
│   │   │   ├── http/              # HTTP controllers, routes, middlewares
│   │   │   └── websocket/         # WebSocket handler, session manager
│   │   └── secondary/             # Driven adapters (outgoing)
│   │       └── redis/             # Redis repositories
│   └── infrastructure/            # Cross-cutting concerns
│       └── config/                # YAML config loading
└── pkg/
    ├── logger/                    # Zerolog wrapper
    └── utils/                     # Shared utilities
```

## Hexagonal Architecture Layers

### Domain Layer (`internal/domain/`)
- **Entities**: Pure business objects (Game, Player, Role, Vote)
- **Ports**: Interface contracts for services, repositories, broadcasting
- **Errors**: Structured error types with error codes
- **No external dependencies**

### Application Layer (`internal/application/`)
- **Services**: Use case implementations (GameService, PlayerService, VoteService, etc.)
- **Orchestration**: GameEngineV2 - orchestrates game flow and phase transitions
- **Depends only on domain ports**

### Adapters Layer (`internal/adapters/`)
- **Primary (Driving)**: Handle incoming requests
  - HTTP: REST API controllers, routes, OIDC middleware
  - WebSocket: Real-time communication, session management
- **Secondary (Driven)**: Handle outgoing dependencies
  - Redis: Game/Player/Vote persistence

### Infrastructure Layer (`internal/infrastructure/`)
- **Config**: Application configuration loading
- Cross-cutting concerns only

## Game Mechanics
- **Roles**: Villager, Werewolf, Seer, Witch
- **Phases**: Start → Night → Day → Vote → (loop)
- **Night Actions**: Seer vision, Werewolf vote, Witch heal/poison
- **Day Phase**: Discussion + Village vote to eliminate

## Key Services (Ports)
- `GameService`: Create/join games, update settings, start game
- `PlayerService`: Connect/disconnect, player state
- `GameEngine`: Orchestrates phase transitions and actions
- `VoteService`: Village and werewolf voting
- `NightService`: Night phase action tracking
- `NotificationService`: Server-to-client notifications
- `PromptService`: Interactive prompts with timeouts
- `ChatService`: Chat channel permissions
- `VisibilityService`: Role visibility rules
