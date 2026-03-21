# SHAMUS BACKEND - START HERE

## Welcome

**Shamus** is a production-grade Werewolf/Mafia game backend built in Go with hexagonal architecture.

**Project:** Werewolf Game Backend  
**Language:** Go 1.24  
**Architecture:** Hexagonal (Ports & Adapters)  
**Real-time:** WebSocket with Command/Prompt/Notification pattern  
**Persistence:** Redis  
**Auth:** OIDC  
**Updated:** March 2026

## Quick Navigation

| I want to... | Read this |
|--------------|-----------|
| Understand the project | `onboarding_summary` |
| Learn the architecture | `architecture_deep_dive` |
| Find a type/interface | `key_types_and_interfaces` |
| Understand data flows | `integration_and_data_flow` |
| Learn testing patterns | `testing_and_validation` |
| See code conventions | `code_style_conventions` |
| Find useful commands | `suggested_commands` |
| Understand API docs | `api_documentation` |

## Reading Order

1. **`onboarding_summary`** (10 min) - Quick overview of everything
2. **`architecture_deep_dive`** (20 min) - Layer-by-layer breakdown
3. **`key_types_and_interfaces`** (Reference) - Type catalog
4. **`integration_and_data_flow`** (15 min) - How data moves
5. **`api_documentation`** (5 min) - REST + WebSocket API docs

## Architecture at a Glance

```
┌──────────────────────────────────────────────┐
│  Primary Adapters                             │
│  ├─ HTTP: REST API, health checks            │
│  └─ WebSocket: Commands, Prompts, Notifs     │
├──────────────────────────────────────────────┤
│  Application Layer                            │
│  ├─ Services: Game, Player, Vote, Night...   │
│  └─ Orchestration: GameEngineV2              │
├──────────────────────────────────────────────┤
│  Domain Layer                                 │
│  ├─ Entities: Game, Player, Role, Prompt...  │
│  ├─ Ports: Interfaces for all components     │
│  └─ Errors, Constants, Helpers               │
├──────────────────────────────────────────────┤
│  Secondary Adapters                           │
│  └─ Redis: Game, Player, Vote repositories   │
└──────────────────────────────────────────────┘
```

## Key Directory Structure

```
shamus-backend/
├── cmd/server/main.go                    # Entry point
├── internal/
│   ├── domain/                           # Business logic (NO external deps)
│   │   ├── entities/                     # Game, Player, Role, Vote, Prompt, etc.
│   │   ├── ports/                        # Interface contracts
│   │   ├── errors/                       # Structured errors
│   │   ├── helpers/                      # Utilities
│   │   └── constants/                    # Game config
│   ├── application/
│   │   ├── services/                     # Service implementations
│   │   └── orchestration/                # GameEngineV2
│   ├── adapters/
│   │   ├── primary/
│   │   │   ├── http/                     # REST API
│   │   │   └── websocket/                # WebSocket handling
│   │   └── secondary/
│   │       └── redis/                    # Persistence
│   └── infrastructure/config/            # YAML config
├── docs/api/                             # API documentation
│   ├── openapi.yaml                      # REST (OpenAPI 3.1)
│   └── asyncapi.yaml                     # WebSocket (AsyncAPI 2.6)
└── pkg/                                  # Shared utilities
```

## WebSocket Message Flow

```
Client                          Server
   │                               │
   │── command ──────────────────▶│ CommandHandler
   │◀── ack/error ────────────────│
   │                               │
   │◀── notification ─────────────│ NotificationService
   │                               │
   │◀── prompt ───────────────────│ PromptService
   │── response ─────────────────▶│
   │◀── ack/error ────────────────│
```

## Game Flow

```
CREATE → CONFIGURE → START → PLAY → END
  │         │          │       │      │
  POST    update_    start_  Night   Win
  /game   settings   game   →Day→   condition
                            Vote
```

## Essential Commands

```bash
# Build & Run
go build ./...
go run cmd/server/main.go

# Test & Lint
go test ./...
golangci-lint run

# Full check
go fmt ./... && go vet ./... && go test ./... && go build ./...
```

## API Documentation (Debug Mode)

When `config.Debug: true`:
- `/docs/rest` - Swagger UI
- `/docs/ws` - AsyncAPI UI

## Memory Files

| Memory | Purpose |
|--------|---------|
| `00_START_HERE` | This file - navigation |
| `onboarding_summary` | Comprehensive overview |
| `architecture_deep_dive` | Layer details |
| `key_types_and_interfaces` | Type reference |
| `integration_and_data_flow` | Data flow diagrams |
| `project_overview` | Quick overview |
| `code_style_conventions` | Coding standards |
| `suggested_commands` | CLI commands |
| `testing_and_validation` | Test patterns |
| `api_documentation` | API docs reference |
| `task_completion_checklist` | PR checklist |

---

**Start with `onboarding_summary` for the full picture!**
