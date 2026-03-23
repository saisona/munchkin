package game

import "time"

type EventType string

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

type CommandRejectedEvent struct {
	BaseEvent
	CommandType string `json:"commandType"`
	Reason      string `json:"reason"`
}
type PlayerJoinedEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
}

type PlayerLeftEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
}

type CardPlayedEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
	CardID   string `json:"cardID"`
}

type CardDrawnEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
	CardID   string `json:"cardID"`
}

type TurnPhaseChangedEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
	Phase    string `json:"phase"`
}
