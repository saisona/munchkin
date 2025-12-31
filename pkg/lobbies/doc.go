// Package lobbies manages pre-game player coordination and lobby lifecycle.
//
// A lobby represents a temporary, server-side construct where players
// gather before starting a game session. The lobby package is responsible
// for lobby creation, discovery, membership management, and host-driven
// state transitions.
//
// This package intentionally does not implement gameplay rules, turn
// handling, or real-time communication. Once a lobby transitions into
// an active game, responsibility is expected to move to a dedicated
// game or engine package.
//
// The lobby package is transport-agnostic and contains no HTTP-specific
// logic. Interaction with lobbies over HTTP or other protocols is handled
// by adapter layers.
package lobbies
