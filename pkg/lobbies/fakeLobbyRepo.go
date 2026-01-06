package lobbies

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"time"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/log"
)

var (
	_jsonLogger = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug.Level(),
		AddSource: true,
	})

	logger = slog.New(log.NewOTelHandler(_jsonLogger)).WithGroup("Lobby")
)

var _baseFakeDate = []Lobby{
	{Players: make([]*auth.Player, 0), ID: "fake_id", State: StateAvailable, CreatedAt: time.Now()},
	{
		Players: []*auth.Player{
			{
				ID:           "p_fake_id",
				Username:     "admin",
				CreatedAt:    time.Now(),
				PasswordHash: "fake_hash",
			},
			{
				ID:           "p_fake_id_1",
				Username:     "admin_1",
				CreatedAt:    time.Now(),
				PasswordHash: "fake_hash",
			},
		},
		ID:        "fake_id",
		State:     StateAvailable,
		CreatedAt: time.Now(),
	},
	{Players: make([]*auth.Player, 0), ID: "fake_id", State: StateAvailable, CreatedAt: time.Now()},
}

type FakeLobbyRepo struct {
	fakeDate []Lobby
}

// Create implements [LobbyRepository].
func (fpr FakeLobbyRepo) Create(ctx context.Context, l *Lobby) error {
	logger.DebugContext(ctx, "Launched create with successs")
	if len(l.Players) == 0 {
		return errors.New("no players in the lobby")
	}
	return nil
}

// FinishGame implements [LobbyRepository].
func (fpr FakeLobbyRepo) FinishGame(context.Context, string) error {
	panic("unimplemented")
}

// Find implements [LobbyRepository].
func (fpr FakeLobbyRepo) Find(context.Context, string) (*Lobby, error) {
	return &fpr.fakeDate[0], nil
}

var envMaxPlayerInLobby = os.Getenv("MUNCHIN_MAX_PLAYER_IN_LOBBY")

// StartGame implements [LobbyRepository].
func (fpr FakeLobbyRepo) StartGame(ctx context.Context, lobbyID string) error {
	if lobbyID == "bad_uuid" {
		return ErrUnknownLobby
	}

	return nil
}

// AddPlayer implements [LobbyRepository].
func (fpr FakeLobbyRepo) AddPlayer(ctx context.Context, lobbyID string, p *auth.Player) error {
	l, errFind := fpr.Find(ctx, lobbyID)
	if errFind != nil {
		return errFind
	}
	maxPlayerInLobby, err := strconv.Atoi(envMaxPlayerInLobby)
	if err != nil {
		return err
	}

	if slices.Contains(l.Players, p) {
		return ErrPlayerAlreadyInLobby
	} else if len(l.Players) > maxPlayerInLobby {
		return ErrFullLobby
	}

	l.Players = append(l.Players, p)
	logger.With(slog.String("playerAdded", p.ID)).
		With(slog.Int("lobbySize", len(l.Players))).
		DebugContext(ctx, "added new player to lobby")
	return nil
}

func NewFakeLobbyRepo() LobbyRepository {
	return FakeLobbyRepo{fakeDate: _baseFakeDate}
}
