package main

import (
	"context"
	"fmt"
	"github.com/ervitis/golang-performance/common"
	"github.com/ervitis/golang-performance/infra/metrics"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func main() {
	done := make(chan struct{})

	metricsHandler := metrics.NewMetricsHandler()
	router := http.NewServeMux()
	router.Handle(metricsHandler.Address.Url, metricsHandler.Handler)

	processMetrics := metrics.NewProcessTimeMetric(metrics.ExecutionTimeName)

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", metricsHandler.Address.Port),
		Handler: router,
	}
	defer close(done)
	go func() {
		<-common.SignalHandler.InterruptSignal
		log.Println("Shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println(err)
		}
		done <- struct{}{}
	}()

	go func() {
		log.Println("Starting metrics server")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		for {
			start := time.Now()
			run()
			end := time.Now()
			processMetrics.Set(end.Sub(start).Seconds())
			const min, max = 3, 11
			rndWait := rand.New(rand.NewSource(time.Now().Unix())).Intn(max-min) + min
			time.Sleep(time.Duration(rndWait) * time.Second)
		}
	}()

	<-done
}

func run() {
	if err := os.MkdirAll("/tmp/gochannel", 0776); err != nil {
		panic(err)
	}

	fo := common.NewFileOperatorChannel("/tmp/gochannel/myfulldatawithchannels.csv")

	done := make(chan struct{})

	input := fo.Read(fo.OpenFiles("data"))
	fo.Write(done, input)
	<-done
}
