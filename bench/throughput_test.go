package bench

import (
	"context"
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/huyuncong/MerkleSquare/verifier/verifierd"

	"github.com/huyuncong/MerkleSquare/client"
	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/grpcint"
)

const NumConnectionsShift = 3
const NumConnections = (1 << NumConnectionsShift)
const NumConnectionsMask = (NumConnections - 1)

const AppendPercent = 0.003
const MonitoringPercent = 0.158
const LookupPercent = 0.839

// NumUserInRealWorld = 1 Billion
// #Monitoring := NumUserInRealWorld / MonitoringInterval(s)
// #Writes := 200 appends/s
// #Lookup := (NumUserInRealWorld * 43 * 0.64) / (24 * 60 * 60)

var NumThreads = 35 //runtime.NumCPU() << 4

var c []*client.Client
var v []*verifierd.Verifier
var ids []string
var monitoringRequests []*verifierd.KeyAppendRequest
var contents []byte

func helperSetupKVs(ctx context.Context, size uint64) {
	var err error
	ids = make([]string, NumThreads)
	monitoringRequests = make([]*verifierd.KeyAppendRequest, NumThreads)
	contents = make([]byte, size)

	if _, err = rand.Read(contents); err != nil {
		panic(err)
	}

	// wg := &sync.WaitGroup{}
	// wg.Add(len(ids))
	// runtime.GOMAXPROCS(runtime.NumCPU())

	for j := range ids {
		var err2 error
		var verifierRequest *grpcint.VerifyAppendRequest
		// var appendRequest *server.KeyAppendRequest

		if ids[j], err2 = HelperCreateKV(ctx, c[j&NumConnectionsMask]); err2 != nil {
			// if ids[i], err2 = helperCreateKV(ctx, c[i]); err2 != nil {
			fmt.Println("Error on register")
			panic(err2)
		}

		if _, verifierRequest, err2 = c[j&NumConnectionsMask].Append(ctx, []byte(ids[j]), contents); err2 != nil {
			// if _, err2 = c[i].Append(ctx, []byte(ids[i]), contents); err2 != nil {
			fmt.Println("Error on append")
			panic(err2)
		}

		if monitoringRequests[j], err2 = verifierd.GetBenchmarkAppendRequest(ctx, verifierRequest); err2 != nil {
			panic(err2)
		}
	}
}

func helperSetupClients(serverAddr string, auditorAddr string, verifierAddr string) error {
	c = make([]*client.Client, NumConnections)
	// c = make([]*client.Client, NumThreads)
	v = make([]*verifierd.Verifier, NumConnections)
	var err error
	for i := range c {
		if c[i], err = HelperNewClient(serverAddr, auditorAddr, verifierAddr); err != nil {
			panic(err)
		}
		if v[i], err = HelperNewVerifier(serverAddr); err != nil {
			panic(err)
		}
	}
	return nil
}

func helperSetupThroughputTest(ctx context.Context) error {
	err := helperSetupClients(ServerAddr, "", "")
	if err != nil {
		return err
	}

	helperSetupKVs(ctx, 256)

	time.Sleep(constants.EpochDuration)
	return nil
}

func helperMeasureThroughput(b *testing.B, numThreads int, operation1 func(id int) error, operation2 func(id int) error, operation3 func(id int) error) {
	//fmt.Println("start!")
	done := make(chan struct{}, 102400)
	var finished uint32

	runtime.GOMAXPROCS(runtime.NumCPU())

	wg := new(sync.WaitGroup)
	wg.Add(numThreads)
	for i := 0; i != numThreads; i++ {
		go func(id int) {
			source := mathrand.NewSource(int64(id))
			generator := mathrand.New(source)
			for atomic.LoadUint32(&finished) == 0 {
				var err error
				tmp := generator.Float64()
				if tmp < AppendPercent {
					if err = operation1(id); err != nil {
						panic(err)
					}
				} else if tmp < MonitoringPercent+AppendPercent {
					if err = operation2(id); err != nil {
						panic(err)
					}
				} else if tmp < LookupPercent+AppendPercent+MonitoringPercent {
					if err = operation3(id); err != nil {
						panic(err)
					}
				}
				done <- struct{}{}
			}
			wg.Done()
		}(i)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		<-done
	}
	b.StopTimer()

	atomic.StoreUint32(&finished, 1)

	// Wait for everything to finish, and drain the channel
	go func() {
		for _ = range done {
		}
	}()
	wg.Wait()
	close(done)
}

func helperBenchmarkMixThroughput(b *testing.B) {
	b.StopTimer()
	// fmt.Println(runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := helperSetupThroughputTest(ctx)
	if err != nil {
		b.Error(err)
	}

	helperMeasureThroughput(b, NumThreads, func(id int) error {
		_, _, err := c[id&NumConnectionsMask].Append(ctx, []byte(ids[id]), contents)
		return err
	},
		func(id int) error {
			err := v[id&NumConnectionsMask].VerifyBenchmarkAppend(ctx, monitoringRequests[id])
			return err
		},
		func(id int) error {
			_, _, err := c[id&NumConnectionsMask].LookUpPK(ctx, []byte(ids[id]))
			return err
		})
}

func BenchmarkMixThroughput(b *testing.B) {
	// fmt.Println(runtime.NumCPU())
	helperBenchmarkMixThroughput(b)
}
