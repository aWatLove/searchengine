package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"runtime"
	"time"
)

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP-запросов",
		},
		[]string{"handler", "method", "status"},
	)
	errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Общее количество ошибок HTTP",
		},
		[]string{"handler", "method"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Время обработки HTTP-запросов",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"handler", "method"},
	)
	cpuUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_app_cpu_usage_percent",
			Help: "Текущее использование CPU",
		},
	)
	memUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "go_app_ram_bytes",
			Help: "Потребление памяти приложением (байты)",
		},
	)
)

func RequestRPS(handler, method, status string) {
	requestsTotal.WithLabelValues(handler, method, status).Inc()
}

func ErrorRPS(handler, method string) {
	errorsTotal.WithLabelValues(handler, method).Inc()
}
func RequestDuration(handler, method string, val float64) {
	requestDuration.WithLabelValues(handler, method).Observe(val)
}

func init() {
	prometheus.MustRegister(requestsTotal, errorsTotal, requestDuration, cpuUsage, memUsage)
}

func RecordSysMetrics() {
	var m runtime.MemStats
	for {
		runtime.ReadMemStats(&m)
		memUsage.Set(float64(m.Alloc))
		// CPU: Тут будет примерное значение (например, занятое Goroutines)
		cpuUsage.Set(float64(runtime.NumGoroutine()))
		time.Sleep(5 * time.Second)
	}
}
