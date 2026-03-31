# Munchin API

Munchin is a Go backend for an educational, turn-based, card-game-inspired experience influenced by Munchkin.

The project currently focuses on backend foundations:
- authentication
- lobby lifecycle
- real-time room connectivity over WebSocket
- early gameplay state management
- observability
- testability

Gameplay is still incomplete, but the repository now includes the first building blocks for turn state, player projection, smoke testing, and API documentation.

---

## Current Scope

Implemented today:
- player registration and login
- JWT-based protected routes
- lobby creation, listing, join, start, delete
- health and readiness probes
- Swagger/OpenAPI source annotations
- WebSocket room connection
- game start setup with initial dealing
- initial game snapshot and first turn actions
- unit tests around the game core
- a Go smoke test for end-to-end validation

Not fully implemented yet:
- full deck management
- complete combat resolution
- full Munchkin card taxonomy
- advanced class/race abilities
- production-grade reconnect semantics

---

## Tech Stack

- Go
- Echo v4
- GORM + PostgreSQL
- Gorilla WebSocket
- JWT authentication
- Prometheus / Grafana
- Swagger via `swag`

---

## Architecture

The backend follows a layered design:

- `pkg/auth`: authentication and JWT
- `pkg/lobbies`: lobby lifecycle and pre-game coordination
- `pkg/game`: in-memory game room, state, commands, events, DTOs
- `pkg/api`: route wiring
- `pkg/health`: liveness, startup, readiness probes
- `pkg/telemetry`: metrics and tracing integration

Guiding principles:
- services own business logic
- handlers adapt transport concerns
- routing stays thin
- state mutation should flow through the room/game engine
- observability is explicit and centralized

---

## Main Endpoints

### Auth

- `POST /auth/register`
- `POST /auth/login`

### Health

- `GET /healthz`
- `GET /startupz`
- `GET /readyz`
- `GET /metrics`

### Lobbies

- `POST /lobby`
- `GET /lobby`
- `GET /lobby/model`
- `DELETE /lobby/{id}`
- `POST /lobby/{id}/join`
- `POST /lobby/{id}/start`
- `GET /lobby/{id}/ws`

### Swagger

- `GET /swagger/index.html`

---

## WebSocket Notes

The current WebSocket endpoint is:

`GET /lobby/{id}/ws?token=<jwt>`

Important:
- the WebSocket authentication currently relies on the `token` query parameter
- this is documented in the Swagger godoc for the endpoint
- this differs from the rest of the HTTP API, which uses `Authorization: Bearer <jwt>`

The current protocol reference lives in [PROTOCOL.md](/Users/inarix-alexandre-saison/Workspace/Dev/PERSO/Munchin/munchin-api/PROTOCOL.md).

Current MVP gameplay slice over WebSocket:
- start game from lobby
- receive initial `game_snapshot`
- send `PLAYER_ACTION`
- open the door
- reveal a dungeon card
- acknowledge a revealed card placeholder when needed
- discard during charity with `DISCARD_FOR_CHARITY`
- move to the next turn phase with validation

---

## Local Run

### Prerequisites

- Go installed and available in `PATH`
- PostgreSQL, or Docker Compose
- `JWT_SECRET` configured

Example:

```bash
export JWT_SECRET=$(openssl rand -base64 32)
export PORT=1337
```

### Run with Go

```bash
go run ./cmd/server
```

### Run with Docker Compose

```bash
docker compose up --build
```

Default local URLs:

- API: [http://localhost:1337](http://localhost:1337)
- Swagger: [http://localhost:1337/swagger/index.html](http://localhost:1337/swagger/index.html)
- Metrics: [http://localhost:1337/metrics](http://localhost:1337/metrics)

---

## Useful Commands

Available through the `Makefile`:

```bash
make build
make run
make dev
make test
make fmt
make lint
make swagger
make e2e-smoke
```

What they do:
- `make test`: runs Go tests with race detector and coverage
- `make swagger`: regenerates Swagger docs from source comments
- `make e2e-smoke`: runs the Go end-to-end smoke test

---

## Testing

### Unit Tests

Run all tests:

```bash
make test
```

Run only the game package:

```bash
go test ./pkg/game -v
```

Current unit-test coverage focuses on:
- command decoding
- player defaults
- combat strength computation
- turn phase progression
- invalid turn / invalid phase rejection
- DTO projection behavior

### End-to-End Smoke Test

The repository includes a small Go smoke test under `cmd/e2e-smoke`.

It validates the happy path:
- register player
- create lobby
- list lobbies
- start game
- connect to WebSocket
- read initial snapshot
- play through a first MVP turn:
- `OPEN_DOOR`
- optional `ACK_REVEALED_CARD`
- `LOOK_FOR_TROUBLE`
- `LOOT_ROOM`
- optional `DISCARD_FOR_CHARITY`
- `END_TURN`

Run it with:

```bash
go run ./cmd/e2e-smoke
```

or:

```bash
make e2e-smoke
```

Optional environment variables:

```bash
MUNCHIN_BASE_URL=http://localhost:1337
MUNCHIN_E2E_USERNAME=my-test-user
MUNCHIN_E2E_PASSWORD=secret123
```

If no username is provided, the smoke test generates a unique one automatically.

---

## Swagger / OpenAPI

Swagger source annotations are maintained in handlers and shared DTOs.

To regenerate docs:

```bash
make swagger
```

This uses:

```bash
swag init -g cmd/server/main.go --parseInternal
```

If `swag` is not installed:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## Observability

### Metrics

Prometheus metrics are exposed at:

`GET /metrics`

Examples:
- `munchin_http_request_duration_seconds`
- `munchin_auth_success_total`
- `munchin_auth_failures_total`
- `munchin_lobby_active`

### Logging

- structured logging
- request-scoped middleware
- telemetry hooks for traces and metrics

---

## Project Status

This repository is in an active foundation-building phase.

What is already improving:
- domain shape
- room/state flow
- DTO consistency
- testability
- API documentation
- local validation workflow

What still needs major implementation work:
- full rules engine
- deck/discard behavior
- treasure/dungeon flow
- combat and flee resolution
- richer client/server event contract

---

## Related Docs

- [AGENTS.md](/Users/inarix-alexandre-saison/Workspace/Dev/PERSO/Munchin/munchin-api/AGENTS.md)
- [PROTOCOL.md](/Users/inarix-alexandre-saison/Workspace/Dev/PERSO/Munchin/munchin-api/PROTOCOL.md)
- [CHANGELOG.md](/Users/inarix-alexandre-saison/Workspace/Dev/PERSO/Munchin/munchin-api/CHANGELOG.md)
