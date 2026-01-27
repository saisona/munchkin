package telemetry

import "github.com/prometheus/client_golang/prometheus"

var (
	// CommandsTotal counts all received commands by type
	CommandsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "game",
			Name:      "commands_total",
			Help:      "Total number of game commands received",
		},
		[]string{"command"},
	)

	// CommandErrors counts rejected commands by type
	CommandErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "game",
			Name:      "command_errors_total",
			Help:      "Total number of rejected game commands",
		},
		[]string{"command"},
	)

	// CommandDuration measures how long command processing takes
	CommandDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "munchin",
			Subsystem: "game",
			Name:      "command_duration_seconds",
			Help:      "Duration of game command processing",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"command"},
	)

	// EventsEmitted counts domain events emitted by type
	EventsEmitted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "game",
			Name:      "events_emitted_total",
			Help:      "Total number of game events emitted",
		},
		[]string{"event"},
	)
)

var (
	GameRoomsCreated = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "game",
			Name:      "rooms_created_total",
			Help:      "Total number of game rooms created",
		},
	)

	GameRoomsDestroyed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "game",
			Name:      "rooms_destroyed_total",
			Help:      "Total number of game rooms destroyed",
		},
		[]string{"reason"},
	)

	GameRoomsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "munchin",
			Subsystem: "game",
			Name:      "rooms_active",
			Help:      "Number of active game rooms",
		},
	)

	GameRoomJoins = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "game",
			Name:      "room_joins_total",
			Help:      "Total number of players joining game rooms",
		},
	)
)

func MustRegisterGameMetrics(reg prometheus.Registerer) {
	reg.MustRegister(
		CommandsTotal,
		CommandErrors,
		CommandDuration,
		EventsEmitted,

		GameRoomJoins,
		GameRoomsActive,
		GameRoomsCreated,
		GameRoomsDestroyed,
	)
}
