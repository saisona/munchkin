package lobbies

import (
	"context"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
)

type LobbyCreationResponse struct {
	LobbyID string `json:"lobby_id"`
	Error   string `json:"error,omitempty"`
}

type LobbyRepository interface {
	Create(context.Context, *Lobby) error
	Find(context.Context, string) (*Lobby, error)
	FinishGame(context.Context, string) error
	StartGame(context.Context, string) error
	AddPlayer(context.Context, string, *auth.Player) error
}
