package game

// GameStateDTO is a client-safe snapshot of the game.
type GameStateDTO struct {
	// Unique game identifier.
	GameID        string         `json:"gameId"`

	// 1-based turn counter.
	Turn          int            `json:"turn"`

	// Current phase of the active turn.
	Phase         string         `json:"phase"`

	// Monotonic state version for synchronization.
	Version       int            `json:"version"`

	// Identifier of the active player.
	CurrentPlayer string         `json:"currentPlayerId,omitempty"`

	// Public state for all players.
	Players       []PlayerDTO    `json:"players"`

	// Private state for the requesting player.
	You           *PlayerViewDTO `json:"you,omitempty"`
}

// PlayerDTO is the public projection of a player in the game snapshot.
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

// CardDTO is a client-safe card payload.
type CardDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// EquipmentDTO is the client-safe representation of an item on board or carried.
type EquipmentDTO struct {
	CardID     string        `json:"cardId"`
	Name       string        `json:"name"`
	Slot       EquipmentSlot `json:"slot"`
	Bonus      int           `json:"bonus"`
	IsEquipped bool          `json:"isEquipped"`
	IsBig      bool          `json:"isBig"`
}
