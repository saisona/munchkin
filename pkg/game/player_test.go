package game

import "testing"

func TestNewPlayerSetsMunchkinDefaults(t *testing.T) {
	player := NewPlayer("player-1", "Alice")

	if player.ID != "player-1" {
		t.Fatalf("expected player ID player-1, got %q", player.ID)
	}

	if player.Level != 1 {
		t.Fatalf("expected level 1, got %d", player.Level)
	}

	if player.Race != RaceHuman {
		t.Fatalf("expected race %q, got %q", RaceHuman, player.Race)
	}

	if player.Class != ClassNone {
		t.Fatalf("expected class %q, got %q", ClassNone, player.Class)
	}

	if player.Hand == nil || player.EquippedItems == nil || player.CarriedItems == nil {
		t.Fatal("expected initialized card and equipment slices")
	}
}

func TestPlayerCombatStrengthUsesLevelAndEquippedItemsOnly(t *testing.T) {
	player := NewPlayer("player-1", "Alice")
	player.Level = 4
	player.EquippedItems = []Equipment{
		{
			Card:       Card{ID: "eq-1", Name: "Sword"},
			Bonus:      3,
			IsEquipped: true,
		},
		{
			Card:       Card{ID: "eq-2", Name: "Shield"},
			Bonus:      1,
			IsEquipped: true,
		},
	}
	player.CarriedItems = []Equipment{
		{
			Card:       Card{ID: "carry-1", Name: "Spare Bow"},
			Bonus:      99,
			IsEquipped: false,
		},
	}

	if strength := player.CombatStrength(); strength != 8 {
		t.Fatalf("expected combat strength 8, got %d", strength)
	}
}
