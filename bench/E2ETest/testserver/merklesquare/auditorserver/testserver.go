package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"

	"github.com/huyuncong/MerkleSquare/auditor/auditorsrv"
	"github.com/huyuncong/MerkleSquare/bench"
	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/grpcint"
	"google.golang.org/grpc"
)

// Auditor is the port on which a auditor listens for
// incoming connections.
const ServerPort = constants.ServerPort
const AuditorPort = constants.AuditorPort
const serverAddr = bench.ServerAddr

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	auditor, err := auditorsrv.NewAuditor(serverAddr, 0)
	if err != nil {
		panic(err)
	}

	listenSocket, err := net.Listen("tcp", AuditorPort)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	grpcint.RegisterAuditorServer(s, auditor)

	go func() {
		if err = s.Serve(listenSocket); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	fmt.Println("Press enter [q] to query, [t] to terminate this server.")
	for {
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		if input.Text() == "t" {
			break
		} else if input.Text() == "q" {
			auditor.QueryServer()
			fmt.Print("queried server.")
		}
	}
}
