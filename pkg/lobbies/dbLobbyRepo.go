package lobbies

import (
	"context"
	"log/slog"
	"slices"
	"strconv"
	"time"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBLobbyRepo struct {
	db     *gorm.DB
	tracer trace.Tracer
}

// Delete implements [LobbyRepository].
func (dlr DBLobbyRepo) Delete(ctx context.Context, lobbyID string) error {
	ctxTracer, span := dlr.tracer.Start(ctx, "repo.deleteLobby")
	span.SetAttributes(attribute.String("lobbyID", lobbyID))
	defer span.End()

	if err := dlr.db.Debug().WithContext(ctxTracer).First(&Lobby{ID: lobbyID}).Error; err != nil {
		return err
	} else {
		return dlr.db.WithContext(ctxTracer).Delete(&Lobby{ID: lobbyID}).Error
	}
}

// Fetch implements [LobbyRepository].
func (dlr DBLobbyRepo) Fetch(ctx context.Context) ([]Lobby, error) {
	ctxTracer, span := dlr.tracer.Start(ctx, "repo.fetchLobbies")
	defer span.End()

	lobbies := make([]Lobby, 0, 10)
	if tx := dlr.db.WithContext(ctxTracer).Preload("Players").Limit(10).Find(&lobbies, "state != ?", StateInGame); tx.Error != nil {
		return nil, tx.Error
	}

	return lobbies, nil
}

// Find implements [LobbyRepository].
func (dlr DBLobbyRepo) Find(ctx context.Context, lobbyID string) (*Lobby, error) {
	ctxTracer, span := dlr.tracer.Start(ctx, "repo.find")
	var l Lobby
	tx := dlr.db.WithContext(ctxTracer).First(&l, "ID = ?", lobbyID)
	if tx.Error != nil {
		span.RecordError(tx.Error, trace.WithStackTrace(true))
		return nil, tx.Error
	}
	return &l, tx.Error
}

// StartGame implements [LobbyRepository].
func (dlr DBLobbyRepo) StartGame(ctx context.Context, lobbyID string) error {
	ctxTracer, span := dlr.tracer.Start(ctx, "repo.find")
	tx := dlr.db.WithContext(ctxTracer).Begin()
	defer func() {
		if r := recover(); r != nil {
			span.RecordError(r.(error))
			tx.Rollback()
		}
	}()

	var l Lobby
	if err := tx.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", lobbyID).
		First(&l).Error; err != nil {

		tx.Rollback()
		span.RecordError(err)
		return err
	}

	slog.InfoContext(ctx, "found lobby",
		slog.Any("id", l.ID),
		slog.Any("state", l.State),
	)

	switch l.State {
	case StateFull:
		tx.Rollback()
		span.RecordError(ErrFullLobby)
		return ErrFullLobby

	case StateInGame:
		tx.Rollback()
		span.RecordError(ErrLobbyAlreadyStarted)
		return ErrLobbyAlreadyStarted

	case StateAvailable:
		res := tx.
			Model(&l).
			Select("State").
			Updates(Lobby{State: StateInGame})

		if res.Error != nil {
			tx.Rollback()
			span.RecordError(res.Error)
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
	traceCtx, span := dlr.tracer.Start(ctx, "game.finish")
	defer span.End()
	logger.With(slog.String("lobbyID", lobbyID)).InfoContext(traceCtx, "finishing the game as should never take more than 1day")
	tx := dlr.db.WithContext(traceCtx).Debug().Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ?", lobbyID).
		UpdateColumn("FinishedAt ", time.Now())
	if tx.Error != nil {
		tx.Rollback()
		span.RecordError(tx.Error)
		return tx.Error
	}
	return nil
}

func (dlr DBLobbyRepo) ListForLobbyScene(
	ctx context.Context,
	limit int,
	offset int,
) ([]LobbyListItem, error) {
	items := make([]LobbyListItem, 0, limit)
	traceCtx, span := dlr.tracer.Start(ctx, "lobby.game_list")
	defer span.End()

	tx := dlr.db.
		WithContext(traceCtx).
		Table("lobbies AS l").
		Select(`
			l.id,
			l.state,
			COUNT(lp.player_id) AS player_count
		`).
		Joins(`
			LEFT JOIN lobby_players lp
			ON lp.lobby_id = l.id
		`).
		Where("l.state != ?", StateInGame).
		Group("l.id").
		Order("l.created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&items)

	if tx.Error != nil {
		span.RecordError(tx.Error)
		return nil, tx.Error
	}

	return items, nil
}

func NewDBLobbyRepo(db *gorm.DB) (LobbyRepository, error) {
	if errAutoMigrate := db.AutoMigrate(Lobby{}); errAutoMigrate != nil {
		return nil, errAutoMigrate
	}
	return DBLobbyRepo{db: db, tracer: telemetry.DefaultRepoTracer}, nil
}
