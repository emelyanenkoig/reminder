package bot

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

// Определяем метрики
var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_requests_total",
			Help: "Total number of requests received",
		},
		[]string{"endpoint"}, // Этикетка "endpoint" позволит нам отслеживать метрики по различным запросам
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bot_request_duration_seconds",
			Help:    "Duration of requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"endpoint"},
	)
	errorCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bot_errors_total",
			Help: "Total number of errors encountered",
		},
		[]string{"endpoint"},
	)
)

func init() {
	// Регистрация метрик
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(errorCounter)
}

// Экспорт метрик через HTTP
func startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Starting metrics server... ")
	go func() {
		if err := http.ListenAndServe(":2112", nil); err != nil {
			log.Fatalf("Error starting metrics server: %v", err)
		}
	}()
}
