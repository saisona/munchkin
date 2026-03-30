package game

import "time"

type EventType string

// BaseEvent contains the metadata shared by all emitted game events.
type BaseEvent struct {
	Type      EventType `json:"type"`
	GameID    string    `json:"gameId,omitempty"`
	Timestamp time.Time `json:"timestamp"`
	Version   int       `json:"version"`
}

func (e BaseEvent) EventType() string {
	return string(e.Type)
}

type GameSnapshotEvent struct {
	BaseEvent
	State GameStateDTO `json:"state"`
}

// CommandRejectedEvent reports that a player command was rejected by validation.
type CommandRejectedEvent struct {
	BaseEvent
	CommandType string `json:"commandType"`
	Reason      string `json:"reason"`
}

// PlayerJoinedEvent is broadcast when a player joins a room.
type PlayerJoinedEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
}

// PlayerLeftEvent is broadcast when a player disconnects from a room.
type PlayerLeftEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
}

// CardPlayedEvent is emitted after a successful play-card command.
type CardPlayedEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
	CardID   string `json:"cardID"`
}

// CardDrawnEvent is emitted after a successful draw-card command.
type CardDrawnEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
	CardID   string `json:"cardID"`
}

// TurnPhaseChangedEvent is emitted when the active turn advances to another phase.
type TurnPhaseChangedEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
	Phase    string `json:"phase"`
}
