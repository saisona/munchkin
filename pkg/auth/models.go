package auth

import "time"

type Player struct {
	PlayerID     string    `json:"player_id" gorm:"primaryKey"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"hash"`
	CreatedAt    time.Time `json:"createdAt"`
}
