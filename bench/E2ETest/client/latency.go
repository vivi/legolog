package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/huyuncong/MerkleSquare/bench"
	"github.com/huyuncong/MerkleSquare/client"
	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/core"
	"github.com/immesys/bw2/crypto"

	mathrand "math/rand"
)

const serverPort = constants.ServerPort
const serverAddr = bench.ServerAddr
const auditorAddr = bench.AuditorAddr
const baseUsername = "testuser1"

func E2EClient() {
	ctx := context.Background()

	c, _ := client.NewClient(serverAddr, auditorAddr, "")

	masterSK, masterVK := crypto.GenerateKeypair()
	_, publicVK := crypto.GenerateKeypair()

	ids := make([]string, 1000)
	keys := make([][]byte, 1000)
	pos := make([]uint64, 1000)
	vrf := make([][]byte, 1000)
	sig := make([][]byte, 1000)
	keyhash := make([][]core.KeyHash, 1000)
	for j := range ids {
		var err error
		ids[j] = ("populateuser" + strconv.Itoa(mathrand.Intn(500000-1)))
		keys[j], pos[j], vrf[j], sig[j], err = c.LookUpPKVerifyForTest(ctx, []byte(ids[j]))
		if err != nil {
			panic(err)
		}

		k, p, v, s, err2 := c.LookUpMKVerifyForTest(ctx, []byte(ids[j]))
		if err2 != nil {
			panic(err2)
		}

		keyhash[j] = make([]core.KeyHash, 0)
		keyhash[j] = core.AddKeyHash(keyhash[j], v, k, s, uint32(p))
		keyhash[j] = core.AddKeyHash(keyhash[j], vrf[j], keys[j], sig[j], uint32(pos[j]))
	}

	for iter := 0; iter < 10; iter++ {
		start := time.Now()
		for i := 0; i < 1000; i++ {
			c.Register(ctx, []byte(baseUsername+strconv.Itoa(iter)+":"+strconv.Itoa(i)), masterSK, masterVK)
		}
		fmt.Printf("Register: %v\n", time.Since(start))
		time.Sleep(time.Second * 10)

		start = time.Now()
		for i := 0; i < 1000; i++ {
			c.Append(ctx, []byte(baseUsername+strconv.Itoa(iter)+":"+strconv.Itoa(i)), publicVK)
		}
		fmt.Printf("Append: %v\n", time.Since(start))
		time.Sleep(time.Second * 10)

		// fmt.Println("Press enter after updating server epoch.")
		// input := bufio.NewScanner(os.Stdin)
		// input.Scan()

		start = time.Now()
		for i := 0; i < 1000; i++ {
			c.LookUpMK(ctx, []byte(ids[i]))
		}
		fmt.Printf("LookupMK: %v\n", time.Since(start))
		time.Sleep(time.Second * 10)

		start = time.Now()
		for i := 0; i < 1000; i++ {
			c.LookUpPK(ctx, []byte(ids[i]))
		}
		fmt.Printf("LookupPK: %v\n", time.Since(start))
		time.Sleep(time.Second * 10)

		start = time.Now()
		for i := 0; i < 1000; i++ {
			c.LookUpMKVerify(ctx, []byte(ids[i]))
		}
		fmt.Printf("LookupMKVerify: %v\n", time.Since(start))
		time.Sleep(time.Second * 10)

		start = time.Now()
		for i := 0; i < 1000; i++ {
			c.LookUpPKVerify(ctx, []byte(ids[i]))
		}
		fmt.Printf("LookupPKVerify: %v\n", time.Since(start))
		time.Sleep(time.Second * 10)

		start = time.Now()
		for i := 0; i < 1000; i++ {
			c.MonitoringForLatency(ctx, []byte(ids[i]), int(pos[i]), 0, vrf[i], keys[i], sig[i], keyhash[i])
		}
		fmt.Printf("Monitoring: %v\n", time.Since(start))
		time.Sleep(time.Second * 10)
	}
}

func main() {
	E2EClient()
}
