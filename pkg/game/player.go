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

type Card struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Equipment struct {
	Card
	Slot       EquipmentSlot `json:"slot"`
	Bonus      int           `json:"bonus"`
	IsEquipped bool          `json:"isEquipped"`
	IsBig      bool          `json:"isBig"`
}

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

func (p *Player) CombatStrength() int {
	total := p.Level
	for _, item := range p.EquippedItems {
		total += item.Bonus
	}
	return total
}
