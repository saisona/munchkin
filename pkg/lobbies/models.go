package lobbies

import (
	"time"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
)

type LobbyState string

const (
	StateAvailable LobbyState = "ACTIVE"
	StateFull      LobbyState = "FULL"
	StateInGame    LobbyState = "IN_GAME"
)

type Lobby struct {
	ID         string         `json:"lobby_id"   gorm:"primaryKey"`
	State      LobbyState     `json:"state"`
	CreatedAt  time.Time      `json:"createAt"`
	FinishedAt time.Time      `json:"finishedAt"`
	Players    []*auth.Player `json:"players"    gorm:"many2many:lobby_players"`
}

// LobbyListItem represents a lobby entry in a lobby list.
type LobbyListItem struct {
	// Unique lobby identifier.
	// example: lobby-1234
	ID string `json:"id"`

	// Display name of the lobby.
	// example: Casual Game
	Name string `json:"name"`

	// Number of players currently in the lobby.
	// example: 3
	PlayerCount int `json:"playerCount"`
}
