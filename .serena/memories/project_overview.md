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
├── cmd/server/          # Entry point (main.go)
├── internal/
│   ├── domain/          # Core business logic (entities, ports, errors)
│   │   ├── entities/    # Game, Player, Role, Vote, Ability, Event
│   │   ├── ports/       # Interfaces (services, repositories)
│   │   ├── errors/      # Structured application errors
│   │   ├── helpers/     # Game logic helpers
│   │   └── constants/   # Game constants
│   ├── adapters/        # Interface implementations
│   │   ├── api/ws/      # WebSocket handler (EventService)
│   │   ├── app/         # Application services (GameService, etc.)
│   │   └── infra/       # Redis repositories
│   ├── application/     # (empty - use adapters/app)
│   └── infrastructure/  # HTTP layer
│       ├── config/      # Config loading
│       ├── controllers/ # HTTP handlers
│       ├── routes/      # Route definitions
│       └── middlewares/ # Auth, etc.
└── pkg/
    ├── logger/          # Zerolog wrapper
    └── utils/           # Shared utilities
```

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
- `TimerService`: Phase/role timers
- `ChatService`: Chat channel permissions
- `VisibilityService`: Role visibility rules
