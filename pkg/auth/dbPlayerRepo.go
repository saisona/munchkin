package auth

import (
	"context"
	"errors"
	"log/slog"

	"gorm.io/gorm"
)

type DBPlayerRepo struct {
	db *gorm.DB
}

func NewDBPlayerRepo(dbConn *gorm.DB) (PlayerRepository, error) {
	if errAutoMigrate := dbConn.AutoMigrate(&Player{}); errAutoMigrate != nil {
		return nil, errAutoMigrate
	}
	return &DBPlayerRepo{db: dbConn}, nil
}

var ErrCreatePlayerInDB = errors.New("failed creation of users in DB")

func (pr *DBPlayerRepo) Create(ctx context.Context, p *Player) error {
	if fi := pr.db.WithContext(ctx).Model(&Player{}).Create(&p); fi.Error != nil {
		return fi.Error
	}
	return nil
}

func (pr *DBPlayerRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	p, err := pr.FindByUsername(ctx, username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	return p != nil, nil
}

func (pr *DBPlayerRepo) FindByUsername(ctx context.Context, username string) (*Player, error) {
	logger.With(slog.String("searched username", username)).DebugContext(ctx, "called FindByUsername")
	var foundPlayer Player
	if err := pr.db.Debug().WithContext(ctx).Model(&Player{}).Where("username = ?", username).First(&foundPlayer).Error; err != nil {
		return nil, err
	}
	return &foundPlayer, nil
}

// FindByID implements [PlayerRepository].
func (pr *DBPlayerRepo) FindByID(context.Context, string) (*Player, error) {
	panic("unimplemented")
}
