package lobbies

import (
	"context"
	"log/slog"
	"slices"
	"strconv"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBLobbyRepo struct {
	db *gorm.DB
}

// Find implements [LobbyRepository].
func (dlr DBLobbyRepo) Find(ctx context.Context, lobbyID string) (*Lobby, error) {
	var l Lobby
	tx := dlr.db.WithContext(ctx).First(&l, "ID = ?", lobbyID)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &l, tx.Error
}

// StartGame implements [LobbyRepository].
func (dlr DBLobbyRepo) StartGame(ctx context.Context, lobbyID string) error {
	tx := dlr.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var l Lobby
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", lobbyID).
		First(&l).Error; err != nil {

		tx.Rollback()
		return err
	}

	slog.InfoContext(ctx, "found lobby",
		slog.Any("id", l.ID),
		slog.Any("state", l.State),
	)

	switch l.State {
	case StateFull:
		tx.Rollback()
		return ErrFullLobby

	case StateInGame:
		tx.Rollback()
		return ErrLobbyAlreadyStarted

	case StateAvailable:
		res := tx.
			Model(&l).
			Select("State").
			Updates(Lobby{State: StateInGame})

		if res.Error != nil {
			tx.Rollback()
			return res.Error
		}
	}

	return tx.Commit().Error
}

// AddPlayer implements [LobbyRepository].
func (dlr DBLobbyRepo) AddPlayer(ctx context.Context, lobbyID string, p *auth.Player) error {
	l, errFindLobby := dlr.Find(ctx, lobbyID)
	if errFindLobby != nil {
		return errFindLobby
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
	if errAppendPlayer := dlr.db.WithContext(ctx).Model(l).Association("Players").Append(p); errAppendPlayer != nil {
		return errAppendPlayer
	}
	logger.With(slog.String("playerAdded", p.ID)).
		With(slog.Int("lobbySize", len(l.Players))).
		DebugContext(ctx, "added new player to lobby")
	return nil
}

// Create implements [LobbyRepository].
func (dlr DBLobbyRepo) Create(ctx context.Context, l *Lobby) error {
	logger.DebugContext(ctx, "Launched create with successs")
	l.State = StateAvailable
	if len(l.Players) == 0 {
		return ErrMissingLobby
	} else if errCreateLobby := dlr.db.WithContext(ctx).Model(l).Create(&l).Error; errCreateLobby != nil {
		return errCreateLobby
	}
	return nil
}

// FinishGame implements [LobbyRepository].
func (dlr DBLobbyRepo) FinishGame(ctx context.Context, lobbyID string) error {
	panic("unimplemented")
}

func NewDBLobbyRepo(db *gorm.DB) (LobbyRepository, error) {
	if errAutoMigrate := db.AutoMigrate(Lobby{}); errAutoMigrate != nil {
		return nil, errAutoMigrate
	}
	return DBLobbyRepo{db: db}, nil
}
