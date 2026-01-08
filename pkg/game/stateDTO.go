package game

// GameStateDTO is a client-safe snapshot of the game.
type GameStateDTO struct {
	GameID  string         `json:"gameId"`
	Turn    int            `json:"turn"`
	Phase   string         `json:"phase"`
	Version int            `json:"version"`
	Players []PlayerDTO    `json:"players"`
	You     *PlayerViewDTO `json:"you,omitempty"`
}

type PlayerDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Score    int    `json:"score"`
	IsActive bool   `json:"isActive"`
}

// PlayerViewDTO contains private player data.
type PlayerViewDTO struct {
	Hand []CardDTO `json:"hand"`
}

type CardDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
