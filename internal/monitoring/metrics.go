package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics — структура для хранения всех метрик
type Metrics struct {
	RequestsTotal   prometheus.Counter
	RequestsLatency prometheus.Histogram
}

// NewMetrics создает новую структуру метрик
func NewMetrics() *Metrics {
	m := &Metrics{
		RequestsTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "grpc_requests_total",
				Help: "Total number of gRPC requests",
			}),
		RequestsLatency: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "grpc_request_latency_seconds",
				Help:    "Histogram of gRPC request latencies",
				Buckets: prometheus.DefBuckets,
			}),
	}
	// Регистрируем метрики
	prometheus.MustRegister(m.RequestsTotal, m.RequestsLatency)
	return m
}
