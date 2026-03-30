package game

func (s *GameState) ToDTO(forPlayerID string) GameStateDTO {
	players := make([]PlayerDTO, 0, len(s.players))

	for _, p := range s.players {
		players = append(players, PlayerDTO{
			ID:             p.ID,
			Name:           p.Name,
			Level:          p.Level,
			Sex:            p.Sex,
			Race:           p.Race,
			Class:          p.Class,
			IsDead:         p.IsDead,
			IsActive:       s.isActivePlayer(p.ID),
			CombatStrength: p.CombatStrength(),
			Equipment:      toEquipmentDTOs(p.EquippedItems),
		})
	}

	var you *PlayerViewDTO
	if p, ok := s.players[forPlayerID]; ok {
		you = &PlayerViewDTO{
			Hand:              toCardDTOs(p.Hand),
			CarriedItems:      toEquipmentDTOs(p.CarriedItems),
			HasHybridRace:     p.HasHybridRace,
			HasHybridClass:    p.HasHybridClass,
			SecondaryRace:     p.SecondaryRace,
			SecondaryClass:    p.SecondaryClass,
			RunAwayBonus:      p.AvailableRunAwayBonus,
			HasUsedThiefSkill: p.HasUsedThiefAbility,
		}
	}

	return GameStateDTO{
		GameID:        s.id,
		Turn:          s.turn,
		Phase:         string(s.phase),
		Version:       s.version,
		CurrentPlayer: s.currentPlayerID(),
		Players:       players,
		You:           you,
	}
}

func (s *GameState) isActivePlayer(playerID string) bool {
	return s.currentPlayerID() == playerID
}

func toCardDTOs(cards []Card) []CardDTO {
	out := make([]CardDTO, len(cards))
	for i, c := range cards {
		out[i] = CardDTO(c)
	}
	return out
}

func toEquipmentDTOs(items []Equipment) []EquipmentDTO {
	out := make([]EquipmentDTO, len(items))
	for i, item := range items {
		out[i] = EquipmentDTO{
			CardID:     item.ID,
			Name:       item.Name,
			Slot:       item.Slot,
			Bonus:      item.Bonus,
			IsEquipped: item.IsEquipped,
			IsBig:      item.IsBig,
		}
	}
	return out
}
