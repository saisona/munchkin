package lobbies

import (
	"context"
	"log/slog"
	"slices"
	"strconv"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"gorm.io/gorm"
)

type DBLobbyRepo struct {
	db *gorm.DB
}

// Find implements [LobbyRepository].
func (dlr DBLobbyRepo) Find(ctx context.Context, lobbyID string) (*Lobby, error) {
	var lobby Lobby
	tx := dlr.db.WithContext(ctx).Debug().First(&lobby, "LobbyID = ?", lobbyID)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &lobby, tx.Error
}

// StartGame implements [LobbyRepository].
func (dlr DBLobbyRepo) StartGame(ctx context.Context, lobbyID string) error {
	// dlr.db.Model(&Lobby{}).Debug().WithContext(ctx).First()
	return nil
}

// AddPlayer implements [LobbyRepository].
func (dlr DBLobbyRepo) AddPlayer(ctx context.Context, l *Lobby, p *auth.Player) error {
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
	if errAppendPlayer := dlr.db.WithContext(ctx).Model(l).Association("Players").Append(p); errAppendPlayer != nil {
		return errAppendPlayer
	}
	logger.With(slog.String("playerAdded", p.PlayerID)).
		With(slog.Int("lobbySize", len(l.Players))).
		DebugContext(ctx, "added new player to lobby")
	return nil
}

// Create implements [LobbyRepository].
func (dlr DBLobbyRepo) Create(ctx context.Context, l *Lobby) error {
	logger.DebugContext(ctx, "Launched create with successs")
	if len(l.Players) == 0 {
		return ErrMissingLobby
	}
	return nil
}

// FinishGame implements [LobbyRepository].
func (dlr DBLobbyRepo) FinishGame(ctx context.Context, l *Lobby) error {
	panic("unimplemented")
}

func NewDBLobbyRepo(db *gorm.DB) (LobbyRepository, error) {
	if errAutoMigrate := db.AutoMigrate(&Lobby{}); errAutoMigrate != nil {
		return nil, errAutoMigrate
	}
	return DBLobbyRepo{db: db}, nil
}
