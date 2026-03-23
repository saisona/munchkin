package lobbies

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/game"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024, WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // tighten later
	},
}

// GameWS godoc
//
// @Summary Connect lobby WebSocket
// @Description Upgrades the HTTP connection to WebSocket for real-time lobby and game events.
// @Tags lobby
// @Param id path string true "Lobby ID"
// @Success 101 {string} string "Switching Protocols"
// @Failure 404 {string} string "Lobby not found"
// @Failure 500 {string} string "Internal server error"
// @Router /lobby/{id}/ws [get]
// @Security BearerAuth
func (h *Handler) GameWS(c echo.Context) error {
	lobbyID := c.Param("id")
	playerID := c.Get("playerID").(string)
	logger.With(slog.String("lobbyID", lobbyID)).Info("GameWS called")

	room, ok := h.gh.GetRoom(lobbyID)
	if !ok {
		l, err := h.s.repo.Find(c.Request().Context(), lobbyID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUnknownLobby
		}
		if err != nil {
			return err
		}

		players := make([]*game.Player, 0, len(l.Players))
		for _, lobbyPlayer := range l.Players {
			players = append(players, game.NewPlayer(lobbyPlayer.ID, lobbyPlayer.Username))
		}

		gameStateName := fmt.Sprintf("gs-%s", l.ID)
		gm, err := h.gh.CreateRoom(lobbyID, game.NewGameState(gameStateName, players))
		if err != nil {
			return err
		}
		room = gm
	}

	conn, err := upgrader.Upgrade(
		c.Response(),
		c.Request(),
		nil,
	)
	if err != nil {
		telemetry.WSUpgradeFailures.With(prometheus.Labels{"reason": err.Error()}).Inc()
		return nil
	}

	player := &game.PlayerConn{
		PlayerID: playerID,
		Conn:     conn,
		Send:     make(chan game.Event, 16),
		Room:     room,
	}

	room.Join(player)

	go player.WriteLoop()
	go player.ReadLoop(room)

	return nil
}
