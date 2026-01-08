package lobbies

import (
	"net/http"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/game"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // tighten later
	},
}

func (h *Handler) GameWS(c echo.Context) error {
	lobbyID := c.Param("lobbyID")
	playerID := c.Get("playerID").(string)

	conn, err := upgrader.Upgrade(
		c.Response(),
		c.Request(),
		nil,
	)
	if err != nil {
		telemetry.WSUpgradeFailures.With(prometheus.Labels{"reason": err.Error()}).Inc()
		return err
	}

	room, ok := h.gh.GetRoom(lobbyID)
	if !ok {
		conn.Close()
		return nil
	}

	player := &game.PlayerConn{
		PlayerID: playerID,
		Conn:     conn,
		Send:     make(chan game.Event, 16),
	}

	room.Join(player)

	// go player.WriteLoop()
	// go player.ReadLoop(room)

	return nil
}
