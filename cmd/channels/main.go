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
		log.Println("Shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	go func() {
		log.Println("Starting metrics server")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	go func() {
		for {
			select {
			case <-common.SignalHandler:
				log.Println("Killing...")
			default:
				run()
				const min, max = 3, 8
				rndWait := rand.New(rand.NewSource(time.Now().Unix())).Intn(max-min) + min
				time.Sleep(time.Duration(rndWait) * time.Second)
			}
		}
	}()

	<-done
	log.Println("End process")
}

func run() {
	if err := os.MkdirAll("/tmp/gochannel", 0776); err != nil {
		panic(err)
	}

	fo := common.NewFileOperatorChannel("/tmp/gochannel/myfulldata.csv")

	done := make(chan struct{})

	input := fo.Read(fo.OpenFiles("data"))
	fo.Write(done, input)
	<-done
}
