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

type LobbyListItem struct {
	ID          string     `json:"id"`
	State       LobbyState `json:"state"`
	PlayerCount int        `json:"playerCount"`
}
