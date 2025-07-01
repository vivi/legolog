package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/huyuncong/MerkleSquare/bench"
	"github.com/huyuncong/MerkleSquare/constants"
)

const serverPort = constants.ServerPort
const serverAddr = bench.ServerAddr
const auditorAddr = bench.AuditorAddr
const baseUsername = "testuser"

func E2EAuditor() {
	auditorServ := bench.SetupAuditor()
	auditorServ.QueryServer()
	basecheckpoint := auditorServ.Checkpoint

	fmt.Println("Press enter after updating server epoch.")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	for i := 0; i < 10; i++ {
		start := time.Now()
		for j := 0; j < 1000; j++ {
			auditorServ.QueryServer()
			auditorServ.Checkpoint = basecheckpoint
		}
		fmt.Printf("Auditor: %v\n", time.Since(start))
		time.Sleep(time.Second * 10)
	}
}

func main() {
	E2EAuditor()
}
