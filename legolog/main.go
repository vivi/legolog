package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/core"
	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"
	legolog "github.com/huyuncong/MerkleSquare/legolog/server"

	"github.com/huyuncong/MerkleSquare/lib/storage"
	"google.golang.org/grpc"
)

// ServerPort is the port on which a server listens for
// incoming connections.
const ServerPort = constants.ServerPort
const EpochDuration = constants.EpochDuration
const TestingEpochDuration = time.Second * 10

// const DefaultNumPartitions = 4
const TestingNumPartitions = 1
const TestingAggHistory = false

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

	configPtr := flag.String("config", "", "config file path")
	flag.Parse()
	updateEpochDuration := TestingEpochDuration
	numPartitions := TestingNumPartitions
	aggHistory := TestingAggHistory
	var config core.Config
	if *configPtr == "" {
		log.Println("config file not specified, using default values")
	} else {
		config, err = core.ParseConfig(*configPtr)
		if err != nil {
			panic(err)
		}
		updateEpochDuration = config.UpdatePeriod
		numPartitions = int(config.Partitions)
		aggHistory = config.AggHistory
	}

	fmt.Println("aggHistory: ", aggHistory)
	tmpdir, err := os.MkdirTemp("", "legolog")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)
	serv := legolog.NewServer(storage.NewMapStorage(), updateEpochDuration, numPartitions, aggHistory, &config, tmpdir)

	listenSocket, err := net.Listen("tcp", ServerPort)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	legolog_grpcint.RegisterLegoLogServer(s, serv)

	if err = s.Serve(listenSocket); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
