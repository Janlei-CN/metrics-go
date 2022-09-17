package main

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"metrics/benchmark"
	_ "metrics/benchmark"
	"net/http"
)

func main() {
	benchmark.SetUp()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
