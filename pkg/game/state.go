package game

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

// GameState is the authoritative, mutable state.
// NEVER send this directly to clients.
type GameState struct {
	id      string
	players map[string]*Player
	order   []string

	turn    int
	phase   Phase
	version int

	// internal only
	pendingEvents []Event
}

var ErrUnknownCommand = errors.New("command is unknown or not yet implemented")

func NewGameState(gameID string, players []*Player) *GameState {
	m := make(map[string]*Player, len(players))
	order := make([]string, 0, len(players))
	for _, p := range players {
		m[p.ID] = p
		order = append(order, p.ID)
	}

	return &GameState{
		id:      gameID,
		players: m,
		order:   order,
		turn:    1,
		phase:   PhaseSetup,
		version: 1,
	}
}

func (s *GameState) Apply(cmd Command) error {
	switch cmd := cmd.(type) {
	case PlayerActionCommand:
		if err := s.applyPlayerAction(cmd); err != nil {
			return err
		}
	case PlayCardCommand:

		if err := s.playCard(cmd); err != nil {
			return err
		}
		// Generate domain event
		s.pendingEvents = append(s.pendingEvents, CardPlayedEvent{
			BaseEvent: BaseEvent{
				Type:      "card_played",
				GameID:    s.id,
				Timestamp: time.Now(),
				Version:   s.version,
			},
			PlayerID: cmd.GetPlayerID(),
			CardID:   cmd.GetCardID(),
		})
	case DrawCardCommand:
		card, err := s.drawCard(cmd.GetPlayerID())
		if err != nil {
			return err
		}
		s.pendingEvents = append(s.pendingEvents, CardDrawnEvent{
			BaseEvent: BaseEvent{
				Type:      "card_drawn",
				GameID:    s.id,
				Timestamp: time.Now(),
				Version:   s.version,
			},
			PlayerID: cmd.GetPlayerID(),
			CardID:   card.ID,
		})

	default:
		return ErrUnknownCommand
	}

	telemetry.CommandsTotal.With(prometheus.Labels{"command": cmd.Type()}).Inc()

	// bump version after successful command
	s.version++

	return nil
}

func (s *GameState) AddPlayer(player *Player) {
	if player == nil || player.ID == "" {
		return
	}

	if existing, ok := s.players[player.ID]; ok {
		if existing.Name == "" && player.Name != "" {
			existing.Name = player.Name
		}
		return
	}

	s.players[player.ID] = player
	s.order = append(s.order, player.ID)
}

func (s *GameState) RemovePlayer(playerID string) {
	delete(s.players, playerID)

	for i, id := range s.order {
		if id != playerID {
			continue
		}

		s.order = append(s.order[:i], s.order[i+1:]...)
		break
	}

	if len(s.order) == 0 {
		s.phase = PhaseSetup
		s.turn = 1
		return
	}

	if s.turn > len(s.order) {
		s.turn = len(s.order)
	}
}

func (s *GameState) currentPlayerID() string {
	if len(s.order) == 0 {
		return ""
	}

	index := s.turn - 1
	if index < 0 {
		index = 0
	}
	if index >= len(s.order) {
		index = index % len(s.order)
	}

	return s.order[index]
}

// Events returns all pending events and clears the buffer.
// Must be called **after Apply(cmd)**.
func (s *GameState) Events() []Event {
	// Make a copy so we can safely return it
	events := make([]Event, len(s.pendingEvents))
	copy(events, s.pendingEvents)

	// Clear pending events
	s.pendingEvents = s.pendingEvents[:0]

	return events
}

func (s *GameState) playCard(pcc PlayCardCommand) error {
	logger.With(slog.String("CardID", pcc.GetCardID()), slog.String("playerID", pcc.GetPlayerID())).Info("card played")
	return nil
}

func (s *GameState) applyPlayerAction(cmd PlayerActionCommand) error {
	if len(s.order) == 0 {
		return ErrNoPlayersInGame
	}

	if cmd.PlayerID != s.currentPlayerID() {
		return ErrNotYourTurn
	}

	switch cmd.Action {
	case ActionOpenDoor:
		if s.phase != PhaseSetup && s.phase != PhaseOpenDoor {
			return ErrInvalidPhase
		}
		s.phase = PhaseLookForTrouble
	case ActionLookForTrouble:
		if s.phase != PhaseLookForTrouble {
			return ErrInvalidPhase
		}
		s.phase = PhaseLootRoom
	case ActionLootRoom:
		if s.phase != PhaseLootRoom {
			return ErrInvalidPhase
		}
		s.phase = PhaseCharity
	case ActionEndTurn:
		if s.phase != PhaseCharity && s.phase != PhaseLookForTrouble && s.phase != PhaseLootRoom {
			return ErrInvalidPhase
		}
		s.turn = (s.turn % len(s.order)) + 1
		s.phase = PhaseOpenDoor
	default:
		return ErrUnknownCommand
	}

	s.pendingEvents = append(s.pendingEvents, TurnPhaseChangedEvent{
		BaseEvent: BaseEvent{
			Type:      "TURN_PHASE_CHANGE",
			GameID:    s.id,
			Timestamp: time.Now(),
			Version:   s.version,
		},
		PlayerID: s.currentPlayerID(),
		Phase:    string(s.phase),
	})

	return nil
}

// drawCard for a playerID given as attribute
// Should return a new card everytime, or exects to return the associated error
func (s *GameState) drawCard(playerID string) (Card, error) {
	c := Card{
		Name: generateCardName(),
		ID:   uuid.NewString(),
	}

	logger.With(slog.String("CardID", c.ID), slog.String("playerID", playerID)).Info("card drawn")
	return c, nil
}

// TODO: remove this function when a real Card Drawer is implemented
func generateCardName() string {
	strBuilder := strings.Builder{}
	defer strBuilder.Reset()

	strBuilder.WriteString("fake_card-")
	strBuilder.WriteString(uuid.NewString())
	value := strBuilder.String()
	return value
}
