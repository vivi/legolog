package main

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"

	"github.com/huyuncong/MerkleSquare/bench"

	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/grpcint"
	"github.com/huyuncong/MerkleSquare/merkleserver/merklesrv"
	"github.com/huyuncong/MerkleSquare/lib/storage"

	"google.golang.org/grpc"
)

// ServerPort is the port on which a server listens for
// incoming connections.
const ServerPort = constants.ServerPort
const EpochDuration = 0

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctx := context.Background()

	dir, err := ioutil.TempDir("", "teststore")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	// var err error
	// dir := "./temp"

	db := storage.OpenFile(dir)
	if err != nil {
		panic(err)
	}
	defer db.Close(ctx)

	serv := merklesrv.NewServer(db, EpochDuration)
	serv.CacheExtensionProofs = false

	listenSocket, err := net.Listen("tcp", ServerPort)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	grpcint.RegisterMerkleSquareServer(s, serv)

	go func() {
		if err = s.Serve(listenSocket); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	bench.PopulateStorage(serv, 1000000, 100)
	serv.IncrementEpoch()
	fmt.Print("Press enter after first auditor query.")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	bench.PopulateEpochs(serv, 256, 1)
	fmt.Println("Finished epoch updates.")

	// fmt.Println("Press [t] to terminate this server.")
	// for {
	// 	input := bufio.NewScanner(os.Stdin)
	// 	input.Scan()
	// 	if input.Text() == "t" {
	// 		break
	// 	}
	// }
	fmt.Println("Press enter [u] to update, [t] to terminate this server.")
	for {
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		if input.Text() == "t" {
			break
		} else if input.Text() == "u" {
			serv.IncrementEpoch()
			fmt.Printf("updated server to epoch %v.\n", serv.GetEpoch())
		}
	}
}
