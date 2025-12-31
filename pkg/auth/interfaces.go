package auth

import "context"

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type PlayerRepository interface {
	Create(context.Context, *Player) error
	UsernameExists(context.Context, string) (bool, error)
	FindByUsername(context.Context, string) (*Player, error)
	FindByID(context.Context, string) (*Player, error)
}

type TokenIssuer interface {
	Issue(playerID string, key []byte) (string, error)
}

type Service struct {
	repo   PlayerRepository
	ti     TokenIssuer
	jwtKey []byte
}
