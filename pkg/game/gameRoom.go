package game

import (
	"log/slog"
	"time"

	"dev.azure.com/saisona/Munchin/munchin-api/pkg/telemetry"
)

type GameRoom struct {
	lobbyID string

	// Channels
	joinCh  chan *PlayerConn
	leaveCh chan string
	cmdCh   chan Command
	stopCh  chan struct{}

	players map[string]*PlayerConn
	state   *GameState
}

func NewGameRoom(
	lobbyID string,
	initialState *GameState,
) *GameRoom {
	return &GameRoom{
		lobbyID: lobbyID,
		joinCh:  make(chan *PlayerConn),
		leaveCh: make(chan string),
		cmdCh:   make(chan Command, 32),
		stopCh:  make(chan struct{}),

		players: make(map[string]*PlayerConn),
		state:   initialState,
	}
}

func (r *GameRoom) getLogger() *slog.Logger {
	return logger.With(slog.String("component", "game_room"), slog.String("lobbyID", r.lobbyID))
}

func (r *GameRoom) Run() {
	logger.Info("game room started", "lobbyID", r.lobbyID)
	telemetry.RoomStarted.Inc()

	for {
		select {
		case p := <-r.joinCh:
			r.handleJoin(p)

		case playerID := <-r.leaveCh:
			r.handleLeave(playerID)

		case cmd := <-r.cmdCh:
			r.handleCommand(cmd)

		case <-r.stopCh:
			r.shutdown()
			return
		}
	}
}

func (r *GameRoom) handleCommand(cmd Command) {
	start := time.Now()
	logger.With(slog.String("cmdType", cmd.Type()), slog.String("playerID", cmd.GetPlayerID())).Debug("received command to handle")

	telemetry.CommandsTotal.
		WithLabelValues(cmd.Type()).
		Inc()

	if err := r.state.Apply(cmd); err != nil {
		telemetry.CommandErrors.
			WithLabelValues(cmd.Type()).
			Inc()

		rejected := CommandRejectedEvent{
			BaseEvent: BaseEvent{
				Type:      "command_rejected",
				GameID:    r.lobbyID,
				Timestamp: time.Now(),
				Version:   r.state.version,
			},
			CommandType: cmd.Type(),
			Reason:      err.Error(),
		}

		r.sendTo(cmd.GetPlayerID(), rejected)

		logger.With(slog.String("playerID", cmd.GetPlayerID()), slog.String("command", cmd.Type())).Warn("command rejected", "error", err)

		return
	}

	events := r.state.Events()
	for _, evt := range events {
		r.broadcast(evt)

		telemetry.EventsEmitted.
			WithLabelValues(evt.EventType()).
			Inc()
	}

	telemetry.CommandDuration.
		WithLabelValues(cmd.Type()).
		Observe(time.Since(start).Seconds())
}

func (r *GameRoom) handleJoin(p *PlayerConn) {
	// Register player
	r.players[p.PlayerID] = p
	telemetry.PlayersConnected.Inc()

	logger.Info(
		"player joined game",
		"lobbyID", r.lobbyID,
		"playerID", p.PlayerID,
		"players", len(r.players),
	)

	// Send full snapshot FIRST (private view)
	snapshot := GameSnapshotEvent{
		BaseEvent: BaseEvent{
			Type:      "game_snapshot",
			GameID:    r.lobbyID,
			Timestamp: time.Now(),
			Version:   r.state.version,
		},
		State: r.state.ToDTO(p.PlayerID),
	}

	p.Send <- snapshot

	// Notify other players
	joined := PlayerJoinedEvent{
		BaseEvent: BaseEvent{
			Type:      "player_joined",
			GameID:    r.lobbyID,
			Timestamp: time.Now(),
			Version:   r.state.version,
		},
		PlayerID: p.PlayerID,
	}

	r.broadcastExcept(joined, p.PlayerID)
}

func (r *GameRoom) Join(p *PlayerConn) {
	r.joinCh <- p
}

func (r *GameRoom) handleLeave(playerID string) {
	delete(r.players, playerID)
	telemetry.PlayersDisconnected.Inc()

	r.broadcast(PlayerLeftEvent{
		BaseEvent: BaseEvent{
			Type:      "player_left",
			GameID:    r.lobbyID,
			Timestamp: time.Now(),
			Version:   r.state.version,
		},
		PlayerID: playerID,
	})

	if len(r.players) == 0 {
		r.stopCh <- struct{}{}
	}
}

func (r *GameRoom) shutdown() {
	r.getLogger().Info("game room shutting down")
}

func (r *GameRoom) broadcast(evt Event) {
	r.getLogger().With(slog.String("evtType", evt.EventType())).Info("sending event to everyone")
	for _, p := range r.players {
		p.Send <- evt
	}
}

func (r *GameRoom) broadcastExcept(evt Event, exceptPlayerID string) {
	r.getLogger().With(slog.String("evtType", evt.EventType())).Info("sending event to everyone but sender")
	for id, p := range r.players {
		if id == exceptPlayerID {
			continue
		}
		p.Send <- evt
	}
}

func (r *GameRoom) sendTo(playerID string, evt Event) {
	r.getLogger().With(slog.String("evtType", evt.EventType())).Info("sending event to sender")
	r.players[playerID].Send <- evt
}
