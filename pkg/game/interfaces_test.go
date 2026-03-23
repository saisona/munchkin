package game

import (
	"errors"
	"testing"
	"time"
)

func TestDecodeCommandMessagePlayerAction(t *testing.T) {
	timestamp := "2026-03-23T10:00:00Z"
	payload := []byte(`{"type":"PLAYER_ACTION","data":{"action":"OPEN_DOOR","timestamp":"` + timestamp + `"}}`)

	cmd, err := DecodeCommandMessage("player-1", payload)
	if err != nil {
		t.Fatalf("DecodeCommandMessage returned an error: %v", err)
	}

	playerAction, ok := cmd.(PlayerActionCommand)
	if !ok {
		t.Fatalf("expected PlayerActionCommand, got %T", cmd)
	}

	if playerAction.PlayerID != "player-1" {
		t.Fatalf("expected player ID player-1, got %q", playerAction.PlayerID)
	}

	if playerAction.Action != ActionOpenDoor {
		t.Fatalf("expected action %q, got %q", ActionOpenDoor, playerAction.Action)
	}

	expectedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		t.Fatalf("failed to parse timestamp fixture: %v", err)
	}

	if !playerAction.Timestamp.Equal(expectedTime) {
		t.Fatalf("expected timestamp %s, got %s", expectedTime, playerAction.Timestamp)
	}
}

func TestDecodeCommandMessageRejectsUnknownType(t *testing.T) {
	payload := []byte(`{"type":"UNKNOWN","data":{}}`)

	_, err := DecodeCommandMessage("player-1", payload)
	if !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("expected ErrUnknownCommand, got %v", err)
	}
}

func TestDecodeCommandMessageRejectsInvalidJSON(t *testing.T) {
	payload := []byte(`{"type":"PLAYER_ACTION","data":`)

	_, err := DecodeCommandMessage("player-1", payload)
	if err == nil {
		t.Fatal("expected JSON decoding error, got nil")
	}
}
