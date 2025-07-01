package main

import (
	"MerkleSquare/core"
	"errors"
	"flag"
	"log"
	"net"
	"runtime"

	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/legolog/auditor/auditorsrv"
	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"
	"google.golang.org/grpc"
)

// AuditorPort is the port on which a auditor listens for
// incoming connections.
const ServerPort = constants.ServerPort
const AuditorPort = constants.AuditorPort

func main() {
	experimentConfigPtr := flag.String("exp_config", "../experiments/exp_configs/test.yaml", "experiment config file path")
	configPtr := flag.String("config", "../experiments/configs/test.yaml", "config file path")
	flag.Parse()
	expCfg, err := core.ParseExperimentConfig(*experimentConfigPtr)
	if err != nil {
		panic(errors.New("Failed to load experiment config: " + err.Error()))
	}
	if *configPtr == "" {
		panic("config file not specified")
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	serv, err := auditorsrv.NewAuditor(expCfg.ServerAddr+ServerPort, *configPtr)
	if err != nil {
		panic(err)
	}

	listenSocket, err := net.Listen("tcp", AuditorPort)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	legolog_grpcint.RegisterAuditorServer(s, serv)

	if err = s.Serve(listenSocket); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
