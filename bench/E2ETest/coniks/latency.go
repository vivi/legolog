package main

import (
	"context"
	"fmt"
	mathrand "math/rand"
	"os"
	"strconv"
	"time"

	conikssys "github.com/huyuncong/MerkleSquare/coniks/coniks-system"
	"github.com/huyuncong/MerkleSquare/constants"

	"github.com/immesys/bw2/crypto"
)

const serverPort = constants.ServerPort
const serverAddr = "172.31.0.252" + constants.ServerPort
const auditorAddr = "172.31.5.175" + constants.AuditorPort
const baseUsername = "appenduser"

func E2EConiksClient(numKeys int) {
	ctx := context.Background()

	c, _ := conikssys.NewClient(serverAddr, auditorAddr, "")

	ids := make([]string, 1000)
	for j := range ids {
		ids[j] = ("epochuser" + strconv.Itoa(mathrand.Intn(numKeys-1)))
	}

	_, publicVK := crypto.GenerateKeypair()
	start := time.Now()
	for i := 0; i < 1000; i++ {
		c.Append(ctx, []byte(ids[i]+strconv.Itoa(i)), publicVK)
	}
	fmt.Printf("Append: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

	pks := make([][]byte, 1000)
	start = time.Now()
	for i := 0; i < 1000; i++ {
		pks[i], _ = c.LookUpPK(ctx, []byte(ids[i]))
	}
	fmt.Printf("LookupPK: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		c.MonitoringPKForLatency(ctx, []byte(ids[i]), pks[i], 1)
	}
	fmt.Printf("MonitorPK 1: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		c.MonitoringPKForLatency(ctx, []byte(ids[i]), pks[i], 10)
	}
	fmt.Printf("MonitorPK 10: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		c.MonitoringPKForLatency(ctx, []byte(ids[i]), pks[i], 100)
	}
	fmt.Printf("MonitorPK 100: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

	start = time.Now()
	for i := 0; i < 1000; i++ {
		c.MonitoringPKForLatency(ctx, []byte(ids[i]), pks[i], 1000)
	}
	fmt.Printf("MonitorPK 1000: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

}

func main() {
	numUsers, _ := strconv.Atoi(os.Args[1])
	E2EConiksClient(numUsers)
}
