package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/core"
	auditor_client "github.com/huyuncong/MerkleSquare/legolog/auditor/auditorclt"
	client "github.com/huyuncong/MerkleSquare/legolog/client"
	libcrypto "github.com/huyuncong/MerkleSquare/lib/crypto"
	"github.com/immesys/bw2/crypto"
)

var masterVK = []byte{
	62, 211, 17, 153, 196, 54, 94, 151, 176, 0, 248, 70, 42, 138, 114, 235, 1, 171, 213, 145, 8, 145, 100, 71, 109, 6, 10, 238, 69, 23, 4, 136,
}

var masterSK = []byte{
	232, 197, 35, 104, 194, 130, 102, 207, 237, 150, 222, 125, 105, 185, 219, 217, 27, 243, 247, 40, 137, 252, 232, 107, 208, 104, 230, 160, 105, 179, 150, 61,
}

var NumThreads = 32

var c []*client.Client
var a []auditor_client.Client

const NumClients = 1 << 3
const NumClientsMask = NumClients - 1

var wg *sync.WaitGroup
var ids []string

func setUpClients(serverAddr string, auditorAddr string) {
	c = make([]*client.Client, NumClients)
	a = make([]auditor_client.Client, NumClients)
	var err error
	for i := range c {
		c[i], err = client.NewClient(serverAddr, "", "")
		if err != nil {
			panic(err)
		}
		a[i], err = auditor_client.NewAuditorClient(auditorAddr)
		if err != nil {
			panic(err)
		}
	}
}

func setupKVs(ctx context.Context) {
	ids = make([]string, NumThreads)
	for i := range ids {
		ids[i] = strconv.Itoa(rand.Intn(10000))
		_, err := c[i&NumClientsMask].Register(ctx, []byte(ids[i]), masterSK, masterVK)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func measureThroughput(ctx context.Context, expCfg *core.ExperimentConfig, function func(id int) error) float64 {
	startC := make(chan interface{})
	closeC := make(chan interface{})
	result := make(chan int64, NumThreads)

	wg = new(sync.WaitGroup)
	wg.Add(NumThreads)

	for i := 0; i < NumThreads; i++ {
		go handler(i, startC, closeC, result, function)
	}

	close(startC)

	time.Sleep(expCfg.TestThroughputDuration)

	close(closeC)

	var totalOps int64
	for i := 0; i < NumThreads; i++ {
		totalOps += <-result
	}
	wg.Wait()
	close(result)
	return float64(totalOps) / float64(expCfg.TestThroughputDuration.Seconds())
}

func handler(id int, startC chan interface{}, closeC chan interface{}, result chan int64, function func(id int) error) {
	var count int64
	<-startC
	for {
		var err error
		if err = function(id); err != nil {
			fmt.Println(err)
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

func GetPartitionForIdentifier(identifier []byte, partitions uint) uint {
	hash := uint(binary.BigEndian.Uint64(libcrypto.Hash(identifier)))
	return hash % partitions
}

func spaceOutAppends(ctx context.Context, cfg core.Config, expCfg core.ExperimentConfig, _c *client.Client, idx int) {
	numAppends := expCfg.WritesPerUpdatePeriod
	value := generateRandomByteArray(32)
	timeToSleep := cfg.UpdatePeriod / time.Duration(numAppends)
	for i := 0; i < int(numAppends); i++ {
		_, _, err := c[idx&NumClientsMask].Append(ctx, []byte(ids[idx]), []byte(ids[idx]), value)
		if err != nil {
			fmt.Println(err)
		}
		idx = (idx + 1) % NumThreads
		time.Sleep(timeToSleep)
	}
}

func main() {
	rand.Seed(12345)
	experimentConfigPtr := flag.String("exp_config", "../experiments/exp_configs/test.yaml", "experiment config file path")
	configPtr := flag.String("config", "../experiments/configs/test.yaml", "config file path")
	isWriterPtr := flag.Bool("write", false, "true if doing writes, false if doing reads")
	flag.Parse()

	expCfg, err := core.ParseExperimentConfig(*experimentConfigPtr)
	if err != nil {
		panic(errors.New("Failed to load experiment config: " + err.Error()))
	}

	cfg, err := core.ParseConfig(*configPtr)
	if err != nil {
		panic(errors.New("Failed to load config: " + err.Error()))
	}

	serverAddr := expCfg.ServerAddr + constants.ServerPort
	auditorAddr := expCfg.AuditorAddr + constants.AuditorPort

	ctx := context.Background()
	c, err := client.NewClient(serverAddr, "", "")
	if err != nil {
		panic(errors.New("Failed to start server client: " + err.Error()))
	}

	/*
		for i := 0; i < 1000; i++ {
			_, err := c.Register(ctx, []byte(strconv.Itoa(i)), masterSK, masterVK)
			if err != nil {
				// fmt.Println(err)
			}
		}
	*/

	setUpClients(serverAddr, auditorAddr)
	setupKVs(ctx)

	var startTime, endTime time.Time

	if *isWriterPtr {
		idx := 0
		for i := 0; i < int(expCfg.NumVerificationPeriods); i++ {
			for j := 0; j < int(cfg.VerificationPeriod/cfg.UpdatePeriod); j++ {
				startTime = time.Now()
				spaceOutAppends(ctx, cfg, expCfg, c, idx)
				endTime = time.Now()
				fmt.Println("Time to append: " + fmt.Sprint(endTime.Sub(startTime).Milliseconds()))
				idx += int(expCfg.WritesPerUpdatePeriod)
				idx = idx % NumThreads
			}
		}
	} else {
		/*
			auditorClient, err := auditor_client.NewAuditorClient(auditorAddr)
			if err != nil {
				panic(errors.New("Failed to start auditor client: " + err.Error()))
			}
		*/
		res := measureThroughput(ctx, &expCfg, func(i int) error {
			id := ids[i]
			partition := GetPartitionForIdentifier([]byte(id), uint(cfg.Partitions))
			digest, _, _, err := a[i&NumClientsMask].GetEpochUpdateForPartition(ctx, uint64(partition))
			if err != nil {
				fmt.Println(err)
			}
			var proof core.LegologExistenceProof
			val, pos, sig, proofJson, err := c.LookUpPKVerify(ctx, []byte("0"), []byte(id))
			if err != nil {
				fmt.Println("pos", pos)
				fmt.Println(err)
			}
			err = json.Unmarshal(proofJson, &proof)
			if err != nil {
				fmt.Println(err)
			}
			var v core.AggHistVerifier
			ok, err := v.ValidatePKProof(digest, &proof, []byte(id), val, sig, pos, masterVK)
			if !ok || err != nil {
				fmt.Println(i, err)
			}
			return nil
		})
		fmt.Println("throughput: ", res, "ops/s")
	}
	return

	aliceUsername := []byte("alice")
	aliceIdentifier1 :=
		[]byte("alice_key")
	masterSK, masterVK := crypto.GenerateKeypair()
	_, aliceVK1 := crypto.GenerateKeypair()
	/*
		Register Alice.
	*/
	pos, err := c.Register(ctx, aliceUsername, masterSK, masterVK)
	if err != nil {
		panic(errors.New("Failed to register first user: " + err.Error()))
	}
	if pos != 0 {
		panic(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	/*
		Start adding several identifier-value pairs for Alice.
	*/
	for i := 0; i < 10000; i++ {
		aliceId1Pos, _, err := c.Append(ctx, aliceUsername, aliceIdentifier1, aliceVK1)
		if err != nil {
			panic(err)
		}
		fmt.Println(aliceId1Pos)
	}

	time.Sleep(10 * time.Second)
}

func generateRandomByteArray(size int) []byte {
	res := make([]byte, size)

	_, err := rand.Read(res)
	if err != nil {
		panic(err)
	}

	return res
}
