package auth

import (
	"context"
	"errors"
	"log/slog"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type DBPlayerRepo struct {
	db     *gorm.DB
	tracer trace.Tracer
}

func NewDBPlayerRepo(dbConn *gorm.DB) (PlayerRepository, error) {
	if errAutoMigrate := dbConn.AutoMigrate(&Player{}); errAutoMigrate != nil {
		return nil, errAutoMigrate
	}
	return &DBPlayerRepo{db: dbConn, tracer: telemetry.DefaultRepoTracer}, nil
}

var ErrCreatePlayerInDB = errors.New("failed creation of users in DB")

func (pr *DBPlayerRepo) Create(ctx context.Context, p *Player) error {
	ctxSp, sp := pr.tracer.Start(ctx, "repo.create")
	defer sp.End()
	if fi := pr.db.WithContext(ctxSp).Model(&Player{}).Create(&p); fi.Error != nil {
		sp.RecordError(fi.Error, trace.WithStackTrace(true))
		return fi.Error
	}
	return nil
}

func (pr *DBPlayerRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	ctxSp, sp := pr.tracer.Start(ctx, "repo.usernameExists")
	defer sp.End()
	p, err := pr.FindByUsername(ctxSp, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		sp.RecordError(err, trace.WithStackTrace(true))
		return false, err
	}
	return p != nil, nil
}

func (pr *DBPlayerRepo) FindByUsername(ctx context.Context, username string) (*Player, error) {
	ctxSp, sp := pr.tracer.Start(ctx, "repo.FindByUsername")
	defer sp.End()
	logger.With(slog.String("searched username", username)).DebugContext(ctxSp, "called FindByUsername")
	var foundPlayer Player
	if err := pr.db.WithContext(ctxSp).Model(&Player{}).Where("username = ?", username).First(&foundPlayer).Error; err != nil {
		sp.RecordError(err, trace.WithStackTrace(true))
		return nil, err
	}
	return &foundPlayer, nil
}

// FindByID implements [PlayerRepository].
func (pr *DBPlayerRepo) FindByID(ctx context.Context, playerID string) (*Player, error) {
	ctxSp, sp := pr.tracer.Start(ctx, "repo.FindByUsername")
	defer sp.End()
	logger.With(slog.String("id", playerID)).DebugContext(ctxSp, "called FindByID")
	var foundPlayer Player
	if err := pr.db.WithContext(ctxSp).Model(&Player{}).Where("id = ?", playerID).First(&foundPlayer).Error; err != nil {
		sp.RecordError(err, trace.WithStackTrace(true))
		return nil, err
	}
	return &foundPlayer, nil
}
