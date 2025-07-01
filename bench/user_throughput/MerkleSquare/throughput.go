package main

import (
	//	"bufio"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	mathrand "math/rand"
	//	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/huyuncong/MerkleSquare/bench/user_throughput"
	"github.com/huyuncong/MerkleSquare/client"
	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/core"
	"github.com/immesys/bw2/crypto"
)

// const ServerAddr = "localhost" + constants.ServerPort
const ServerAddr = "34.215.95.58" + constants.ServerPort

const NumConnectionsShift = 3
const NumConnections = (1 << NumConnectionsShift)
const NumConnectionsMask = (NumConnections - 1)

var NumIDs = 200 //runtime.NumCPU() << 4
var NumUsers = 8000000
var MonitoringInterval = 1 * 60 * 60
var EpochInterval = 1
var testLength = 2 * 30
var treeSize = 1000000
var maxMsgSize = 800000000

var numKeys = 1000000
var debug = false //true
var bound = 2 * time.Second

// const timeout = 2

// NumUserInRealWorld = 1 Billion
// #Monitoring := NumUserInRealWorld / MonitoringInterval(s)
// #Writes := 200 appends/s
// #Lookup := (NumUserInRealWorld * 43 * 0.64) / (24 * 60 * 60)

var c []*client.Client
var ids []string
var heights []uint32
var positions []int
var monitorOn []bool

var wg *sync.WaitGroup

type Result struct {
	res []int64
}

var result *Result
var Type []int
var opNum []int
var Count []int

func HelperNewClient(serverAddr string, auditorAddr string, verifierAddr string) (*client.Client, error) {
	return client.NewClientForUserThroughput(serverAddr, maxMsgSize)
}

// func HelperMerkleSquareCreateKV(ctx context.Context, c *client.Client) ([]byte, int, error) {
// 	// myseed := atomic.AddUint32(&seed, 1)
// 	username := make([]byte, 32)
// 	if _, err := rand.Read(username); err != nil {
// 		return nil, 0, err
// 	}
// 	masterSK, masterVK := crypto.GenerateKeypair()
// 	pos, err := c.Register(ctx, username, masterSK, masterVK)
// 	return username, int(pos), err
// }

// func HelperCreateKV(ctx context.Context, c *client.Client) (string, int, error) {
// 	var username []byte
// 	var err error
// 	var pos int
//
// 	username, pos, err = HelperMerkleSquareCreateKV(ctx, c)
//
// 	return string(username), pos, err
// }

func helperSetupKVs(ctx context.Context) {
	ids = make([]string, NumIDs)
	heights = make([]uint32, NumIDs)
	positions = make([]int, NumIDs)
	monitorOn = make([]bool, NumIDs)

	// wg := &sync.WaitGroup{}
	// wg.Add(len(ids))
	// runtime.GOMAXPROCS(runtime.NumCPU())
	contents := make([]byte, 256)
	if _, err := rand.Read(contents); err != nil {
		panic(err)
	}

	appendsOffline := float64(MonitoringInterval) * float64(NumUsers) * float64(user_throughput.AppendUserDay) / float64(user_throughput.SecondsPerDay)

	for j := range ids {
		// var err2 error
		var pos uint64
		// pos = uint64(mathrand.Intn(numKeys-1) / 2)
		ids[j] = "populateuser" + strconv.Itoa(mathrand.Intn(numKeys-1)/2)
		_, pos, _ = c[j&NumConnectionsMask].LookUpPK(ctx, []byte(ids[j]))

		var height uint32

		monitorOn[j] = true
		if int(appendsOffline) >= treeSize {
			height = 0
		} else {
			height = core.GetOldDepth(uint32(pos), uint32(treeSize-int(appendsOffline)))
			// fmt.Println(height)
			// fmt.Println(appendsOffline)
			// fmt.Println(float64(MonitoringInterval) * float64(NumUsers) * float64(user_throughput.AppendUserDay))
			// fmt.Println(float64(user_throughput.SecondsPerDay))
			// fmt.Println(MonitoringInterval)
			// fmt.Println(uint32(treeSize - int(appendsOffline)))

			tmpheight := core.GetOldDepth(uint32(pos), uint32(treeSize))
			if tmpheight == height {
				monitorOn[j] = false
			}
		}

		heights[j] = height
		positions[j] = int(pos)
		// fmt.Printf("Height: %d, Pos: %d\n", height, pos)
	}
}

func helperSetupClients(serverAddr string, auditorAddr string, verifierAddr string) error {
	c = make([]*client.Client, NumConnections)
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

	masterSK, masterVK := crypto.GenerateKeypair()

	cnt := 0
	for i := 0; i < testLength; i++ {
		start := time.Now()

		wg.Add(len(workload.Work[i].Ops))
		for _, op := range workload.Work[i].Ops {
			if op.Type == 0 {
				go handler(cnt, op.ID, func(id int) error {
					_, err := c[id&NumConnectionsMask].RegisterForThroughput(ctx, []byte(time.Now().String()+strconv.Itoa(mathrand.Intn(numUsers-1))), masterSK, masterVK)
					return err
				})
			} else if op.Type == 1 {
				go handler(cnt, op.ID, func(id int) error {
					_, _, err := c[id&NumConnectionsMask].LookUpPKVerifyForThroughput(ctx, []byte(ids[id]))
					return err
				})
			} else if op.Type == 2 {
				if monitorOn[op.ID] {
					go handler(cnt, op.ID, func(id int) error {
						err := c[id&NumConnectionsMask].MonitoringForThroughput(ctx, []byte(ids[id]), int(positions[id]), int(heights[id]))
						//fmt.Println(heights[id])
						return err
					})
				}
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

func handler(resID int, id int, function func(id int) error) {
	var err error
	start := time.Now()
	if err = function(id); err != nil {
		panic(err)
	}

	// fmt.Println(time.Since(start).Milliseconds())
	tmp := time.Since(start)

	// if pastEpoch == 0 && tmp > 5*time.Second {
	// 	panic("infinite latency!")
	// }

	if debug {
		fmt.Printf("%d: ", resID)
		fmt.Println(tmp)
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

	//	buf := bufio.NewReader(os.Stdin)
	//	fmt.Println("Press Enter to continue:")
	//	_, err = buf.ReadBytes('\n')
	//	if err != nil {
	//		panic(err)
	//	}

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
