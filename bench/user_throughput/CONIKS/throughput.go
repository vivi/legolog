package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	mathrand "math/rand"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/huyuncong/MerkleSquare/bench/user_throughput"
	conikssys "github.com/huyuncong/MerkleSquare/coniks/coniks-system"
	"github.com/huyuncong/MerkleSquare/constants"
)

//const ServerAddr = "localhost" + constants.ServerPort
const ServerAddr = "35.166.181.196" + constants.ServerPort

const NumConnectionsShift = 3
const NumConnections = (1 << NumConnectionsShift)
const NumConnectionsMask = (NumConnections - 1)

var NumIDs = 200 //runtime.NumCPU() << 4
var NumUsers = 10000000
var MonitoringInterval = 1 * 60 * 60
var EpochInterval = 60 * 60 //60 * 60
var testLength = 8 * 30
var maxMsgSize = 800000000

var numKeys = 1000000
var debug = false //true
var bound = 2 * time.Second

// const timeout = 2

// NumUserInRealWorld = 1 Billion
// #Monitoring := NumUserInRealWorld / MonitoringInterval(s)
// #Writes := 200 appends/s
// #Lookup := (NumUserInRealWorld * 43 * 0.64) / (24 * 60 * 60)

var c []*conikssys.Client
var ids []string

var wg *sync.WaitGroup

type Result struct {
	res []int64
}

var result *Result
var Type []int
var opNum []int
var Count []int

func HelperNewClient(serverAddr string, auditorAddr string, verifierAddr string) (*conikssys.Client, error) {
	return conikssys.NewClientForUserThroughput(serverAddr, maxMsgSize)
}

func helperSetupKVs(ctx context.Context) {
	ids = make([]string, NumIDs)

	// wg := &sync.WaitGroup{}
	// wg.Add(len(ids))
	// runtime.GOMAXPROCS(runtime.NumCPU())

	for j := range ids {
		// var verifierRequest *grpcint.VerifyAppendRequest
		// var appendRequest *server.KeyAppendRequest

		ids[j] = ("epochuser" + strconv.Itoa(mathrand.Intn(numKeys-1)))

		// if ids[j], err2 = bench.HelperCreateKV(ctx, c[j&NumConnectionsMask]); err2 != nil {
		// 	// if ids[i], err2 = helperCreateKV(ctx, c[i]); err2 != nil {
		// 	fmt.Println("Error on register")
		// 	panic(err2)
		// }
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

func helperSetupThroughputTest(ctx context.Context) error {
	err := helperSetupClients(ServerAddr, "", "")
	if err != nil {
		return err
	}

	helperSetupKVs(ctx)

	time.Sleep(constants.EpochDuration)
	return nil
}

func helperMeasureThroughput(ctx context.Context, numUsers int, epochInterval int, monitoringInterval int) (float64, float64, float64, float64) {

	wg = new(sync.WaitGroup)

	tot_op, workload := user_throughput.GenerateWorkLoad(numUsers, epochInterval, monitoringInterval, NumIDs, testLength)

	contents := make([]byte, 256)
	result = new(Result)
	result.res = make([]int64, tot_op)
	Type = make([]int, tot_op)
	opNum = make([]int, 3)
	Count = make([]int, tot_op)

	fmt.Printf("Total Operations: %d\n", tot_op)

	if _, err := rand.Read(contents); err != nil {
		panic(err)
	}

	cnt := 0
	for i := 0; i < testLength; i++ {
		start := time.Now()

		wg.Add(len(workload.Work[i].Ops))
		for _, op := range workload.Work[i].Ops {
			if op.Type == 0 {
				go handler(cnt, op.ID, op.PastEpoch, func(id int, pastEpoch int) error {
					_, _, err := c[id&NumConnectionsMask].Append(ctx, []byte(time.Now().String()+strconv.Itoa(mathrand.Intn(numUsers-1))), contents)
					return err
				})
			} else if op.Type == 1 {
				go handler(cnt, op.ID, op.PastEpoch, func(id int, pastEpoch int) error {
					err := c[id&NumConnectionsMask].LookUpPKForThroughput(ctx, []byte(ids[id]))
					return err
				})
			} else if op.Type == 2 {
				go handler(cnt, op.ID, op.PastEpoch, func(id int, pastEpoch int) error {
					err := c[id&NumConnectionsMask].MonitoringPKForThroughput(ctx, []byte(ids[id]), uint64(pastEpoch))
					return err
				})
			}
			Type[cnt] = op.Type
			if i >= 5 && i <= testLength-5 {
				Count[cnt] = 1
				opNum[op.Type] = opNum[op.Type] + 1
			}
			cnt = cnt + 1
		}

		if time.Since(start) < time.Second {
			time.Sleep(time.Second - time.Since(start))
		}
	}

	wg.Wait()

	if cnt != tot_op {
		panic("wrong number of ops")
	}

	var total []float64
	total = make([]float64, 3)
	for i, r := range result.res {
		total[Type[i]] = total[Type[i]] + (float64(r) * float64(Count[i]))
	}

	return total[0] / float64(opNum[0]), total[1] / float64(opNum[1]), total[2] / float64(opNum[2]), (total[0] + total[1] + total[2]) / float64(opNum[0]+opNum[1]+opNum[2])
}

func handler(resID int, id int, pastEpoch int, function func(id int, pastEpoch int) error) {
	var err error
	start := time.Now()
	if err = function(id, pastEpoch); err != nil {
		panic(err)
	}

	// fmt.Println(time.Since(start).Milliseconds())
	tmp := time.Since(start)

	// if pastEpoch == 0 && tmp > 5*time.Second {
	// 	panic("infinite latency!")
	// }

	if debug {
		if tmp > time.Second && pastEpoch == 0 {
			fmt.Printf("A or L:")
			fmt.Println(tmp)
		} else if tmp > bound {
			fmt.Printf("%d: ", resID)
			fmt.Println(tmp)
		}
	}

	result.res[resID] = tmp.Milliseconds()
	wg.Done()
}

func main() {
	fmt.Println("================")
	fmt.Println(NumUsers)
	fmt.Println(testLength)
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mathrand.Seed(time.Now().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := helperSetupThroughputTest(ctx)
	if err != nil {
		panic(err)
	}

	resA, resL, resM, res := helperMeasureThroughput(ctx, NumUsers, EpochInterval, MonitoringInterval)

	fmt.Println("finished!")
	fmt.Println(NumUsers)
	fmt.Println(testLength)
	fmt.Printf("%d %d %d\n", opNum[0], opNum[1], opNum[2])
	fmt.Printf("avg append latency: %f ms\n", resA)
	fmt.Printf("avg lookup latency: %f ms\n", resL)
	fmt.Printf("avg monitoring latency: %f ms\n", resM)
	fmt.Printf("avg latency: %f ms\n", res)
}
