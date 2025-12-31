package lobbies

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"slices"
	"strconv"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
)

var (
	_jsonLogger = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug.Level(),
	})
	logger = slog.New(_jsonLogger).WithGroup("Lobby")
)

type FakeLobbyRepo struct{}

// Find implements [LobbyRepository].
func (fpr FakeLobbyRepo) Find(context.Context, string) (*Lobby, error) {
	panic("unimplemented")
}

var envMaxPlayerInLobby = os.Getenv("MUNCHIN_MAX_PLAYER_IN_LOBBY")

// StartGame implements [LobbyRepository].
func (fpr FakeLobbyRepo) StartGame(ctx context.Context, lobbyID string) error {
	if lobbyID == "bad_uuid" {
		return errors.New("cannot start game with bad uuid")
	}
	return nil
}

// AddPlayer implements [LobbyRepository].
func (fpr FakeLobbyRepo) AddPlayer(ctx context.Context, l *Lobby, p *auth.Player) error {
	maxPlayerInLobby, err := strconv.Atoi(envMaxPlayerInLobby)
	if err != nil {
		return err
	}

	if slices.Contains(l.Players, *p) {
		return ErrPlayerAlreadyInLobby
	} else if len(l.Players) > maxPlayerInLobby {
		return ErrFullLobby
	}

	l.Players = append(l.Players, *p)
	logger.With(slog.String("playerAdded", p.PlayerID)).
		With(slog.Int("lobbySize", len(l.Players))).
		DebugContext(ctx, "added new player to lobby")
	return nil
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
func (fpr FakeLobbyRepo) FinishGame(ctx context.Context, l *Lobby) error {
	panic("unimplemented")
}

func NewFakeLobbyRepo() LobbyRepository {
	return FakeLobbyRepo{}
}
