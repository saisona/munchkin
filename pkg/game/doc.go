// Package game implements the real-time game engine for Munchin.
//
// The game package is responsible for managing the lifecycle of a game session,
// including player connections, command handling, state transitions, and
// event broadcasting. It is designed as a deterministic, event-driven state
// machine intended to be driven over WebSockets.
//
// Architecture overview:
//
//   - GameHub
//     A concurrency-safe registry of active game rooms. It is responsible for
//     creating, retrieving, and deleting GameRoom instances by lobby ID.
//
//   - GameRoom
//     Represents a single running game session. Each room owns exactly one
//     GameState and a set of connected players. All game logic is executed
//     serially inside the room to guarantee consistency.
//
//   - GameState
//     A pure domain model that applies Commands and produces Events. It contains
//     no networking, logging, or metrics logic. GameState is deterministic and
//     versioned to support client synchronization.
//
//   - PlayerConn
//     Represents a single player WebSocket connection. It handles encoding and
//     decoding of commands and events, but does not contain game logic.
//
// Data flow:
//
//	Client ──Command──▶ PlayerConn ──▶ GameRoom ──▶ GameState
//	                               ▲               │
//	                               └─── Event ◀────┘
//
// Commands:
//
// Commands represent player intents (e.g. play a card, draw a card). They are
// validated and applied by the GameState. Invalid commands result in explicit
// rejection events sent back to the issuing player.
//
// Events:
//
// Events represent immutable facts that occurred in the game. After each
// successful command, the GameState produces one or more events which are then
// broadcast to all connected players. Clients rebuild their local state solely
// from events and snapshots.
//
// Concurrency model:
//
// Each GameRoom runs as a single logical owner of its state. All commands,
// joins, leaves, and broadcasts are funneled through the room, ensuring
// sequential execution without requiring fine-grained locking.
//
// Observability:
//
// The game package is instrumented via the telemetry package for Prometheus
// metrics and uses structured logging (slog) at integration boundaries. Domain
// logic remains free of side effects.
//
// This package is framework-agnostic and does not depend on Echo, HTTP, or
// WebSocket implementations. Transport and authentication concerns are handled
// by higher layers.
package game
