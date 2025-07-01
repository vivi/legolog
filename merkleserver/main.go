package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"

	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/grpcint"
	"github.com/huyuncong/MerkleSquare/merkleserver/merklesrv"
	"github.com/huyuncong/MerkleSquare/lib/storage"
	"google.golang.org/grpc"
)

// ServerPort is the port on which a server listens for
// incoming connections.
const ServerPort = constants.ServerPort
const EpochDuration = constants.EpochDuration

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

	serv := merklesrv.NewServer(storage.NewMapStorage(), EpochDuration)

	listenSocket, err := net.Listen("tcp", ServerPort)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	grpcint.RegisterMerkleSquareServer(s, serv)

	if err = s.Serve(listenSocket); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
