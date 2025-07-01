package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/huyuncong/MerkleSquare/auditor/auditorclt"
	"github.com/huyuncong/MerkleSquare/bench"
	"github.com/huyuncong/MerkleSquare/client"
	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/core"

	"github.com/immesys/bw2/crypto"
)

const serverPort = constants.ServerPort
const serverAddr = "172.31.0.252" + constants.ServerPort
const verifierAddr = "localhost" + constants.VerifierPort
const auditorAddr = "172.31.5.175" + constants.AuditorPort
const baseUsername = "testuser"
const iterations = 1000

func E2EVerifier() {
	ctx := context.Background()
	c, _ := client.NewClient(serverAddr, "", verifierAddr)
	verifier := bench.SetupVerifier()

	masterSK, masterVK := crypto.GenerateKeypair()
	_, publicVK := crypto.GenerateKeypair()

	//client does four ops, somehow passes it to verifier.
	for i := 0; i < iterations; i++ {
		c.Register(ctx, []byte(baseUsername+strconv.Itoa(i)), masterSK, masterVK)
	}

	for i := 0; i < iterations; i++ {
		c.Append(ctx, []byte(baseUsername+strconv.Itoa(i)), publicVK)
	}

	fmt.Println("Press enter after updating server epoch.")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	for i := 0; i < iterations; i++ {
		c.LookUpMK(ctx, []byte(baseUsername+strconv.Itoa(i)))
	}

	for i := 0; i < iterations; i++ {
		c.LookUpPK(ctx, []byte(baseUsername+strconv.Itoa(i)))
	}

	auditorClient, err := auditorclt.NewAuditorClient(auditorAddr)
	if err != nil {
		panic("could not create auditor client")
	}

	fmt.Println("Press enter after updating server epoch.")
	input = bufio.NewScanner(os.Stdin)
	input.Scan()

	//Run functions using request and digest.
	start := time.Now()
	for _, registerRequest := range verifier.RegisterRequests {
		auditorResponse, err := auditorClient.GetEpochUpdate(ctx)
		if err != nil {
			fmt.Println("Could not query auditor server, skipping this verify cycle")
			return
		}
		var digest = new(core.Digest)
		marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
		json.Unmarshal(marshaledDigest, &digest)
		verifier.VerifyRegister(ctx, registerRequest, digest)
	}
	fmt.Printf("Nonexistence: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

	start = time.Now()
	for _, appendRequest := range verifier.AppendRequests {
		auditorResponse, err := auditorClient.GetEpochUpdate(ctx)
		if err != nil {
			fmt.Println("Could not query auditor server, skipping this verify cycle")
			return
		}
		var digest = new(core.Digest)
		marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
		json.Unmarshal(marshaledDigest, &digest)
		verifier.VerifyAppend(ctx, appendRequest, digest)
	}
	fmt.Printf("Existence: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

	start = time.Now()
	for _, lookUpRequest := range verifier.LookupRequests {
		if lookUpRequest.ProofRequest.GetIsMasterKey() {
			auditorResponse, err := auditorClient.GetEpochUpdate(ctx)
			if err != nil {
				fmt.Println("Could not query auditor server, skipping this verify cycle")
				return
			}
			var digest = new(core.Digest)
			marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
			json.Unmarshal(marshaledDigest, &digest)
			verifier.VerifyLookUp(ctx, lookUpRequest, digest)
		}
	}
	fmt.Printf("LookupMK: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

	start = time.Now()
	for _, lookUpRequest := range verifier.LookupRequests {
		if !lookUpRequest.ProofRequest.GetIsMasterKey() {
			auditorResponse, err := auditorClient.GetEpochUpdate(ctx)
			if err != nil {
				fmt.Println("Could not query auditor server, skipping this verify cycle")
				return
			}
			var digest = new(core.Digest)
			marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
			json.Unmarshal(marshaledDigest, &digest)
			verifier.VerifyLookUp(ctx, lookUpRequest, digest)
		}
	}
	fmt.Printf("LookupPK: %v\n", time.Since(start))
	time.Sleep(time.Second * 10)

	// }

}

func main() {
	E2EVerifier()
}
