package telemetry

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	AuthFailures = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "munchin",
			Subsystem: "auth",
			Name:      "failures_total",
			Help:      "Total number of authentication failures",
		},
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
)

func Register() {
	prometheus.MustRegister(
		AuthFailures,
		AuthSuccess,
		LobbyActive,
		LobbyClosedTotal,
		LobbyCreatedTotal,
	)
}
