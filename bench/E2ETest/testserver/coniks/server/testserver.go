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
	coniksgrpcint "github.com/huyuncong/MerkleSquare/coniks/coniks-grpcint"
	conikssys "github.com/huyuncong/MerkleSquare/coniks/coniks-system"
	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/lib/storage"

	"google.golang.org/grpc"
)

// ServerPort is the port on which a server listens for
// incoming connections.
const ServerPort = constants.ServerPort
const EpochDuration = 0

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

	serv := conikssys.NewServer(db, EpochDuration, 5000000)

	s := grpc.NewServer()
	start := time.Now()
	numEpochs := numLeaves / batchSize
	bench.PopulateConiksEpochs(serv, batchSize, numEpochs)

	APIProvider := &conikssys.ServerAPIProvider{
		Server: serv,
	}

	listenSocket, err := net.Listen("tcp", ServerPort)
	if err != nil {
		panic(err)
	}

	coniksgrpcint.RegisterCONIKSServer(s, APIProvider)

	fmt.Printf("done initialization! current epoch %v\n", serv.ConiksPAD.LatestSTR().Epoch)
	fmt.Println(time.Since(start))

	go func() {
		if err = s.Serve(listenSocket); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	fmt.Println("Press enter [u] to update, [t] to terminate this server.")
	for {
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		if input.Text() == "t" {
			break
		} else if input.Text() == "u" {
			serv.IncrementEpoch()
			fmt.Printf("updated server to epoch %v.\n", serv.ConiksPAD.LatestSTR().Epoch)
		}
	}
}
