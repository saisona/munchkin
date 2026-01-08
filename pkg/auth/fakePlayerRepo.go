package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"golang.org/x/crypto/bcrypt"
)

type FakePlayerRepo struct{}

var ErrPlayerIsNil = errors.New("given argument player is nil, something went wrong")

var logger = slog.New(otelslog.NewHandler("munchin")).WithGroup("Auth")

func (fpr *FakePlayerRepo) Create(ctx context.Context, p *Player) error {
	if p == nil {
		logger.WarnContext(ctx, "given argument player is nil")
		return ErrPlayerIsNil
	}
	logger.DebugContext(ctx, "Launched create with successs")
	return nil
}

func (fpr *FakePlayerRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	return strings.Compare(username, "badUsername") == 0, nil
}

func (fpr *FakePlayerRepo) FindByUsername(ctx context.Context, username string) (*Player, error) {
	if username != "admin" {
		return nil, fmt.Errorf("didn't found player called %s", username)
	}

	hash, _ := bcrypt.GenerateFromPassword(
		[]byte("fake_hash"),
		bcrypt.DefaultCost,
	)
	return &Player{
		ID:           "fake_id",
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}, nil
}
