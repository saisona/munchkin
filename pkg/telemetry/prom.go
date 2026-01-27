package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	AuthFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "auth",
			Name:      "failures_total",
			Help:      "Total number of authentication failures",
		},
		[]string{"reason"},
	)
	AuthSuccess = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "auth",
			Name:      "success_total",
			Help:      "Total number of successful authentications",
		},
	)
)

var (
	PlayersDisconnected = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "game",
			Name:      "player_disconnected",
			Help:      "Total number of player disconnected.",
		})
	PlayersConnected = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "munchin",
		Subsystem: "game",
		Name:      "player_connected",
	})
	RoomStarted = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "munchin",
		Subsystem: "game",
		Name:      "room_started",
	})
)

func Register() {
	prometheus.MustRegister(
		AuthFailures,
		AuthSuccess,
	)

	MustRegisterGameMetrics(prometheus.DefaultRegisterer)
	MustRegisterLobbyMetrics(prometheus.DefaultRegisterer)
}
