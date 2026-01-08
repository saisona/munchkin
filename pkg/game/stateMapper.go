package game

func (s *GameState) ToDTO(forPlayerID string) GameStateDTO {
	players := make([]PlayerDTO, 0, len(s.players))

	for _, p := range s.players {
		players = append(players, PlayerDTO{
			ID:       p.ID,
			Name:     p.Name,
			Score:    p.Score,
			IsActive: s.isActivePlayer(p.ID),
		})
	}

	var you *PlayerViewDTO
	if p, ok := s.players[forPlayerID]; ok {
		you = &PlayerViewDTO{
			Hand: toCardDTOs(p.Hand),
		}
	}

	return GameStateDTO{
		GameID:  s.id,
		Turn:    s.turn,
		Phase:   string(s.phase),
		Version: s.version,
		Players: players,
		You:     you,
	}
}

func (s *GameState) isActivePlayer(_ string) bool {
	// example rule: current turn belongs to player order
	// adapt as needed
	return s.turn%len(s.players) == 0
}

func toCardDTOs(cards []Card) []CardDTO {
	out := make([]CardDTO, len(cards))
	for i, c := range cards {
		out[i] = CardDTO(c)
	}
	return out
}
