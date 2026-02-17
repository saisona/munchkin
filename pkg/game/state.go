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

	turn    int
	phase   Phase
	version int

	// internal only
	pendingEvents []Event
}

var ErrUnknownCommand = errors.New("command is unknown or not yet implemented")

func NewGameState(gameID string, players []*Player) *GameState {
	m := make(map[string]*Player, len(players))
	for _, p := range players {
		m[p.ID] = p
	}

	return &GameState{
		id:      gameID,
		players: m,
		turn:    1,
		phase:   PhaseSetup,
		version: 1,
	}
}

func (s *GameState) Apply(cmd Command) error {
	switch cmd := cmd.(type) {
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
