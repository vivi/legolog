package bench

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/huyuncong/MerkleSquare/auditor/auditorsrv"
	coniksgrpcint "github.com/huyuncong/MerkleSquare/coniks/coniks-grpcint"
	conikssys "github.com/huyuncong/MerkleSquare/coniks/coniks-system"
	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/grpcint"
	"github.com/huyuncong/MerkleSquare/lib/storage"
	"github.com/huyuncong/MerkleSquare/merkleserver/merklesrv"
	"github.com/huyuncong/MerkleSquare/verifier/verifierd"

	"github.com/immesys/bw2/crypto"

	"google.golang.org/grpc"
)

const serverPort = constants.ServerPort
const auditorPort = constants.AuditorPort
const verifierPort = constants.VerifierPort
const serverAddr = ServerAddr
const auditorAddr = AuditorAddr
const verifierAddr = VerifierAddr
const epochDuration = 0 //constants.EpochDuration
const verifierCycleDuration = 0

func setupServer(db storage.Storage) *merklesrv.Server {
	serv := merklesrv.NewServer(db, epochDuration)
	ServerListenSocket, err := net.Listen("tcp", serverPort)
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	grpcint.RegisterMerkleSquareServer(s, serv)
	go func() {
		if err = s.Serve(ServerListenSocket); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return serv
}

// A function to populate the server with many entries.
func PopulateStorage(s *merklesrv.Server, numUsers int, batchSize int) {
	//TODO change this function to load serialized merkle tree
	//and connect to prebuilt k-v store
	ctx := context.Background()
	masterSK, masterVK := crypto.GenerateKeypair()
	signature := make([]byte, 64)
	crypto.SignBlob(masterSK, masterVK, signature, masterVK)
	for i := 0; i < numUsers/batchSize; i++ {
		var wg sync.WaitGroup
		wg.Add(batchSize / 2)
		for j := 0; j < batchSize/2; j++ {
			go func(i int, j int) {
				defer wg.Done()
				username := []byte("populateuser" + strconv.Itoa(i*batchSize/2+j))
				s.RegisterUserKey(ctx, username, masterVK, signature, false)
				s.AppendUserKey(ctx, username, masterVK, masterSK)
			}(i, j)
		}
		wg.Wait()
	}
}

func PopulateEpochs(s *merklesrv.Server, numAppends int, numEpochs int) {
	//TODO change this function to load serialized merkle tree
	//and connect to prebuilt k-v store
	ctx := context.Background()
	masterSK, masterVK := crypto.GenerateKeypair()
	signature := make([]byte, 64)
	crypto.SignBlob(masterSK, masterVK, signature, masterVK)
	numAppends /= 2
	for i := 0; i < numEpochs; i++ {
		var wg sync.WaitGroup
		wg.Add(numAppends)
		for j := 0; j < numAppends; j++ {
			go func(i int, j int) {
				defer wg.Done()
				username := []byte("epochuser" + strconv.Itoa(i*numAppends+j))
				s.RegisterUserKey(ctx, username, masterVK, signature, false)
				s.AppendUserKey(ctx, username, masterVK, masterSK)
			}(i, j)
		}
		wg.Wait()
		fmt.Printf("update %v\n", i)
		s.IncrementEpoch()
	}
}

func PopulateConiksEpochs(s *conikssys.Server, numAppends int, numEpochs int) {
	//TODO change this function to load serialized merkle tree
	//and connect to prebuilt k-v store
	ctx := context.Background()
	masterSK, masterVK := crypto.GenerateKeypair()
	signature := make([]byte, 64)
	crypto.SignBlob(masterSK, masterVK, signature, masterVK)
	for i := 0; i < numEpochs; i++ {
		var wg sync.WaitGroup
		wg.Add(numAppends)
		for j := 0; j < numAppends; j++ {
			go func(i int, j int) {
				defer wg.Done()
				username := []byte("epochuser" + strconv.Itoa(i*numAppends+j))
				s.AppendUserKey(ctx, username, masterVK)
			}(i, j)
		}
		wg.Wait()
		//fmt.Printf("update %v\n", i)
		s.IncrementEpoch()
	}
}

func SetupAuditor() *auditorsrv.Auditor {
	auditorSrv, err := auditorsrv.NewAuditor(serverAddr, epochDuration)
	if err != nil {
		panic(err)
	}

	AuditorListenSocket, err := net.Listen("tcp", auditorPort)
	if err != nil {
		panic(err)
	}

	a := grpc.NewServer()
	grpcint.RegisterAuditorServer(a, auditorSrv)

	go func() {
		if err = a.Serve(AuditorListenSocket); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return auditorSrv
}

func SetupCONIKSAuditor(addr string) *conikssys.AuditorServer {
	auditorServ, err := conikssys.NewAuditorServer(addr, epochDuration)
	if err != nil {
		panic(err)
	}

	AuditorAPIProvider := &conikssys.AuditorAPIProvider{
		AuditorServer: auditorServ,
	}

	AuditorListenSocket, err := net.Listen("tcp", auditorPort)
	if err != nil {
		panic(err)
	}

	a := grpc.NewServer()
	coniksgrpcint.RegisterAuditorServer(a, AuditorAPIProvider)

	go func() {
		if err = a.Serve(AuditorListenSocket); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return auditorServ
}

func SetupVerifier() *verifierd.Verifier {
	verifierD, err := verifierd.NewVerifier(serverAddr, auditorAddr, verifierCycleDuration)
	if err != nil {
		panic(err)
	}

	VerifierListenSocket, err := net.Listen("tcp", verifierPort)
	if err != nil {
		panic(err)
	}

	v := grpc.NewServer()
	grpcint.RegisterVerifierServer(v, verifierD)

	go func() {
		if err = v.Serve(VerifierListenSocket); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return verifierD
}
