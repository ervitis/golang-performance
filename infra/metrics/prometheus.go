package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type Address struct {
	Url  string
	Port int
}

const (
	ExecutionTimeName = "process_time"
)

type Metrics struct {
	Handler http.Handler
	Address Address
}

func NewMetricsHandler() Metrics {
	return Metrics{
		Handler: promhttp.Handler(),
		Address: Address{
			Url:  "/metrics",
			Port: 2112,
		},
	}
}

func NewProcessTimeMetric(namespace string) prometheus.Gauge {
	h := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Name:      "execution_time",
		Help:      "Execution process time",
	})

	prometheus.MustRegister(h)
	return h
}
