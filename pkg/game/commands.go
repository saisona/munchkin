package game

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	MessageTypePlayerAction MessageType = "PLAYER_ACTION"
)

type PlayerActionType string

const (
	ActionOpenDoor       PlayerActionType = "OPEN_DOOR"
	ActionLookForTrouble PlayerActionType = "LOOK_FOR_TROUBLE"
	ActionLootRoom       PlayerActionType = "LOOT_ROOM"
	ActionEndTurn        PlayerActionType = "END_TURN"
)

// CommandEnvelope is the generic WebSocket message envelope.
type CommandEnvelope struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

// PlayerActionPayload is the WebSocket payload for a turn action.
type PlayerActionPayload struct {
	Action    PlayerActionType `json:"action"`
	Timestamp time.Time        `json:"timestamp"`
}

// PlayerActionCommand is the decoded server-side command for turn actions.
type PlayerActionCommand struct {
	PlayerID  string
	Action    PlayerActionType
	Timestamp time.Time
}

func (c PlayerActionCommand) GetPlayerID() string {
	return c.PlayerID
}

func (c PlayerActionCommand) Type() string {
	return string(MessageTypePlayerAction)
}

// PlayCardCommand models a future explicit play-card action.
type PlayCardCommand struct {
	PlayerID string
	CardID   string
}

func (c PlayCardCommand) GetPlayerID() string {
	return c.PlayerID
}

func (c PlayCardCommand) Type() string {
	return "PLAY_CARD"
}

func (c PlayCardCommand) GetCardID() string {
	return c.CardID
}

// DrawCardCommand models a future explicit draw-card action.
type DrawCardCommand struct {
	PlayerID string
}

func (c DrawCardCommand) GetPlayerID() string {
	return c.PlayerID
}

func (c DrawCardCommand) Type() string {
	return "DRAW_CARD"
}
