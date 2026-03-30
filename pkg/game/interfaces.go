package game

import (
	"encoding/json"
	"log/slog"
	"os"

	"github.com/gorilla/websocket"
)

var (
	_jsonLogger = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger = slog.New(_jsonLogger).WithGroup("game")
)

type Command interface {
	GetPlayerID() string
	Type() string
}

type GameHub interface {
	GetRoom(lobbyID string) (*GameRoom, bool)
	CreateRoom(lobbyID string, initialState *GameState) (*GameRoom, error)
	RemoveRoom(lobbyID string)
}

type Event interface {
	EventType() string
}

type PlayerConn struct {
	PlayerID string
	Conn     *websocket.Conn
	Send     chan Event
	Room     *GameRoom
}

func (p *PlayerConn) WriteLoop() {
	logger.With(slog.String("playerID", p.PlayerID)).
		Debug("pConn started waiting for events")
	defer p.Conn.Close()
	for evt := range p.Send {
		if err := p.Conn.WriteJSON(evt); err != nil {
			logger.With(
				slog.String("playerID", p.PlayerID),
				slog.String("error", err.Error()),
			).Error("write to WS failed")
			break
		}
	}
}

func (p *PlayerConn) ReadLoop(room *GameRoom) {
	logger.With(slog.String("playerID", p.PlayerID)).
		Debug("pConn started waiting for events")
	defer p.Conn.Close()
	defer room.Leave(p.PlayerID)
	for {
		_, data, err := p.Conn.ReadMessage()
		if err != nil {
			logger.With(slog.String("error", err.Error())).Error("p.Conn.ReadMessage")
			break
		}
		logger.With(slog.String("data", string(data))).
			Debug("got message from WS")
		cmd, err := DecodeCommandMessage(p.PlayerID, data)
		if err != nil {
			logger.With(
				slog.String("playerID", p.PlayerID),
				slog.String("lobbyID", p.Room.lobbyID),
				slog.String("error", err.Error()),
			).Warn("read from WS failed")
			room.sendTo(p.PlayerID, CommandRejectedEvent{
				BaseEvent: BaseEvent{
					Type: "ERROR",
				},
				CommandType: "UNKNOWN",
				Reason:      err.Error(),
			})
			continue
		}
		room.Submit(cmd)
	}
}

func DecodeCommandMessage(playerID string, payload []byte) (Command, error) {
	var envelope CommandEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, err
	}

	switch envelope.Type {
	case MessageTypePlayerAction:
		var data PlayerActionPayload
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return nil, err
		}

		return PlayerActionCommand{
			PlayerID:  playerID,
			Action:    data.Action,
			Timestamp: data.Timestamp,
		}, nil
	default:
		return nil, ErrUnknownCommand
	}
}
