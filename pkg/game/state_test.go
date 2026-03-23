package game

import (
	"errors"
	"testing"
)

func TestGameStateApplyPlayerActionFlow(t *testing.T) {
	state := newTestGameState()

	assertApplyPhase(t, state, PlayerActionCommand{PlayerID: "p1", Action: ActionOpenDoor}, PhaseLookForTrouble, "p1")
	assertApplyPhase(t, state, PlayerActionCommand{PlayerID: "p1", Action: ActionLookForTrouble}, PhaseLootRoom, "p1")
	assertApplyPhase(t, state, PlayerActionCommand{PlayerID: "p1", Action: ActionLootRoom}, PhaseCharity, "p1")
	assertApplyPhase(t, state, PlayerActionCommand{PlayerID: "p1", Action: ActionEndTurn}, PhaseOpenDoor, "p2")

	if state.turn != 2 {
		t.Fatalf("expected turn to advance to 2, got %d", state.turn)
	}

	if state.version != 5 {
		t.Fatalf("expected version to be 5 after four successful commands, got %d", state.version)
	}
}

func TestGameStateApplyRejectsCommandFromInactivePlayer(t *testing.T) {
	state := newTestGameState()

	err := state.Apply(PlayerActionCommand{PlayerID: "p2", Action: ActionOpenDoor})
	if !errors.Is(err, ErrNotYourTurn) {
		t.Fatalf("expected ErrNotYourTurn, got %v", err)
	}

	if state.phase != PhaseSetup {
		t.Fatalf("expected phase to stay %q, got %q", PhaseSetup, state.phase)
	}

	if state.version != 1 {
		t.Fatalf("expected version to stay 1, got %d", state.version)
	}
}

func TestGameStateApplyRejectsInvalidPhaseTransition(t *testing.T) {
	state := newTestGameState()

	err := state.Apply(PlayerActionCommand{PlayerID: "p1", Action: ActionLootRoom})
	if !errors.Is(err, ErrInvalidPhase) {
		t.Fatalf("expected ErrInvalidPhase, got %v", err)
	}

	if state.phase != PhaseSetup {
		t.Fatalf("expected phase to stay %q, got %q", PhaseSetup, state.phase)
	}
}

func TestGameStateApplyRequiresPlayers(t *testing.T) {
	state := NewGameState("game-empty", nil)

	err := state.Apply(PlayerActionCommand{PlayerID: "p1", Action: ActionOpenDoor})
	if !errors.Is(err, ErrNoPlayersInGame) {
		t.Fatalf("expected ErrNoPlayersInGame, got %v", err)
	}
}

func TestGameStateAddPlayerDoesNotOverwriteExistingName(t *testing.T) {
	state := newTestGameState()

	state.AddPlayer(&Player{
		ID:   "p1",
		Name: "replacement",
	})

	if got := state.players["p1"].Name; got != "Alice" {
		t.Fatalf("expected existing player name to remain Alice, got %q", got)
	}
}

func TestGameStateRemovePlayerResetsStateWhenEmpty(t *testing.T) {
	state := NewGameState("game-empty-after-remove", []*Player{{
		ID:   "p1",
		Name: "Solo",
	}})

	state.phase = PhaseCharity
	state.turn = 3
	state.RemovePlayer("p1")

	if len(state.players) != 0 {
		t.Fatalf("expected no players left, got %d", len(state.players))
	}

	if state.turn != 1 {
		t.Fatalf("expected turn reset to 1, got %d", state.turn)
	}

	if state.phase != PhaseSetup {
		t.Fatalf("expected phase reset to %q, got %q", PhaseSetup, state.phase)
	}
}

func TestGameStateToDTOExposesPrivateHandOnlyToRequestedPlayer(t *testing.T) {
	state := NewGameState("game-dto", []*Player{
		{
			ID:    "p1",
			Name:  "Alice",
			Score: 1,
			Hand: []Card{
				{ID: "card-1", Name: "Test Card"},
			},
		},
		{
			ID:    "p2",
			Name:  "Bob",
			Score: 1,
			Hand:  []Card{},
		},
	})
	state.phase = PhaseOpenDoor

	dto := state.ToDTO("p1")

	if dto.You == nil {
		t.Fatal("expected private player view for p1")
	}

	if len(dto.You.Hand) != 1 {
		t.Fatalf("expected one private hand card, got %d", len(dto.You.Hand))
	}

	activePlayers := 0
	for _, player := range dto.Players {
		if player.IsActive {
			activePlayers++
			if player.ID != "p1" {
				t.Fatalf("expected p1 to be active, got %s", player.ID)
			}
		}
	}

	if activePlayers != 1 {
		t.Fatalf("expected exactly one active player, got %d", activePlayers)
	}
}

func assertApplyPhase(t *testing.T, state *GameState, cmd PlayerActionCommand, expectedPhase Phase, expectedActivePlayer string) {
	t.Helper()

	if err := state.Apply(cmd); err != nil {
		t.Fatalf("Apply(%s) returned error: %v", cmd.Action, err)
	}

	if state.phase != expectedPhase {
		t.Fatalf("expected phase %q, got %q", expectedPhase, state.phase)
	}

	if state.currentPlayerID() != expectedActivePlayer {
		t.Fatalf("expected active player %q, got %q", expectedActivePlayer, state.currentPlayerID())
	}

	events := state.Events()
	if len(events) != 1 {
		t.Fatalf("expected exactly one pending event, got %d", len(events))
	}

	phaseEvent, ok := events[0].(TurnPhaseChangedEvent)
	if !ok {
		t.Fatalf("expected TurnPhaseChangedEvent, got %T", events[0])
	}

	if phaseEvent.Phase != string(expectedPhase) {
		t.Fatalf("expected event phase %q, got %q", expectedPhase, phaseEvent.Phase)
	}

	if phaseEvent.PlayerID != expectedActivePlayer {
		t.Fatalf("expected event active player %q, got %q", expectedActivePlayer, phaseEvent.PlayerID)
	}
}

func newTestGameState() *GameState {
	return NewGameState("game-test", []*Player{
		{
			ID:   "p1",
			Name: "Alice",
		},
		{
			ID:   "p2",
			Name: "Bob",
		},
	})
}
