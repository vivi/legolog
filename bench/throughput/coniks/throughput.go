package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	mathrand "math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	conikssys "github.com/huyuncong/MerkleSquare/coniks/coniks-system"
	"github.com/huyuncong/MerkleSquare/constants"
)

// const ServerAddr = "localhost" + constants.ServerPort
// const ServerAddr = "172.31.0.252" + constants.ServerPort
const ServerAddr = "34.215.95.58" + constants.ServerPort

const NumConnectionsShift = 3
const NumConnections = (1 << NumConnectionsShift)
const NumConnectionsMask = (NumConnections - 1)

var timeOut = flag.Int("t", 60, "timeout second")
var NumThreads = 328 //runtime.NumCPU() << 4

// const timeout = 2

// NumUserInRealWorld = 1 Billion
// #Monitoring := NumUserInRealWorld / MonitoringInterval(s)
// #Writes := 200 appends/s
// #Lookup := (NumUserInRealWorld * 43 * 0.64) / (24 * 60 * 60)

var c []*conikssys.Client
var ids []string
var contents = make([]byte, 256)

// var maxMsgSize = 800000000
var wg *sync.WaitGroup

func HelperNewClient(serverAddr string, auditorAddr string, verifierAddr string) (*conikssys.Client, error) {
	// return conikssys.NewClientForUserThroughput(serverAddr, maxMsgSize)
	return conikssys.NewClient(serverAddr, auditorAddr, verifierAddr)
}

func helperSetupKVs(ctx context.Context, numKeys int) {
	ids = make([]string, NumThreads)
	for j := range ids {
		ids[j] = ("epochuser" + strconv.Itoa(mathrand.Intn(numKeys-1)))
	}
}

func helperSetupClients(serverAddr string, auditorAddr string, verifierAddr string) error {
	c = make([]*conikssys.Client, NumConnections)
	// c = make([]*client.Client, NumThreads)
	var err error
	for i := range c {
		if c[i], err = HelperNewClient(serverAddr, auditorAddr, verifierAddr); err != nil {
			panic(err)
		}
	}
	return nil
}

func helperSetupThroughputTest(ctx context.Context, numKeys int) error {
	err := helperSetupClients(ServerAddr, "", "")
	if err != nil {
		return err
	}

	helperSetupKVs(ctx, numKeys)

	time.Sleep(constants.EpochDuration)
	return nil
}

func helperMeasureThroughput(ctx context.Context, numThreads int, function func(id int) error) float64 {
	startC := make(chan interface{})
	closeC := make(chan interface{})
	result := make(chan int64, NumThreads)

	wg = new(sync.WaitGroup)
	wg.Add(numThreads)

	if _, err := rand.Read(contents); err != nil {
		panic(err)
	}

	for i := 0; i < NumThreads; i++ {
		go handler(i, startC, closeC, result, function)
	}

	close(startC)

	time.Sleep(time.Duration(*timeOut) * time.Second)
	// stop
	close(closeC)

	var totalOps int64
	for i := 0; i < NumThreads; i++ {
		totalOps += <-result
	}

	wg.Wait()
	close(result)

	return float64(totalOps) / float64(*timeOut)
}

func handler(id int, startC chan interface{}, closeC chan interface{}, result chan int64, function func(id int) error) {
	var count int64
	//buf := make([]byte, 2*(*msgLen))

	<-startC

	for {
		var err error
		if err = function(id); err != nil {
			panic(err)
		}
		select {
		case <-closeC:
			result <- count
			wg.Done()
			return
		default:
			count++
		}
	}
}

func main() {
	numUsers, _ := strconv.Atoi(os.Args[1])
	testType, _ := strconv.Atoi(os.Args[2])
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runtime.GOMAXPROCS(runtime.NumCPU())

	err := helperSetupThroughputTest(ctx, numUsers)
	if err != nil {
		panic(err)
	}

	var res float64

	fmt.Printf("%d %d\n", *timeOut, NumThreads)
	if testType == 0 {
		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			_, _, err := c[id&NumConnectionsMask].Append(ctx, []byte(time.Now().String()), contents)
			return err
		})
		fmt.Printf("Append Throughtput: %f(ops/s)\n", res)
	} else if testType == 1 {
		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			err := c[id&NumConnectionsMask].LookUpPKForThroughput(ctx, []byte(ids[id]))
			return err
		})
		fmt.Printf("LookUp Throughtput: %f(ops/s)\n", res)
	} else if testType == 2 {
		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			err := c[id&NumConnectionsMask].MonitoringPKForThroughput(ctx, []byte(ids[id]), 1)
			return err
		})
		fmt.Printf("Monitor Throughtput for last 1 epoch: %f(ops/s)\n", res)
	} else if testType == 3 {
		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			err := c[id&NumConnectionsMask].MonitoringPKForThroughput(ctx, []byte(ids[id]), 10)
			return err
		})
		fmt.Printf("Monitor Throughtput for last 10 epochs: %f(ops/s)\n", res)
	} else if testType == 4 {
		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			err := c[id&NumConnectionsMask].MonitoringPKForThroughput(ctx, []byte(ids[id]), 100)
			return err
		})
		fmt.Printf("Monitor Throughtput for last 100 epochs: %f(ops/s)\n", res)
	} else if testType == 5 {
		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			err := c[id&NumConnectionsMask].MonitoringPKForThroughput(ctx, []byte(ids[id]), 1000)
			return err
		})
		fmt.Printf("Monitor Throughtput for last 1000 epochs: %f(ops/s)\n", res)
	}
	//}
}
