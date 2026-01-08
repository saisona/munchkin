package telemetry

import "github.com/prometheus/client_golang/prometheus"

var (
	LobbyActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "munchin",
		Subsystem: "lobby",
		Name:      "active",
		Help:      "Number of currently active lobbies.",
	})
	LobbyCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "munchin",
		Subsystem: "lobby",
		Name:      "created_total",
		Help:      "Total number of lobbies created.",
	})
	LobbyClosedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "lobby",
			Name:      "closed_total",
			Help:      "Total number of closed lobbies by reason.",
		},
		[]string{"reason"},
	)
	WSUpgradeFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "lobby",
			Name:      "ws_upgrade_failures",
			Help:      "Total number of websocket upgrade failure by reason.",
		},
		[]string{"reason"})
)

func MustRegisterLobbyMetrics(reg prometheus.Registerer) {
	reg.MustRegister(
		WSUpgradeFailures,
		LobbyClosedTotal,
		LobbyCreatedTotal,
		LobbyActive,
	)
}
