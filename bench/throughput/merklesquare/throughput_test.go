package main

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

// const NumConnectionsShift = 3
// const NumConnections = (1 << NumConnectionsShift)
// const NumConnectionsMask = (NumConnections - 1)

// NumUserInRealWorld = 1 Billion
// #Monitoring := NumUserInRealWorld / MonitoringInterval(s)
// #Writes := 200 appends/s
// #Lookup := (NumUserInRealWorld * 43 * 0.64) / (24 * 60 * 60)

func helperMeasureThroughputTest(b *testing.B, numThreads int, operation func(id int) error) {
	//fmt.Println("start!")
	done := make(chan struct{}, 1024000)
	var finished uint32

	runtime.GOMAXPROCS(runtime.NumCPU())

	wg := new(sync.WaitGroup)
	wg.Add(numThreads)
	for i := 0; i != numThreads; i++ {
		go func(id int) {
			for atomic.LoadUint32(&finished) == 0 {
				var err error
				if err = operation(id); err != nil {
					b.Fatal(err)
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

func helperBenchmarkThroughputTest(ctx context.Context, b *testing.B, function func(id int) error) {
	b.StopTimer()
	// fmt.Println(runtime.NumCPU())

	err := helperSetupThroughputTest(ctx)
	if err != nil {
		b.Error(err)
	}

	helperMeasureThroughputTest(b, NumThreads, function)
}

func BenchmarkLookUpPKVerifyThroughput(b *testing.B) {
	// fmt.Println(runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	helperBenchmarkThroughputTest(ctx, b, func(id int) error {
		_, _, err := c[id&NumConnectionsMask].LookUpPKVerifyForThroughput(ctx, []byte(ids[id]))
		return err
	})
}

func BenchmarkLookUpMKVerifyThroughput(b *testing.B) {
	// fmt.Println(runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	helperBenchmarkThroughputTest(ctx, b, func(id int) error {
		_, _, err := c[id&NumConnectionsMask].LookUpMKVerifyForThroughput(ctx, []byte(ids[id]))
		return err
	})
}

func BenchmarkMonitoringThroughput(b *testing.B) {
	// fmt.Println(runtime.NumCPU())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	helperBenchmarkThroughputTest(ctx, b, func(id int) error {
		err := c[id&NumConnectionsMask].MonitoringForThroughput(ctx, []byte(ids[id]), int(pos[id]), 0)
		return err
	})
}
