package game

import (
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
	close := p.Conn.CloseHandler()
	defer close(200, "normal CLOSE")
	for {
		var cmd Command
		msgType, data, err := p.Conn.ReadMessage()
		if err != nil {
			logger.With(slog.String("error", err.Error())).Error("p.Conn.ReadMessage")
		}
		logger.With(slog.Int("msgType", msgType), slog.String("data", string(data))).
			Debug("got message from WS")
		if err := p.Conn.ReadJSON(&cmd); err != nil {
			logger.With(
				slog.String("playerID", p.PlayerID),
				slog.String("lobbyID", p.Room.lobbyID),
				slog.String("error", err.Error()),
			).Error("read from WS failed")
			break
		}
		room.handleCommand(cmd)
	}
}
