package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

type Address struct {
	Url  string
	Port int
}

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
