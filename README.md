# Installation

Make sure your project is using Go Modules (it will have a `go.mod` file in its root if it already is):

```bash
go mod init
```

Benchmark test to get tps, please click for [details](benchmark/tps_test.go)


# Usage
## prometheus
Using Prometheus to collect display timing data, you need to prioritize the deployment of the configuration service to expose the collection information.
1. Create a new directory and generate prometheus.yaml
```shell
mkdir /opt/prometheus
cd /opt/prometheus/
vim prometheus.yml
```
2. Prometheus yaml, please update the prometheus configuration to set your IP.
```shell
global:
  scrape_interval:     60s
  evaluation_interval: 60s

scrape_configs:
  - job_name: prometheus
    static_configs:
      - targets: ['ip:9090']
        labels:
          instance: prometheus

  - job_name: linux
    static_configs:
      - targets: ['ip:9100']
        labels:
          instance: localhost

  - job_name: metrics-go
    static_configs:
      - targets: ['ip:2112']
        labels:
          instance: metrics-go

```

## Start Service Order
```shell
# 1.client
go run main.go

# 2.docker prometheus 
docker run  -d \
    -p 9090:9090 \
    -v /opt/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
```
## Prometheus UI
Please click http://ip:9090, view node acquisition timing data graphs