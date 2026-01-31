package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	ActiveConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "active_connections",
		Help: "Active ulanishlar",
	})
	BroadcastMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "broadcast_messages_total",
		Help: "Jami uzatilgan xabarlar soni",
	})
)
