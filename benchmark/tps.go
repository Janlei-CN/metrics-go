package benchmark

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	sdk "go.mpcvault.com/mpcvault-go-sdk"
	"go.mpcvault.com/mpcvault-go-sdk/proto/mpcvault/cloudmpc/v1"
	"golang.org/x/net/context"
	"net/http"
	"sync/atomic"
	"time"
)

type RequestStatus struct {
	Timestamp time.Time
	Success   bool
}

type AggregateResult struct {
	SuccessTPS int64
	FailureTPS int64
	TotalTPS   int64
}

var (
	//success
	SuccessRequestGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "request_success_num_total",
			Help: "The total number of processed events",
		},
		[]string{"name"})

	//failure
	FailureRequestGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "request_failure_num_total",
		},
		[]string{"name"})

	// total
	TotalRequestGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "request_all_num_total",
		},
		[]string{"name"})
)

var mpcvault *sdk.API

func initSdk() {
	mpcvault = &sdk.API{}
	mpcvault.PrintRequestLog = false
	err := mpcvault.SetUp(apiKey, privateKey, "1234qwer")
	if err != nil {
		panic(err)
	}
}

func SetUp() {
	prometheus.MustRegister(SuccessRequestGauge, FailureRequestGauge, TotalRequestGauge)
	initSdk()
	go StartTesting()
}

// tps : Compress the number of services that can be requested per second
func StartTesting() {
	responseCostChannel := make(chan *RequestStatus, 10000) // Response results are presented in a new style of time sequence
	tpsChannel := make(chan *AggregateResult, 10)           // Aggregate the number of requests per second (successful requests, failed requests)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go gatherIntoGauge(tpsChannel)

	go statisticsPerSecond(ctx, responseCostChannel, tpsChannel)

	go sendRequestByLoop(ctx, responseCostChannel)

	select {} //blocking to avoid stopping tps
	//time.Sleep(10 * time.Second)
}

func gatherIntoGauge(tpsCh <-chan *AggregateResult) {
	for {
		select {
		case tpsTag, ok := <-tpsCh:
			if ok {
				if mpcvault.PrintRequestLog {
					fmt.Println("now:", time.Now())
					fmt.Println("SuccessTPS", tpsTag.SuccessTPS)
					fmt.Println("FailureTPS", tpsTag.FailureTPS)
					fmt.Println("TotalTPS", tpsTag.TotalTPS)
				}

				// prometheus metrics gather
				SuccessRequestGauge.With(prometheus.Labels{
					"name": "SuccessTPS",
				}).Set(float64(tpsTag.SuccessTPS))
				FailureRequestGauge.With(prometheus.Labels{
					"name": "FailureTPS",
				}).Set(float64(tpsTag.FailureTPS))
				TotalRequestGauge.With(prometheus.Labels{
					"name": "TotalTPS",
				}).Set(float64(tpsTag.TotalTPS))
			}
		default:
		}
	}
}

func sendRequestByLoop(cxt context.Context, channel chan<- *RequestStatus) {
	defer close(channel)
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

				select {
				case <-cxt.Done():
					return
				case channel <- &RequestStatus{Timestamp: time.Now(), Success: SendRequest() == nil}:
				}
			}()
		}
	}
}

func statisticsPerSecond(cxt context.Context, channel <-chan *RequestStatus, tpsCh chan<- *AggregateResult) {
	defer close(tpsCh)

	var successTps int64 = 0
	var failureTps int64 = 0
	var allTps int64 = 0
	ticker := time.NewTicker(1 * time.Second)

	for {
		select {
		case <-cxt.Done():
			return

		case <-ticker.C:
			tpsCh <- &AggregateResult{atomic.LoadInt64(&successTps), atomic.LoadInt64(&failureTps), atomic.LoadInt64(&allTps)}
			reset(&successTps, &failureTps, &allTps)

		case tag, ok := <-channel:
			//fmt.Println("tag time:", tag.Timestamp)
			if ok {
				if tag.Success {
					atomic.AddInt64(&successTps, 1)
				} else {
					atomic.AddInt64(&failureTps, 1)
				}
				atomic.AddInt64(&allTps, 1)
			}
		}
	}
}

func SendRequest() (err error) {
	_, err = http.Get("https://www.baidu.com")
	return
}

// TestGenerateWalletAddress
func testGenerateWalletAddress() error {
	idempotentKey := uuid.NewString()
	_, err := mpcvault.CloudMPCClient.CreateKey(
		context.Background(),
		&cloudmpc.CreateKeyRequest{
			KeyType: cloudmpc.KeyType_KEYTYPE_ECC_SECP256K1,
		},
		// setting idempotent key using sdk.NewIdempotentRequestCallOption
		// if you reuse the same idempotent key within the first 24 hours of making the reqeust,
		//you get back the same response and by extension, the same key
		sdk.NewIdempotentRequestCallOption(idempotentKey),
	)
	return err
}

// TestDescribeKeyAndGenerateWalletAddress
func testDescribeKeyAndGenerateWalletAddress() error {
	_, err := mpcvault.CloudMPCClient.DescribeKey(context.Background(), &cloudmpc.DescribeKeyRequest{
		KeyId: "793a91b7-fa8d-4578-bb2e-8d008987a01d",
	})
	return err
}

//TestSignAndVerify
func testSignAndVerify() error {
	// Sign message
	message := []byte("test-message")
	keyID := "793a91b7-fa8d-4578-bb2e-8d008987a01d"
	_, err := mpcvault.CloudMPCClient.Sign(context.Background(), &cloudmpc.SignRequest{
		KeyId:       keyID,
		SigningAlgo: cloudmpc.SigningAlgo_SIGNINGALGO_ECDSA,
		Message:     message,
	})
	return err
}

func reset(successTps *int64, failureTps *int64, allTps *int64) {
	atomic.StoreInt64(successTps, 0)
	atomic.StoreInt64(failureTps, 0)
	atomic.StoreInt64(allTps, 0)
}

// Initialization variables (please change the following value to your apikey and private key
var apiKey = "FtCeDztyafURQcYC5wpvouqXsxvgxPqt4thdhf9a7u4="

var privateKey = `
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAACmFlczI1Ni1jdHIAAAAGYmNyeXB0AAAAGAAAABBxUORWXl
z0HeGDdvfKJ1DjAAAAZAAAAAEAAAAzAAAAC3NzaC1lZDI1NTE5AAAAIDcjhEh4X89v8gHT
MYRR3r7Jxd/fQuU7ZD9pMQ5EIL53AAAAoOJFbcFWr9bTfnWixq2Ucrr/uwzumGxlOxpgQK
8TOzY43rELvlSgCC6wnJb0hk+H2iD1sREfR1xEPwcxOwZLdYm+7maIxotzUKtRnHJGOEDC
PtBBXIzUrD1TPvMRjlUst9aw017xf1zrQiL6grVsu3Um4Lniq3orWYT92pRgh0iw48M5MV
ej9RRBh1dhR52V74k8CHyDVPpV3mLTtMFG00g=
-----END OPENSSH PRIVATE KEY-----
`
