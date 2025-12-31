package lobbies

import (
	"errors"
	"log/slog"
	"net/http"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	s *Service
}

var (
	ErrMissingLobby         = errors.New("parameter lobby not provided")
	ErrUnknownLobby         = errors.New("requested lobby not found or game already started")
	ErrPlayerAlreadyInLobby = errors.New("cannot join an already joined game")
	ErrFullLobby            = errors.New("cannot join a full lobby")
)

// func mapRegisterError(err error) error {
// 	switch {
// 	case errors.Is(err, ErrFullLobby):
// 		return echo.NewHTTPError(
// 			http.StatusLocked,
// 			ErrFullLobby,
// 		)
// 	default:
// 		return echo.ErrInternalServerError
// 	}
// }

func NewLobbyHandler(svc *Service) Handler {
	return Handler{s: svc}
}

func (h Handler) HandleNewLobby(c echo.Context) error {
	playerID := c.Get("playerID").(string)

	lobbyID, err := h.s.CreateLobby(c.Request().Context(), playerID)
	if err != nil {
		return err
	}

	logger.With(slog.String("id", lobbyID)).DebugContext(c.Request().Context(), "lobby created")
	telemetry.LobbyCreatedTotal.Inc()
	lcr := LobbyCreationResponse{
		LobbyID: lobbyID,
		Error:   nil,
	}
	return c.JSON(http.StatusCreated, lcr)
}

func (h Handler) HandleStartGame(c echo.Context) error {
	lobbyID := c.Param("id")
	if lobbyID == "" {
		return ErrMissingLobby
	}
	logger.With(slog.String("id", lobbyID)).DebugContext(c.Request().Context(), "start game requested")
	telemetry.LobbyActive.Inc()
	return nil
}

func (h Handler) HandleJoinGame(c echo.Context) error {
	lobbyID := c.Param("id")
	if lobbyID == "" {
		return ErrMissingLobby
	}

	playerID := c.Get("playerID").(string)
	logger.With(slog.String("id", lobbyID), slog.String("playerID", playerID)).DebugContext(c.Request().Context(), "game joined")

	return nil
}
