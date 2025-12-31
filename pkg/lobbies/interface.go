package lobbies

import (
	"context"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
)

type LobbyCreationResponse struct {
	LobbyID string `json:"lobby_id"`
	Error   error  `json:"error,omitempty"`
}

type LobbyRepository interface {
	Create(context.Context, *Lobby) error
	Find(context.Context, string) (*Lobby, error)
	FinishGame(context.Context, *Lobby) error
	StartGame(context.Context, string) error
	AddPlayer(context.Context, *Lobby, *auth.Player) error
}
