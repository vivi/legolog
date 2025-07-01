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

	"github.com/huyuncong/MerkleSquare/bench"
	"github.com/huyuncong/MerkleSquare/client"
	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/verifier/verifierd"
	"github.com/immesys/bw2/crypto"
)

// const ServerAddr = "localhost" + constants.ServerPort

// const ServerAddr = "172.31.0.252" + constants.ServerPort
const ServerAddr = "34.215.95.58" + constants.ServerPort

const NumConnectionsShift = 3
const NumConnections = (1 << NumConnectionsShift)
const NumConnectionsMask = (NumConnections - 1)

var timeOut = flag.Int("t", 60, "timeout second")
var NumThreads = 328 //runtime.NumCPU() << 4
var Height = 16

// const timeout = 2

// NumUserInRealWorld = 1 Billion
// #Monitoring := NumUserInRealWorld / MonitoringInterval(s)
// #Writes := 200 appends/s
// #Lookup := (NumUserInRealWorld * 43 * 0.64) / (24 * 60 * 60)

var c []*client.Client
var v []*verifierd.Verifier
var ids []string
var pos []uint64

//var monitoringRequests []*server.KeyAppendRequest
var contents []byte

var wg *sync.WaitGroup

func helperSetupKVs(ctx context.Context, size uint64) {
	var err error
	ids = make([]string, NumThreads)
	pos = make([]uint64, NumThreads)
	contents = make([]byte, size)

	if _, err = rand.Read(contents); err != nil {
		panic(err)
	}

	// wg := &sync.WaitGroup{}
	// wg.Add(len(ids))
	// runtime.GOMAXPROCS(runtime.NumCPU())

	for j := range ids {
		// var err2 error
		// var verifierRequest *grpcint.VerifyAppendRequest
		// var appendRequest *server.KeyAppendRequest

		ids[j] = "populateuser" + strconv.Itoa(mathrand.Intn(500000))
		// ids[j] = "populateuser" + strconv.Itoa(2)
		_, pos[j], _ = c[j&NumConnectionsMask].LookUpMK(ctx, []byte(ids[j]))
		/*
			if ids[j], err2 = bench.HelperCreateKV(ctx, c[j&NumConnectionsMask]); err2 != nil {
				// if ids[i], err2 = helperCreateKV(ctx, c[i]); err2 != nil {
				fmt.Println("Error on register")
				panic(err2)
			}

					if _, verifierRequest, err2 = c[j&NumConnectionsMask].Append(ctx, []byte(ids[j]), contents); err2 != nil {
						// if _, err2 = c[i].Append(ctx, []byte(ids[i]), contents); err2 != nil {
						fmt.Println("Error on append")
						panic(err2)
					}

				if monitoringRequests[j], err2 = server.GetBenchmarkAppendRequest(ctx, verifierRequest); err2 != nil {
					panic(err2)
				}
		*/
	}
}

func helperSetupClients(serverAddr string, auditorAddr string, verifierAddr string) error {
	c = make([]*client.Client, NumConnections)
	// c = make([]*client.Client, NumThreads)
	// v = make([]*server.VerifierServer, NumConnections)
	var err error
	for i := range c {
		if c[i], err = bench.HelperNewClient(serverAddr, auditorAddr, verifierAddr); err != nil {
			panic(err)
		}
		// if v[i], err = bench.HelperNewVerifier(serverAddr); err != nil {
		// 	panic(err)
		// }
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

func helperMeasureThroughput(ctx context.Context, numThreads int, function func(id int) error) float64 {
	startC := make(chan interface{})
	closeC := make(chan interface{})
	result := make(chan int64, NumThreads)

	wg = new(sync.WaitGroup)
	wg.Add(numThreads)

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
	testType, _ := strconv.Atoi(os.Args[1])
	runtime.GOMAXPROCS(runtime.NumCPU())

	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := helperSetupThroughputTest(ctx)
	if err != nil {
		panic(err)
	}

	var res float64
	masterSK, masterVK := crypto.GenerateKeypair()

	fmt.Printf("%d %d %d\n", *timeOut, NumThreads, Height)
	if testType == 0 {
		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			_, err := c[id&NumConnectionsMask].RegisterForThroughput(ctx, []byte(time.Now().String()), masterSK, masterVK)
			return err
		})
		fmt.Printf("Register Throughtput: %f(ops/s)\n", res)
	} else if testType == 1 {
		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			err := c[id&NumConnectionsMask].MonitoringForThroughput(ctx, []byte(ids[id]), int(pos[id]), Height)
			return err
		})
		fmt.Printf("Monitoring Throughtput: %f(ops/s)\n", res)
	} else if testType == 2 {

		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			_, _, err := c[id&NumConnectionsMask].LookUpMKVerifyForThroughput(ctx, []byte(ids[id]))
			return err
		})
		fmt.Printf("LookupMKVerify Throughtput: %f(ops/s)\n", res)
	} else if testType == 3 {
		res = helperMeasureThroughput(ctx, NumThreads, func(id int) error {
			_, _, err := c[id&NumConnectionsMask].LookUpPKVerifyForThroughput(ctx, []byte(ids[id]))
			return err
		})
		fmt.Printf("LookupPKVerify Throughtput: %f(ops/s)\n", res)
	}
	fmt.Println()
}
