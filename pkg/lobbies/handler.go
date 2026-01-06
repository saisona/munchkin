package lobbies

import (
	"errors"
	"log/slog"
	"net/http"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler struct {
	s *Service
}

var (
	ErrMissingLobby         = errors.New("parameter lobby not provided")
	ErrUnknownLobby         = errors.New("requested lobby not found or game already started")
	ErrPlayerAlreadyInLobby = errors.New("cannot join an already joined game")
	ErrLobbyAlreadyStarted  = errors.New("game is already started.")
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
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		lcr := LobbyCreationResponse{
			LobbyID: "",
			Error:   ErrUnknownLobby.Error(),
		}
		return c.JSON(400, lcr)
	}

	logger.With(slog.String("id", lobbyID)).DebugContext(c.Request().Context(), "lobby created")
	telemetry.LobbyCreatedTotal.Inc()
	lcr := LobbyCreationResponse{
		LobbyID: lobbyID,
		Error:   "",
	}
	return c.JSON(http.StatusCreated, lcr)
}

func (h Handler) HandleStartGame(c echo.Context) error {
	lobbyID := c.Param("id")
	if lobbyID == "" {
		lcr := LobbyCreationResponse{
			LobbyID: "",
			Error:   ErrMissingLobby.Error(),
		}
		return c.JSON(400, lcr)
	}
	ctx := c.Request().Context()
	logger.With(slog.String("id", lobbyID)).DebugContext(ctx, "start game requested")
	if errStartGame := h.s.StartGame(ctx, lobbyID); errStartGame != nil {
		lcr := LobbyCreationResponse{
			LobbyID: "",
			Error:   errStartGame.Error(),
		}
		return c.JSON(400, lcr)
	}
	telemetry.LobbyActive.Inc()

	// TODO: handle WebSocket
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
