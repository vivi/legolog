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
const serverAddr = "34.215.95.58" + constants.ServerPort
const auditorAddr = "34.216.196.129" + constants.AuditorPort
const baseUsername = "appenduser"

func E2EConiksClient(numKeys int) {
	ctx := context.Background()

	c, _ := conikssys.NewClient(serverAddr, auditorAddr, "")

	ids := make([]string, 1000)
	for j := range ids {
		ids[j] = ("epochuser" + strconv.Itoa(mathrand.Intn(numKeys-1)))
	}

	_, publicVK := crypto.GenerateKeypair()
	tot := 0
	for i := 0; i < 1000; i++ {
		s, err := c.AppendForSize(ctx, []byte(ids[i]+strconv.Itoa(i)), publicVK)
		if err != nil {
			panic(err)
		}
		tot = tot + s
	}
	fmt.Printf("Append: %v\n", tot)
	time.Sleep(time.Second * 10)

	tot = 0

	pks := make([][]byte, 1000)
	for i := 0; i < 1000; i++ {
		var err error
		var s int
		s, err, pks[i] = c.LookUpPKForSize(ctx, []byte(ids[i]))
		//fmt.Println(i)
		if err != nil {
			panic(err)
		}
		tot = tot + s
	}
	fmt.Printf("LookupPK: %v\n", tot)
	time.Sleep(time.Second * 10)

	tot = 0
	for i := 0; i < 1000; i++ {
		s, err := c.MonitoringPKForSize(ctx, []byte(ids[i]), pks[i], 1)
		if err != nil {
			panic(err)
		}
		tot = tot + s
	}
	fmt.Printf("MonitorPK 1: %v\n", tot)
	time.Sleep(time.Second * 10)

	tot = 0
	for i := 0; i < 1000; i++ {
		s, err := c.MonitoringPKForSize(ctx, []byte(ids[i]), pks[i], 10)
		if err != nil {
			panic(err)
		}
		tot = tot + s
	}
	fmt.Printf("MonitorPK 10: %v\n", tot)
	time.Sleep(time.Second * 10)

	tot = 0
	for i := 0; i < 1000; i++ {
		s, err := c.MonitoringPKForSize(ctx, []byte(ids[i]), pks[i], 100)
		if err != nil {
			panic(err)
		}
		tot = tot + s
	}
	fmt.Printf("MonitorPK 100: %v\n", tot)
	time.Sleep(time.Second * 10)

	tot = 0
	for i := 0; i < 1000; i++ {
		s, err := c.MonitoringPKForSize(ctx, []byte(ids[i]), pks[i], 1000)
		if err != nil {
			panic(err)
		}
		tot = tot + s
	}
	fmt.Printf("MonitorPK 1000: %v\n", tot)
	time.Sleep(time.Second * 10)

}

func main() {
	numUsers, _ := strconv.Atoi(os.Args[1])
	E2EConiksClient(numUsers)
}
