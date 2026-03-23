package lobbies

import (
	"context"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
)

type LobbyCreationResponse struct {
	// Unique lobby identifier.
	// example: 7f0f7c90-cf17-4f54-a62f-f6fbc6399d7c
	LobbyID string `json:"lobby_id"`

	// Optional error message when the request fails.
	// example: requested lobby not found or game already started
	Error   string `json:"error,omitempty"`
}

type LobbyRepository interface {
	Create(context.Context, *Lobby) error
	Find(context.Context, string) (*Lobby, error)
	Fetch(context.Context) ([]Lobby, error)
	Delete(context.Context, string) error
	FinishGame(context.Context, string) error
	StartGame(context.Context, string) error
	AddPlayer(context.Context, string, *auth.Player) error

	ListForLobbyScene(context.Context, int, int) ([]LobbyListItem, error)
}
