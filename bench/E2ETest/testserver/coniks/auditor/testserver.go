package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"runtime"

	coniksgrpcint "github.com/huyuncong/MerkleSquare/coniks/coniks-grpcint"
	conikssys "github.com/huyuncong/MerkleSquare/coniks/coniks-system"
	"github.com/huyuncong/MerkleSquare/constants"
	"google.golang.org/grpc"
)

// Auditor is the port on which a auditor listens for
// incoming connections.
const ServerPort = constants.ServerPort
const AuditorPort = constants.AuditorPort
const serverAddr = "localhost" + ServerPort
const queryDuration = 0

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	serv, err := conikssys.NewAuditorServer(serverAddr, queryDuration)
	if err != nil {
		panic(err)
	}

	APIProvider := &conikssys.AuditorAPIProvider{
		AuditorServer: serv,
	}

	listenSocket, err := net.Listen("tcp", AuditorPort)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	coniksgrpcint.RegisterAuditorServer(s, APIProvider)

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
			serv.QueryServer()
			fmt.Print("queried server.")
		}
	}
}
