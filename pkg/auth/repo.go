package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUsernameTaken      = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("wrong credentials")
)

func NewService(r PlayerRepository, ti TokenIssuer, jwtKey []byte) *Service {
	return &Service{repo: r, ti: ti, jwtKey: jwtKey, tracer: telemetry.DefaultSvcTracer}
}

func (s *Service) Register(
	ctx context.Context,
	username, password string,
) (string, error) {
	ctxSp, sp := s.tracer.Start(ctx, "service.register")
	defer sp.End()

	// 1. Check uniqueness
	exists, err := s.repo.UsernameExists(ctxSp, username)
	if err != nil {
		sp.RecordError(err)
		logger.With(slog.String("error", err.Error())).ErrorContext(ctxSp, "err on s.repo.UsernameExists")
		return "", err
	}
	if exists {
		telemetry.AuthFailures.WithLabelValues(ErrUsernameTaken.Error()).Inc()
		return "", ErrUsernameTaken
	}

	// 2. Hash password
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		sp.RecordError(err, trace.WithStackTrace(true))
		return "", err
	}

	// 3. Create player
	player := &Player{
		ID:           uuid.NewString(),
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
	}
	logger.With(slog.String("playerID", player.ID)).DebugContext(ctxSp, "player instance creation")

	if errRepoCreate := s.repo.Create(ctxSp, player); errRepoCreate != nil {
		sp.RecordError(err)
		return "", errRepoCreate
	}

	// 4. Issue JWT
	token, err := s.ti.Issue(player.ID, s.jwtKey)
	if err != nil {
		sp.RecordError(err)
		return "", err
	}

	telemetry.AuthSuccess.Inc()
	return token, nil
}

// Login authenticates a player using their email and password.
//
// The method verifies the provided credentials against the stored
// password hash and, on success, issues a new authentication token.
//
// Security notes:
//   - The same error is returned for unknown emails and invalid passwords
//     to avoid leaking account existence.
//   - Password verification uses constant-time bcrypt comparison.
//   - Token generation is delegated to the configured TokenIssuer.
//
// On success, Login returns a signed authentication token and the
// authenticated player's ID.
//
// Possible errors:
//   - ErrInvalidCredentials if authentication fails
//   - Any error returned by the underlying token issuer
//
// Login is transport-agnostic and does not perform any HTTP-specific
// behavior such as setting cookies or headers.
func (s *Service) Login(ctx context.Context, username, password string) (string, string, error) {
	ctxSp, span := s.tracer.Start(ctx, "service.login")
	defer span.End()

	span.SetAttributes(
		attribute.String("auth.username", username),
	)

	player, errRepoFind := s.repo.FindByUsername(ctxSp, username)
	if errRepoFind != nil {
		// Do NOT leak existence information
		logger.With(slog.String("request_username", username)).DebugContext(ctxSp, "FindByUsername failed")
		return "", "", ErrInvalidCredentials
	}

	// Compare bcrypt hash
	if err := bcrypt.CompareHashAndPassword(
		[]byte(player.PasswordHash),
		[]byte(password),
	); err != nil {
		return "", "", ErrInvalidCredentials
	}

	token, errIssuer := s.ti.Issue(player.ID, s.jwtKey)
	if errIssuer != nil {
		return "", "", errIssuer
	}

	return token, player.ID, nil
}
