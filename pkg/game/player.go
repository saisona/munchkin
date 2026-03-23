package game

type Phase string

const (
	PhaseSetup          Phase = "setup"
	PhaseOpenDoor       Phase = "open_door"
	PhaseLookForTrouble Phase = "look_for_trouble"
	PhaseLootRoom       Phase = "loot_room"
	PhaseCharity        Phase = "charity"
)

type Player struct {
	ID    string
	Name  string
	Score int
	Hand  []Card
}
type Card struct {
	ID   string
	Name string
}
