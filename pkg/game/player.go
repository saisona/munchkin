package game

type Phase string

const (
	PhaseSetup Phase = "setup"
	PhasePlay  Phase = "play"
	PhaseEnd   Phase = "end"
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
