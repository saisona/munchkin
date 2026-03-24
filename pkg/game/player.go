package game

type Phase string

const (
	PhaseSetup          Phase = "setup"
	PhaseOpenDoor       Phase = "open_door"
	PhaseLookForTrouble Phase = "look_for_trouble"
	PhaseLootRoom       Phase = "loot_room"
	PhaseCharity        Phase = "charity"
)

type Sex string

const (
	SexMale   Sex = "MALE"
	SexFemale Sex = "FEMALE"
)

type RaceType string

const (
	RaceHuman    RaceType = "HUMAN"
	RaceElf      RaceType = "ELF"
	RaceDwarf    RaceType = "DWARF"
	RaceHalfling RaceType = "HALFLING"
)

type ClassType string

const (
	ClassNone    ClassType = "NONE"
	ClassWarrior ClassType = "WARRIOR"
	ClassThief   ClassType = "THIEF"
	ClassMage    ClassType = "MAGE"
	ClassCleric  ClassType = "CLERIC"
)

type EquipmentSlot string

const (
	EquipmentSlotHead     EquipmentSlot = "HEAD"
	EquipmentSlotArmor    EquipmentSlot = "ARMOR"
	EquipmentSlotFeet     EquipmentSlot = "FEET"
	EquipmentSlotHand     EquipmentSlot = "HAND"
	EquipmentSlotTwoHands EquipmentSlot = "TWO_HANDS"
	EquipmentSlotNone     EquipmentSlot = "NONE"
)

// Card represents a simplified card view used by the current backend state.
type Card struct {
	// Unique card identifier.
	ID   string `json:"id"`

	// Display name of the card.
	Name string `json:"name"`
}

// Equipment represents an item card visible on the board.
type Equipment struct {
	Card

	// Board slot occupied by the equipment.
	Slot       EquipmentSlot `json:"slot"`

	// Combat bonus granted by the equipment.
	Bonus      int           `json:"bonus"`

	// Indicates whether the item is currently equipped.
	IsEquipped bool          `json:"isEquipped"`

	// Indicates whether the item is considered a big item.
	IsBig      bool          `json:"isBig"`
}

// Player is the in-memory authoritative representation of a player state.
type Player struct {
	ID                    string
	Name                  string
	Level                 int
	Sex                   Sex
	Race                  RaceType
	SecondaryRace         RaceType
	Class                 ClassType
	SecondaryClass        ClassType
	HasHybridRace         bool
	HasHybridClass        bool
	EquippedItems         []Equipment
	CarriedItems          []Equipment
	Hand                  []Card
	IsDead                bool
	HasUsedThiefAbility   bool
	AvailableRunAwayBonus int
}

// NewPlayer creates a player with Munchkin-aligned defaults.
func NewPlayer(id string, name string) *Player {
	if name == "" {
		name = id
	}

	return &Player{
		ID:             id,
		Name:           name,
		Level:          1,
		Sex:            SexMale,
		Race:           RaceHuman,
		SecondaryRace:  "",
		Class:          ClassNone,
		SecondaryClass: "",
		EquippedItems:  []Equipment{},
		CarriedItems:   []Equipment{},
		Hand:           []Card{},
	}
}

// CombatStrength returns the current combat strength from level and equipped items.
func (p *Player) CombatStrength() int {
	total := p.Level
	for _, item := range p.EquippedItems {
		total += item.Bonus
	}
	return total
}
