package game

// GameStateDTO is a client-safe snapshot of the game.
type GameStateDTO struct {
	GameID        string         `json:"gameId"`
	Turn          int            `json:"turn"`
	Phase         string         `json:"phase"`
	Version       int            `json:"version"`
	CurrentPlayer string         `json:"currentPlayerId,omitempty"`
	Players       []PlayerDTO    `json:"players"`
	You           *PlayerViewDTO `json:"you,omitempty"`
}

type PlayerDTO struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	Level          int            `json:"level"`
	Sex            Sex            `json:"sex"`
	Race           RaceType       `json:"race"`
	Class          ClassType      `json:"class"`
	IsDead         bool           `json:"isDead"`
	IsActive       bool           `json:"isActive"`
	CombatStrength int            `json:"combatStrength"`
	Equipment      []EquipmentDTO `json:"equipment"`
}

// PlayerViewDTO contains private player data.
type PlayerViewDTO struct {
	Hand              []CardDTO       `json:"hand"`
	CarriedItems      []EquipmentDTO  `json:"carriedItems"`
	HasHybridRace     bool            `json:"hasHybridRace"`
	HasHybridClass    bool            `json:"hasHybridClass"`
	SecondaryRace     RaceType        `json:"secondaryRace,omitempty"`
	SecondaryClass    ClassType       `json:"secondaryClass,omitempty"`
	RunAwayBonus      int             `json:"runAwayBonus"`
	HasUsedThiefSkill bool            `json:"hasUsedThiefSkill"`
}

type CardDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type EquipmentDTO struct {
	CardID     string        `json:"cardId"`
	Name       string        `json:"name"`
	Slot       EquipmentSlot `json:"slot"`
	Bonus      int           `json:"bonus"`
	IsEquipped bool          `json:"isEquipped"`
	IsBig      bool          `json:"isBig"`
}
