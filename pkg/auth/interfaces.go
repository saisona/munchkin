package auth

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// AuthRequest represents authentication credentials.
type AuthRequest struct {
	// Username of the user.
	// example: alice
	Username string `json:"username"`

	// Password of the user.
	// example: strongpassword123
	Password string `json:"password"`
}

// AuthResponse represents a successful authentication response.
type AuthResponse struct {
	// JWT access token.
	// example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
	Token string `json:"token"`
}

type PlayerRepository interface {
	Create(context.Context, *Player) error
	UsernameExists(context.Context, string) (bool, error)
	FindByUsername(context.Context, string) (*Player, error)
	FindByID(context.Context, string) (*Player, error)
}

type TokenIssuer interface {
	Issue(string, []byte) (string, error)
}

type Service struct {
	repo   PlayerRepository
	tracer trace.Tracer
	ti     TokenIssuer
	jwtKey []byte
}
