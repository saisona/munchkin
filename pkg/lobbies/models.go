package lobbies

import (
	"time"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/auth"
)

type Lobby struct {
	LobbyID    string        `json:"lobby_id"   gorm:"primaryKey"`
	CreatedAt  time.Time     `json:"createAt"`
	FinishedAt time.Time     `json:"finishedAt"`
	Players    []auth.Player `json:"players"    gorm:"foreignKey:PlayerID;references:LobbyID"`
}
