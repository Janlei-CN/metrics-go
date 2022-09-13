package main

import (
	"github.com/prometheus/client_golang/prometheus/promauto"
	"metrics/benchmark"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "metrics/benchmark"
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func main() {
	recordMetrics()
	go benchmark.Tps()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}