package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"

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

	configPtr := flag.String("config", "../experiments/configs/test.yaml", "config file path")
	flag.Parse()

	cfg, err := core.ParseConfig(*configPtr)
	if err != nil {
		panic(errors.New("Failed to load config: " + err.Error()))
	}

	tmpdir, err := os.MkdirTemp("", "legologtest")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	serv := legolog.NewStoppedServer(storage.NewMapStorage(), &cfg, tmpdir)
	fmt.Println("preloading server")
	serv.PreloadServer(0, 1e6, 32, 32)
	serv.IncrementUpdateEpoch()
	serv.IncrementVerificationPeriod()
	serv.IncrementVerificationPeriod()

	/*
		f, err := os.Create("cpu-profile.pb.gz")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	*/

	/*
		startIdx := int(1e6)
		for i := 0; i < 2; i++ {
			start := time.Now()
			serv.PreloadServer(startIdx, 10000, 32, 32)
			serv.IncrementUpdateEpoch()
			serv.IncrementVerificationPeriod()
			end := time.Now()
			fmt.Println("preloaded", i, "/100", "in", end.Sub(start).Seconds(), "seconds")
				if i == 49 {
					f, err = os.Create("heap-profile.pb.gz")
					if err != nil {
						panic(err)
					}
					defer f.Close()
					//runtime.GC()
					if err := pprof.WriteHeapProfile(f); err != nil {
						log.Fatal("could not write memory profile: ", err)
					}
				}
			startIdx += 10000
		}
	*/

	fmt.Println("done preloading server")

	serv.Start()

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
