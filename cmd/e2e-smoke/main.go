package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	defaultBaseURL = "http://localhost:1337"
	requestTimeout = 5 * time.Second
	wsReadTimeout  = 5 * time.Second
)

type authRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

type lobbyCreationResponse struct {
	LobbyID string `json:"lobby_id"`
	Error   string `json:"error,omitempty"`
}

type lobbyListResponse struct {
	Items []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		PlayerCount int    `json:"playerCount"`
	} `json:"items"`
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"hasMore"`
}

type wsEnvelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type discardForCharityPayload struct {
	CardIDs []string `json:"cardIds"`
}

type playerActionPayload struct {
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

type gameStateView struct {
	GameID        string `json:"gameId"`
	Turn          int    `json:"turn"`
	Phase         string `json:"phase"`
	Version       int    `json:"version"`
	Started       bool   `json:"started"`
	CurrentPlayer string `json:"currentPlayerId"`
	Decks         struct {
		DungeonRemaining  int `json:"dungeonRemaining"`
		TreasureRemaining int `json:"treasureRemaining"`
		DungeonDiscard    int `json:"dungeonDiscard"`
		TreasureDiscard   int `json:"treasureDiscard"`
	} `json:"decks"`
	RevealedDoor *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Kind string `json:"kind"`
	} `json:"revealedDoor"`
	Combat *struct {
		Monster struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Kind string `json:"kind"`
		} `json:"monster"`
		Pending bool `json:"pending"`
	} `json:"combat"`
	Players []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		IsActive  bool   `json:"isActive"`
		HandCount int    `json:"handCount"`
	} `json:"players"`
	You *struct {
		Hand []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"hand"`
	} `json:"you"`
}

type gameSnapshot struct {
	Type      string        `json:"type"`
	GameID    string        `json:"gameId"`
	Timestamp string        `json:"timestamp"`
	Version   int           `json:"version"`
	State     gameStateView `json:"state"`
}

type turnPhaseChanged struct {
	Type      string `json:"type"`
	GameID    string `json:"gameId"`
	Timestamp string `json:"timestamp"`
	Version   int    `json:"version"`
	PlayerID  string `json:"playerID"`
	Phase     string `json:"phase"`
}

type cardRevealed struct {
	Type     string `json:"type"`
	PlayerID string `json:"playerID"`
	Card     struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Kind string `json:"kind"`
	} `json:"card"`
}

type commandRejected struct {
	Type        string `json:"type"`
	GameID      string `json:"gameId"`
	Timestamp   string `json:"timestamp"`
	Version     int    `json:"version"`
	CommandType string `json:"commandType"`
	Reason      string `json:"reason"`
}

func main() {
	ctx := context.Background()

	baseURL := strings.TrimRight(getEnv("MUNCHIN_BASE_URL", defaultBaseURL), "/")
	password := getEnv("MUNCHIN_E2E_PASSWORD", "secret123")
	username := getEnv("MUNCHIN_E2E_USERNAME", "e2e-"+randomSuffix())

	client := &http.Client{Timeout: requestTimeout}

	logStep("register player", "username=%s", username)
	token, err := register(ctx, client, baseURL, username, password)
	if err != nil {
		fatalf("register failed: %v", err)
	}

	logStep("create lobby")
	lobbyID, err := createLobby(ctx, client, baseURL, token)
	if err != nil {
		fatalf("create lobby failed: %v", err)
	}
	logStep("lobby created", "lobbyID=%s", lobbyID)

	logStep("list lobbies")
	listing, err := listLobbies(ctx, client, baseURL, token)
	if err != nil {
		fatalf("list lobbies failed: %v", err)
	}
	logStep("list lobbies ok", "items=%d", len(listing.Items))

	logStep("start lobby game")
	if err := startLobby(ctx, client, baseURL, token, lobbyID); err != nil {
		fatalf("start game failed: %v", err)
	}

	logStep("connect websocket")
	conn, err := connectWS(baseURL, lobbyID, token)
	if err != nil {
		fatalf("websocket connection failed: %v", err)
	}
	defer conn.Close()

	logStep("read initial snapshot")
	snapshot, err := readSnapshot(conn)
	if err != nil {
		fatalf("failed to read initial snapshot: %v", err)
	}
	logStep(
		"snapshot received",
		"phase=%s currentPlayer=%s players=%d hand=%d",
		snapshot.State.Phase,
		snapshot.State.CurrentPlayer,
		len(snapshot.State.Players),
		len(snapshot.State.You.Hand),
	)

	logStep("send OPEN_DOOR action")
	if err := sendPlayerAction(conn, "OPEN_DOOR"); err != nil {
		fatalf("failed to send player action: %v", err)
	}

	logStep("wait for action result")
	state, err := readActionResult(conn)
	if err != nil {
		fatalf("action result failed: %v", err)
	}

	if state.Combat != nil && state.Combat.Pending {
		logStep("acknowledge revealed monster")
		if err := sendAcknowledgeRevealedCard(conn); err != nil {
			fatalf("failed to send ACK_REVEALED_CARD: %v", err)
		}

		state, err = readActionResult(conn)
		if err != nil {
			fatalf("failed to resolve revealed monster: %v", err)
		}
	}

	if state.Phase == "look_for_trouble" {
		logStep("send LOOK_FOR_TROUBLE action")
		if err := sendPlayerAction(conn, "LOOK_FOR_TROUBLE"); err != nil {
			fatalf("failed to send LOOK_FOR_TROUBLE: %v", err)
		}

		state, err = readActionResult(conn)
		if err != nil {
			fatalf("look for trouble failed: %v", err)
		}
	}

	if state.Phase == "loot_room" {
		logStep("send LOOT_ROOM action")
		if err := sendPlayerAction(conn, "LOOT_ROOM"); err != nil {
			fatalf("failed to send LOOT_ROOM: %v", err)
		}

		state, err = readActionResult(conn)
		if err != nil {
			fatalf("loot room failed: %v", err)
		}
	}

	if state.Phase == "charity" && state.You != nil && len(state.You.Hand) > 5 {
		discardCount := len(state.You.Hand) - 5
		cardIDs := make([]string, 0, discardCount)
		for i := 0; i < discardCount; i++ {
			cardIDs = append(cardIDs, state.You.Hand[i].ID)
		}

		logStep("discard for charity", "count=%d", discardCount)
		if err := sendDiscardForCharity(conn, cardIDs); err != nil {
			fatalf("failed to send charity discard: %v", err)
		}

		state, err = readActionResult(conn)
		if err != nil {
			fatalf("charity discard failed: %v", err)
		}
	}

	logStep("send END_TURN action")
	if err := sendPlayerAction(conn, "END_TURN"); err != nil {
		fatalf("failed to send END_TURN: %v", err)
	}

	state, err = readActionResult(conn)
	if err != nil {
		fatalf("end turn failed: %v", err)
	}

	if state.Phase != "open_door" {
		fatalf("expected next turn to begin at open_door, got %q", state.Phase)
	}

	logStep("smoke test completed successfully")
}

func register(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	username string,
	password string,
) (string, error) {
	var resp authResponse
	err := doJSON(ctx, client, http.MethodPost, baseURL+"/auth/register", "", authRequest{
		Username: username,
		Password: password,
	}, &resp, http.StatusOK)
	if err != nil {
		return "", err
	}
	if resp.Token == "" {
		return "", fmt.Errorf("empty token in register response")
	}
	return resp.Token, nil
}

func createLobby(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	token string,
) (string, error) {
	var resp lobbyCreationResponse
	err := doJSON(
		ctx,
		client,
		http.MethodPost,
		baseURL+"/lobby",
		token,
		nil,
		&resp,
		http.StatusCreated,
	)
	if err != nil {
		return "", err
	}
	if resp.LobbyID == "" {
		return "", fmt.Errorf("empty lobby_id in create lobby response")
	}
	return resp.LobbyID, nil
}

func listLobbies(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	token string,
) (*lobbyListResponse, error) {
	var resp lobbyListResponse
	err := doJSON(ctx, client, http.MethodGet, baseURL+"/lobby", token, nil, &resp, http.StatusOK)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func startLobby(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	token string,
	lobbyID string,
) error {
	return doJSON(
		ctx,
		client,
		http.MethodPost,
		baseURL+"/lobby/"+lobbyID+"/start",
		token,
		nil,
		nil,
		http.StatusOK,
	)
}

func connectWS(baseURL string, lobbyID string, token string) (*websocket.Conn, error) {
	wsURL, err := toWebSocketURL(baseURL + "/lobby/" + lobbyID + "/ws?token=" + token)
	if err != nil {
		return nil, err
	}

	header := http.Header{}

	dialer := websocket.Dialer{
		HandshakeTimeout: requestTimeout,
	}

	conn, resp, err := dialer.Dial(wsURL, header)
	if err != nil {
		if resp != nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf(
				"dial failed with status %d: %s",
				resp.StatusCode,
				strings.TrimSpace(string(body)),
			)
		}
		return nil, err
	}

	return conn, nil
}

func readSnapshot(conn *websocket.Conn) (*gameSnapshot, error) {
	if err := conn.SetReadDeadline(time.Now().Add(wsReadTimeout)); err != nil {
		return nil, err
	}

	_, payload, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	var snapshot gameSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		return nil, fmt.Errorf("invalid snapshot payload: %w", err)
	}

	if snapshot.Type == "" {
		return nil, fmt.Errorf("snapshot missing type field")
	}
	if snapshot.Type != "game_snapshot" {
		return nil, fmt.Errorf("expected game_snapshot event, got %q", snapshot.Type)
	}
	if !snapshot.State.Started {
		return nil, fmt.Errorf("expected started game snapshot")
	}
	if snapshot.State.Phase != "open_door" {
		return nil, fmt.Errorf("expected open_door phase, got %q", snapshot.State.Phase)
	}
	if snapshot.State.You == nil {
		return nil, fmt.Errorf("expected private player snapshot")
	}
	if len(snapshot.State.You.Hand) != 8 {
		return nil, fmt.Errorf("expected initial hand size 8")
	}

	return &snapshot, nil
}

func sendPlayerAction(conn *websocket.Conn, action string) error {
	payload := wsEnvelope{
		Type: "PLAYER_ACTION",
		Data: mustMarshalJSON(playerActionPayload{
			Action:    action,
			Timestamp: time.Now().UTC(),
		}),
	}
	return conn.WriteJSON(payload)
}

func sendDiscardForCharity(conn *websocket.Conn, cardIDs []string) error {
	payload := wsEnvelope{
		Type: "DISCARD_FOR_CHARITY",
		Data: mustMarshalJSON(discardForCharityPayload{
			CardIDs: cardIDs,
		}),
	}
	return conn.WriteJSON(payload)
}

func sendAcknowledgeRevealedCard(conn *websocket.Conn) error {
	payload := wsEnvelope{
		Type: "ACK_REVEALED_CARD",
		Data: mustMarshalJSON(map[string]any{}),
	}
	return conn.WriteJSON(payload)
}

func readActionResult(conn *websocket.Conn) (*gameStateView, error) {
	for {
		if err := conn.SetReadDeadline(time.Now().Add(wsReadTimeout)); err != nil {
			return nil, err
		}

		_, payload, err := conn.ReadMessage()
		if err != nil {
			return nil, err
		}

		var raw map[string]any
		if err := json.Unmarshal(payload, &raw); err != nil {
			return nil, fmt.Errorf("invalid action result payload: %w", err)
		}

		eventType, _ := raw["type"].(string)
		switch eventType {
		case "card_revealed":
			var evt cardRevealed
			if err := json.Unmarshal(payload, &evt); err != nil {
				return nil, err
			}
			logStep("card revealed", "name=%s kind=%s", evt.Card.Name, evt.Card.Kind)
		case "TURN_PHASE_CHANGE":
			var evt turnPhaseChanged
			if err := json.Unmarshal(payload, &evt); err != nil {
				return nil, err
			}
			logStep("phase changed", "playerID=%s phase=%s", evt.PlayerID, evt.Phase)
		case "game_snapshot":
			var evt struct {
				Type  string        `json:"type"`
				State gameStateView `json:"state"`
			}
			if err := json.Unmarshal(payload, &evt); err != nil {
				return nil, err
			}
			handSize := 0
			if evt.State.You != nil {
				handSize = len(evt.State.You.Hand)
			}
			logStep("snapshot update received", "phase=%s hand=%d", evt.State.Phase, handSize)
			return &evt.State, nil
		case "command_rejected", "ERROR":
			var evt commandRejected
			if err := json.Unmarshal(payload, &evt); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("command rejected: %s", evt.Reason)
		default:
			return nil, fmt.Errorf(
				"unexpected websocket event type: %q payload=%s",
				eventType,
				string(payload),
			)
		}
	}
}

func doJSON(
	ctx context.Context,
	client *http.Client,
	method string,
	endpoint string,
	token string,
	requestBody any,
	responseBody any,
	expectedStatus int,
) error {
	var body io.Reader
	if requestBody != nil {
		raw, err := json.Marshal(requestBody)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != expectedStatus {
		return fmt.Errorf(
			"unexpected status %d from %s %s: %s",
			resp.StatusCode,
			method,
			endpoint,
			strings.TrimSpace(string(rawBody)),
		)
	}

	if responseBody == nil || len(rawBody) == 0 {
		return nil
	}

	if err := json.Unmarshal(rawBody, responseBody); err != nil {
		return fmt.Errorf("invalid JSON response from %s %s: %w", method, endpoint, err)
	}

	return nil
}

func toWebSocketURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	switch parsed.Scheme {
	case "http":
		parsed.Scheme = "ws"
	case "https":
		parsed.Scheme = "wss"
	}

	return parsed.String(), nil
}

func mustMarshalJSON(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return raw
}

func randomSuffix() string {
	var buf [4]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf[:])
}

func getEnv(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func logStep(step string, details ...any) {
	if len(details) == 0 {
		fmt.Printf("==> %s\n", step)
		return
	}

	format, ok := details[0].(string)
	if !ok {
		fmt.Printf("==> %s: %v\n", step, details[0])
		return
	}

	fmt.Printf("==> %s: %s\n", step, fmt.Sprintf(format, details[1:]...))
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
