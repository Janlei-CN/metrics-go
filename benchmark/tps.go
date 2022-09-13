package benchmark

import (
	"fmt"
	"math/rand"
	"runtime"
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

// benchmarkForTps : Compress the number of services that can be requested per second
func benchmarkForTps(urls []string) {
	responseCostChannel := make(chan *Tag, 10) // Response results are presented in a new style of time sequence
	tpsChannel := make(chan *TpsTag, 10)       // Aggregate the number of requests per second (successful requests, failed requests)

	go StatisticsAndOutput(tpsChannel)

	go Monitor(responseCostChannel, tpsChannel)

	go transcationFor(urls, responseCostChannel)

	select {} //blocking to avoid stopping benchmarkForTps
}

func StatisticsAndOutput(tpsCh <-chan *TpsTag) {
	for {
		select {
		case tpsTag, _ := <-tpsCh:
			fmt.Println("now:", time.Now())
			fmt.Println("SuccessTps", tpsTag.SuccessTps)
			fmt.Println("FailureTps", tpsTag.FailureTps)
			fmt.Println("AllTps", tpsTag.AllTps)
		default:
		}
	}
}

func transcationFor(urls []string, channel chan<- *Tag) {
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func() {
			for {
				// mock cost for http or rpc
				random := rand.Intn(10)
				time.Sleep(time.Duration(random))
				channel <- &Tag{Time: time.Now(), Pass: rand.Intn(2) == 0}
			}
		}()
	}

}

func Monitor(channel <-chan *Tag, tpsCh chan<- *TpsTag) {
	var successTps int64 = 0
	var failureTps int64 = 0
	var allTps int64 = 0

	for {
		done := make(chan bool)
		go countPerSecond(channel, &successTps, &failureTps, &allTps, done)
		<-done
		tpsCh <- &TpsTag{atomic.LoadInt64(&successTps), atomic.LoadInt64(&failureTps), atomic.LoadInt64(&allTps)}

		reset(&successTps, &failureTps, &allTps)
	}
}

func reset(successTps *int64, failureTps *int64, allTps *int64) {
	atomic.StoreInt64(successTps, 0)
	atomic.StoreInt64(failureTps, 0)
	atomic.StoreInt64(allTps, 0)
}

func countPerSecond(channel <-chan *Tag, successTps *int64, failureTps *int64, allTps *int64, done chan<- bool) {
	ticker := time.NewTicker(1 * time.Second)
	for {
		go func() {
			for {
				tag, ok := <-channel
				//fmt.Println("tag time:", tag.Time)
				if ok {
					if tag.Pass {
						atomic.AddInt64(successTps, 1)
					} else {
						atomic.AddInt64(failureTps, 1)
					}
					atomic.AddInt64(allTps, 1)
				} else {
					return
				}
			}
		}()
		// wait one second
		<-ticker.C
		done <- true
	}
}
