package main

import (
	"github.com/Janlei-CN/metrics-go/benchmark"
	_ "github.com/Janlei-CN/metrics-go/benchmark"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func main() {
	benchmark.SetUp()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
