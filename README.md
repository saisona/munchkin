# Munchin Backend

Munchin is a backend service for an online, turn-based, card-game-inspired experience (loosely inspired by games like *Munchkin*, but with original mechanics and content).

This repository contains the **Go backend**, built with a strong focus on:
- Clear architectural boundaries
- Testability
- Observability (Prometheus + Grafana)
- Incremental scalability

At the current stage, the backend implements **authentication** and **lobby management**. Gameplay will be added later on top of these foundations.

---

## High-level architecture

The backend follows a layered, dependency-inverted design:


Key principles:
- Services are **pure business logic** (no Echo, no HTTP).
- Handlers adapt HTTP requests to services.
- Routing only wires paths to handlers.
- Metrics and middleware are centralized and explicit.

---

## Tech stack

- **Language**: Go (>= 1.22 recommended)
- **HTTP framework**: Echo v4
- **Auth**: JWT (HS256, symmetric key)
- **Metrics**: Prometheus
- **Visualization**: Grafana
- **Password hashing**: bcrypt

---

## Project structure

Each package contains a `doc.go` file describing its responsibilities and boundaries.

---

## Authentication

Authentication is based on **JWT (HS256)**.

### Implemented features
- Player registration (`POST /auth/register`)
- Player login (`POST /auth/login`)
- Identity endpoint (`GET /auth/me`)
- JWT verification middleware
- Prometheus metrics for auth success/failure

### JWT details
- Tokens contain the player ID as the `sub` claim
- Tokens are signed with a symmetric secret (`JWT_SECRET`)
- Token verification happens in middleware
- Business logic never parses JWTs directly

---

## Lobby management

Lobbies represent **pre-game coordination spaces**.

### Implemented features
- Create a lobby
- List lobbies
- Join / leave a lobby
- Host-driven game start (future hook)

Lobbies are intentionally kept **separate from gameplay logic**.  
A lobby will later map 1:1 to a game instance.

---

## Observability

### Metrics
The service exposes Prometheus metrics at: `GET /metrics`

Examples:
- `munchin_http_request_duration_seconds`
- `munchin_auth_success_total`
- `munchin_auth_failures_total`
- `munchin_lobby_active`

Metrics are:
- Declared centrally
- Registered once at startup
- Emitted from middleware and services (not handlers)

### Logging
- Structured logging (Echo middleware)
- Request IDs attached early in the request lifecycle

---

## Configuration

Configuration is provided via environment variables.

| Variable       | Description                          |
|----------------|--------------------------------------|
| `JWT_SECRET`   | Secret used to sign JWTs (required)  |
| `PORT`         | HTTP port (default: 1337)            |

Example:
```bash
export JWT_SECRET=$(openssl rand -base64 32)
export PORT=1337
```
## Running locally

```bash
go run ./cmd/server
```

Then:

API available at `http://localhost:1337`

Metrics at `http://localhost:1337/metrics`

Testing philosophy

Services are tested in isolation using fake repositories

Handlers are tested with httptest

Routing can be tested end-to-end via Echo

Middleware is unit-tested separately

No business logic lives in handlers or routers.

