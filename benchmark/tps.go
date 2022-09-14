package benchmark

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/net/context"
	"math/rand"
	"sync/atomic"
	"time"
)

type Tag struct {
	Time time.Time
	Pass bool
}

type TpsTag struct {
	SuccessTps int64
	FailureTps int64
	AllTps     int64
}

var (
	QueueGauges = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "request_success_num_total",
			Help: "The total number of processed events",
		},
		[]string{"name"})

	QueueGaugef = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "request_failure_num_total",
		},
		[]string{"name"})

	QueueGaugeA = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "request_all_num_total",
		},
		[]string{"name"})
)

func init() {
	prometheus.MustRegister(QueueGauges, QueueGaugef, QueueGaugeA)
}

// tps : Compress the number of services that can be requested per second
func Tps() {
	responseCostChannel := make(chan *Tag, 10000) // Response results are presented in a new style of time sequence
	tpsChannel := make(chan *TpsTag, 10)          // Aggregate the number of requests per second (successful requests, failed requests)

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		close(responseCostChannel)
		close(tpsChannel)
	}()

	go StatisticsAndOutput(tpsChannel)

	go Monitor(ctx, responseCostChannel, tpsChannel)

	go transcationFor(ctx, responseCostChannel)

	select {} //blocking to avoid stopping tps
	//time.Sleep(10 * time.Second)
}

func StatisticsAndOutput(tpsCh <-chan *TpsTag) {
	for {
		select {
		case tpsTag, ok := <-tpsCh:
			if ok {
				fmt.Println("now:", time.Now())
				fmt.Println("SuccessTps", tpsTag.SuccessTps)
				fmt.Println("FailureTps", tpsTag.FailureTps)
				fmt.Println("AllTps", tpsTag.AllTps)

				// prometheus metrics gather
				QueueGauges.With(prometheus.Labels{
					"name": "SuccessTps",
				}).Set(float64(tpsTag.SuccessTps))
				QueueGaugef.With(prometheus.Labels{
					"name": "FailureTps",
				}).Set(float64(tpsTag.FailureTps))
				QueueGaugeA.With(prometheus.Labels{
					"name": "AllTps",
				}).Set(float64(tpsTag.AllTps))
			}
		default:
		}
	}
}

func transcationFor(cxt context.Context, channel chan<- *Tag) {
	limit := make(chan struct{}, 50)
	for {
		select {
		case <-cxt.Done():
			return

		case limit <- struct{}{}:
			go func() {
				defer func() {
					<-limit
				}()

				//err := mockReq()
				select {
				case <-cxt.Done():
					return
				case channel <- &Tag{Time: time.Now(), Pass: rand.Intn(10) == 0}:
				}
			}()
		}
	}
}

func Monitor(cxt context.Context, channel <-chan *Tag, tpsCh chan<- *TpsTag) {
	var successTps int64 = 0
	var failureTps int64 = 0
	var allTps int64 = 0
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-cxt.Done():
			return

		case <-ticker.C:
			tpsCh <- &TpsTag{atomic.LoadInt64(&successTps), atomic.LoadInt64(&failureTps), atomic.LoadInt64(&allTps)}
			reset(&successTps, &failureTps, &allTps)

		case tag, ok := <-channel:
			//fmt.Println("tag time:", tag.Time)
			if ok {
				if tag.Pass {
					atomic.AddInt64(&successTps, 1)
				} else {
					atomic.AddInt64(&failureTps, 1)
				}
				atomic.AddInt64(&allTps, 1)
			}
		}
	}
}

func reset(successTps *int64, failureTps *int64, allTps *int64) {
	atomic.StoreInt64(successTps, 0)
	atomic.StoreInt64(failureTps, 0)
	atomic.StoreInt64(allTps, 0)
}
