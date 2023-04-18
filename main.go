package main

import (
	"github.com/Janlei-CN/go-metrics/benchmark"
	_ "github.com/Janlei-CN/go-metrics/benchmark"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	benchmark.SetUp()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
