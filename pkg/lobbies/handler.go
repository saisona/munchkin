package lobbies

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/game"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Handler struct {
	s  *Service
	gh game.GameHub
}

var (
	ErrMissingLobby         = errors.New("parameter lobby not provided")
	ErrUnknownLobby         = errors.New("requested lobby not found or game already started")
	ErrPlayerAlreadyInLobby = errors.New("cannot join an already joined game")
	ErrLobbyAlreadyStarted  = errors.New("game is already started")
	ErrFullLobby            = errors.New("cannot join a full lobby")
)

func mapRegisterError(err error) error {
	switch {
	case errors.Is(err, ErrFullLobby):
		return echo.NewHTTPError(http.StatusLocked, ErrFullLobby)
	case errors.Is(err, ErrPlayerAlreadyInLobby):
		return echo.NewHTTPError(http.StatusNotModified, ErrPlayerAlreadyInLobby)
	case errors.Is(err, gorm.ErrRecordNotFound):
		return echo.NewHTTPError(http.StatusNotFound, ErrUnknownLobby)
	default:
		return echo.ErrInternalServerError
	}
}

func NewLobbyHandler(svc *Service, gh game.GameHub) Handler {
	return Handler{s: svc, gh: gh}
}

// HandleNewLobby godoc
// @Summary Create lobby
// @Description Creates a new lobby for the authenticated player and initializes the in-memory game room.
// @Security BearerAuth
// @Tags lobby
// @Produce json
// @Success 201 {object} LobbyCreationResponse "Lobby created"
// @Failure 400 {object} LobbyCreationResponse "Invalid request"
// @Failure 500 {string} string "Internal server error"
// @Router /lobby [post]
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
	gameStateID := fmt.Sprintf("gs-%s", playerID)

	logger.With(slog.String("id", lobbyID), slog.String("gameID", gameStateID)).DebugContext(c.Request().Context(), "lobby created")
	// NOTE: Should the GameState be given by the handler ? I think not .. but.... who knows
	// TODO: better handle creation of the playerDTO when creating the ROOM
	// 1. create the Player DTO
	// 2. pre-load or not his hand
	_, errCreateRoom := h.gh.CreateRoom(lobbyID, game.NewGameState(gameStateID, []*game.Player{
		game.NewPlayer(playerID, playerID),
	}))
	if errCreateRoom != nil {
		return errCreateRoom
	}
	telemetry.LobbyCreatedTotal.Inc()
	lcr := LobbyCreationResponse{
		LobbyID: lobbyID,
		Error:   "",
	}
	return c.JSON(http.StatusCreated, lcr)
}

// HandleStartGame godoc
// @Summary Start lobby game
// @Security BearerAuth
// @Description Marks the lobby as started and transitions it into the in-game state.
// @Tags lobby
// @Param id path string true "Lobby ID"
// @Success 200 "Game started"
// @Failure 400 {object} LobbyCreationResponse "Missing lobby ID or invalid lobby"
// @Failure 404 {string} string "Lobby not found"
// @Failure 500 {string} string "Internal server error"
// @Router /lobby/{id}/start [post]
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
			LobbyID: lobbyID,
			Error:   errStartGame.Error(),
		}
		return c.JSON(400, lcr)
	}

	lobby, errFindLobby := h.s.repo.Find(ctx, lobbyID)
	if errFindLobby != nil {
		return mapRegisterError(errFindLobby)
	}

	players := make([]*game.Player, 0, len(lobby.Players))
	for _, player := range lobby.Players {
		players = append(players, game.NewPlayer(player.ID, player.Username))
	}

	room, ok := h.gh.GetRoom(lobbyID)
	if !ok {
		gameStateID := fmt.Sprintf("gs-%s", lobby.ID)
		createdRoom, errCreateRoom := h.gh.CreateRoom(lobbyID, game.NewGameState(gameStateID, players))
		if errCreateRoom != nil {
			return errCreateRoom
		}
		room = createdRoom
	}

	if err := room.Start(players); err != nil {
		return err
	}

	defaultCtx := context.Background()
	go h.s.StopGame(defaultCtx, lobbyID)
	telemetry.LobbyActive.Inc()
	return c.NoContent(http.StatusOK)
}

// HandleJoinGame godoc
// @Summary Join lobby
// @Security BearerAuth
// @Description Adds the authenticated player to an existing lobby before the game starts.
// @Tags lobby
// @Param id path string true "Lobby ID"
// @Success 200 {string} string "Successfully joined the lobby"
// @Failure 304 {string} string "Player already in lobby"
// @Failure 400 {string} string "Missing lobby ID or invalid request"
// @Failure 404 {string} string "Lobby not found"
// @Failure 423 {string} string "Lobby is full"
// @Failure 500 {string} string "Internal server error"
// @Router /lobby/{id}/join [post]
func (h Handler) HandleJoinGame(c echo.Context) error {
	ctx := c.Request().Context()
	lobbyID := c.Param("id")
	playerID := c.Get("playerID").(string)
	if lobbyID == "" {
		logger.With(slog.String("playerID", playerID)).ErrorContext(ctx, "missing param :id for the lobby in HandleJoinGame")
		return ErrMissingLobby
	}

	logger.With(slog.String("id", lobbyID), slog.String("playerID", playerID)).DebugContext(ctx, "joining game")
	if errJoinGame := h.s.JoinGame(ctx, lobbyID, playerID); errJoinGame != nil {
		logger.With(slog.String("error", errJoinGame.Error())).ErrorContext(ctx, "joining game failed")

		return mapRegisterError(errJoinGame)
	}
	telemetry.GameRoomJoins.Inc()
	return nil
}

// GetAllLobbies godoc
// @Summary List all lobbies
// @Security BearerAuth
// @Description Returns the full lobby collection without pagination. Intended mainly for administration or debugging.
// @Tags lobby
// @Produce json
// @Success 200 {array} LobbyListItem "List of lobbies"
// @Failure 500 {string} string "Internal server error"
// @Router /lobby/model [get]
func (h Handler) GetAllLobbies(c echo.Context) error {
	lobbies, err := h.s.repo.Fetch(c.Request().Context())
	if err != nil {
		return mapRegisterError(err)
	}
	return c.JSON(200, lobbies)
}

// DeleteLobby godoc
// @Summary Delete lobby
// @Security BearerAuth
// @Description Deletes a lobby by identifier.
// @Tags lobby
// @Produce json
// @Param id path string true "Lobby ID"
// @Success 204 "When deleting a lobby, a 204 NoContent is received"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Lobby not found"
// @Router /lobby/{id} [delete]
func (h Handler) DeleteLobby(c echo.Context) error {
	if err := h.s.repo.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return mapRegisterError(err)
	}
	return c.NoContent(204)
}

// LobbyListResponse represents a paginated lobby list response.
type LobbyListResponse struct {
	// List of lobby items.
	Items []LobbyListItem `json:"items"`

	// Maximum number of items returned.
	// example: 20
	Limit int `json:"limit"`

	// Offset used for pagination.
	// example: 0
	Offset int `json:"offset"`

	// Indicates whether more items are available.
	// example: true
	HasMore bool `json:"hasMore"`
}

// ListLobbies godoc
// @Summary List lobbies
// @Security BearerAuth
// @Description Retrieves a paginated lobby list for the lobby selection screen.
// @Tags lobby
// @Produce json
// @Param limit query int false "Maximum number of items to return" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} LobbyListResponse "Paginated list of lobbies"
// @Failure 500 {string} string "Internal server error"
// @Router /lobby [get]
func (h *Handler) ListLobbies(c echo.Context) error {
	ctx := c.Request().Context()

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 20
	}

	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// fetch limit + 1 to detect "has more"
	items, err := h.s.repo.ListForLobbyScene(ctx, limit+1, offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	hasMore := false
	if len(items) > limit {
		hasMore = true
		items = items[:limit]
	}

	resp := LobbyListResponse{
		Items:   items,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}

	return c.JSON(http.StatusOK, resp)
}
