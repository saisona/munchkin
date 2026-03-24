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

type playerActionPayload struct {
	Action    string    `json:"action"`
	Timestamp time.Time `json:"timestamp"`
}

type gameSnapshot struct {
	Type      string `json:"type"`
	GameID    string `json:"gameId"`
	Timestamp string `json:"timestamp"`
	Version   int    `json:"version"`
	State     struct {
		GameID        string `json:"gameId"`
		Turn          int    `json:"turn"`
		Phase         string `json:"phase"`
		Version       int    `json:"version"`
		CurrentPlayer string `json:"currentPlayerId"`
		Players       []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			IsActive bool   `json:"isActive"`
		} `json:"players"`
		You *struct {
			Hand []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"hand"`
		} `json:"you"`
	} `json:"state"`
}

type turnPhaseChanged struct {
	Type      string `json:"type"`
	GameID    string `json:"gameId"`
	Timestamp string `json:"timestamp"`
	Version   int    `json:"version"`
	PlayerID  string `json:"playerID"`
	Phase     string `json:"phase"`
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
		"phase=%s currentPlayer=%s players=%d",
		snapshot.State.Phase,
		snapshot.State.CurrentPlayer,
		len(snapshot.State.Players),
	)

	logStep("send OPEN_DOOR action")
	if err := sendPlayerAction(conn, "OPEN_DOOR"); err != nil {
		fatalf("failed to send player action: %v", err)
	}

	logStep("wait for action result")
	if err := readActionResult(conn); err != nil {
		fatalf("action result failed: %v", err)
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

func readActionResult(conn *websocket.Conn) error {
	if err := conn.SetReadDeadline(time.Now().Add(wsReadTimeout)); err != nil {
		return err
	}

	_, payload, err := conn.ReadMessage()
	if err != nil {
		return err
	}

	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return fmt.Errorf("invalid action result payload: %w", err)
	}

	eventType, _ := raw["type"].(string)
	switch eventType {
	case "TURN_PHASE_CHANGE":
		var evt turnPhaseChanged
		if err := json.Unmarshal(payload, &evt); err != nil {
			return err
		}
		logStep("phase changed", "playerID=%s phase=%s", evt.PlayerID, evt.Phase)
		return nil
	case "command_rejected", "ERROR":
		var evt commandRejected
		if err := json.Unmarshal(payload, &evt); err != nil {
			return err
		}
		return fmt.Errorf("command rejected: %s", evt.Reason)
	default:
		return fmt.Errorf(
			"unexpected websocket event type: %q payload=%s",
			eventType,
			string(payload),
		)
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
