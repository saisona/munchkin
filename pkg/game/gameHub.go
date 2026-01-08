package game

import (
	"errors"
	"log/slog"
	"sync"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
)

var ErrRoomAlreadyExists = errors.New("game room already exists")

type AppGameHub struct {
	mu    sync.RWMutex
	rooms map[string]*GameRoom
}

func NewGameHub() GameHub {
	return &AppGameHub{
		rooms: make(map[string]*GameRoom),
		mu:    sync.RWMutex{},
	}
}

func (h *AppGameHub) GetRoom(lobbyID string) (*GameRoom, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	room, ok := h.rooms[lobbyID]
	return room, ok
}

func (h *AppGameHub) CreateRoom(lobbyID string, initialState *GameState) (*GameRoom, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.rooms[lobbyID]; exists {
		return nil, ErrRoomAlreadyExists
	}

	room := NewGameRoom(lobbyID, initialState)

	h.rooms[lobbyID] = room

	go room.Run()

	telemetry.GameRoomsCreated.Inc()

	slog.With(slog.String("component", "game_hub"), slog.String("lobbyID", lobbyID)).
		Info("game room created")
	return room, nil
}

func (h *AppGameHub) RemoveRoom(lobbyID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[lobbyID]; !ok {
		return
	}

	delete(h.rooms, lobbyID)

	telemetry.GameRoomsDestroyed.Inc()

	slog.With(slog.String("component", "game_hub"), slog.String("lobbyID", lobbyID)).
		Info("game room removed")
}
