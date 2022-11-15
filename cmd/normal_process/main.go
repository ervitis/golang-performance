package main

import (
	"context"
	"fmt"
	"github.com/ervitis/golang-performance/common"
	"github.com/ervitis/golang-performance/infra/metrics"
	"log"
	"net/http"
)

func main() {
	common.InitSignal()
	done := make(chan struct{})

	metricsHandler := metrics.NewMetricsHandler()
	router := http.NewServeMux()
	router.Handle(metricsHandler.Address.Url, metricsHandler.Handler)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", metricsHandler.Address.Port),
		Handler: router,
	}
	go func() {
		<-common.SignalHandler
		if err := server.Shutdown(context.Background()); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	log.Println("Starting metrics server")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	<-done
	log.Println("End metric handler")
}
