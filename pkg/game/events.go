package game

import "time"

type EventType string

type BaseEvent struct {
	Type      EventType `json:"type"`
	GameID    string    `json:"gameId"`
	Timestamp time.Time `json:"ts"`
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
	CommandType string
	Reason      string
}
type PlayerJoinedEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
}

type PlayerLeftEvent struct {
	BaseEvent
	PlayerID string
}

type CardPlayedEvent struct {
	BaseEvent
	PlayerID string
	CardID   string
}

type CardDrawnEvent struct {
	BaseEvent
	PlayerID string
	CardID   string
}
