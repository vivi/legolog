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
	"strconv"
	"time"

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
const EpochDuration = 0 //constants.EpochDuration

const epochDuration = time.Second

func main() {
	numLeaves, _ := strconv.Atoi(os.Args[1])
	batchSize, _ := strconv.Atoi(os.Args[2])
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

	start := time.Now()
	bench.PopulateStorage(serv, numLeaves, batchSize)

	serv.IncrementEpoch()

	listenSocket, err := net.Listen("tcp", ServerPort)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	grpcint.RegisterMerkleSquareServer(s, serv)

	fmt.Println("done initialization!")
	fmt.Println(time.Since(start))

	go func() {
		if err = s.Serve(listenSocket); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	fmt.Println("Press enter [u] to update, [t] to terminate this server, [k] to keep updating")
	for {
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		if input.Text() == "t" {
			break
		} else if input.Text() == "u" {
			serv.IncrementEpoch()
			fmt.Println("updated server.")
		} else if input.Text() == "k" {
			serv.SetEpochDuration(epochDuration)
			go serv.EpochLoop(time.Unix(0, time.Now().Add(epochDuration).UnixNano()))
		}
	}
}
