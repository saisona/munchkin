package game

import (
	"log/slog"
	"os"

	"github.com/gorilla/websocket"
)

var (
	_jsonLogger = slog.NewJSONHandler(os.Stdout, nil)
	logger      = slog.New(_jsonLogger).WithGroup("game")
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
	defer p.Conn.Close()
	for evt := range p.Send {
		if err := p.Conn.WriteJSON(evt); err != nil {
			logger.With(
				slog.String("component", "ws"),
				slog.String("playerID", p.PlayerID),
				slog.String("lobbyID", p.Room.lobbyID),
			).Warn("write to WS failed", "error", err)
			break
		}
	}
}

func (p *PlayerConn) ReadLoop(room *GameRoom) {
	defer p.Conn.Close()
	for {
		var cmd Command
		if err := p.Conn.ReadJSON(&cmd); err != nil {
			logger.With(
				slog.String("component", "ws"),
				slog.String("playerID", p.PlayerID),
				slog.String("lobbyID", p.Room.lobbyID),
			).Warn("read from WS failed", "error", err)
			// room.Leave(p)
			break
		}
		// room.HandleCommand(cmd)
	}
}
