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
const serverAddr = "34.215.95.58" + serverPort
const baseUsername = "testuser"

func E2EAuditor() {
	auditorServ := bench.SetupCONIKSAuditor(serverAddr)
	auditorServ.QueryServer()
	basecheckpoint := auditorServ.Checkpoints
	baseEpochs := auditorServ.Epochs

	fmt.Println("Press enter after updating server epoch.")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	for i := 0; i < 10; i++ {
		tot := 0
		for j := 0; j < 1000; j++ {
			s := auditorServ.QueryServerForSize()
			auditorServ.Checkpoints = basecheckpoint
			auditorServ.Epochs = baseEpochs
			tot = tot + s
		}
		fmt.Printf("Auditor: %v\n", tot)
		time.Sleep(time.Second * 10)
	}
}

func main() {
	E2EAuditor()
}
