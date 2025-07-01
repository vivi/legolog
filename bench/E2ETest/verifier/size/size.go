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
const serverAddr = bench.ServerAddr
const verifierAddr = "localhost" + constants.VerifierPort
const auditorAddr = bench.AuditorAddr
const baseUsername = "testuserforverifier"
const iterations = 1000

func E2EVerifier() {
	ctx := context.Background()
	c, _ := client.NewClient(serverAddr, "", verifierAddr)
	v := bench.SetupVerifier()

	masterSK, masterVK := crypto.GenerateKeypair()
	_, publicVK := crypto.GenerateKeypair()

	//client does four ops, somehow passes it to verifier.
	for i := 0; i < iterations; i++ {
		_, err := c.Register(ctx, []byte(baseUsername+strconv.Itoa(i)), masterSK, masterVK)
		if err != nil {
			panic(err)
		}
	}

	for i := 0; i < iterations; i++ {
		_, _, err := c.Append(ctx, []byte(baseUsername+strconv.Itoa(i)), publicVK)
		if err != nil {
			panic(err)
		}
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
		panic("could not dial auditor server")
	}

	fmt.Println("Press enter after updating server epoch.")
	input = bufio.NewScanner(os.Stdin)
	input.Scan()

	//repeat 10 times
	// for i := 0; i < 10; i++ {

	//Run functions using request and digest.
	tot := 0
	fmt.Println(len(v.RegisterRequests))
	for _, registerRequest := range v.RegisterRequests {
		auditorResponse, err := auditorClient.GetEpochUpdate(ctx)
		if err != nil {
			fmt.Println("Could not query auditor server, skipping this verify cycle")
			return
		}
		var digest = new(core.Digest)
		marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
		json.Unmarshal(marshaledDigest, &digest)
		_, _, s := v.VerifyRegisterForSize(ctx, registerRequest, digest)
		tot = tot + s
	}
	fmt.Printf("Nonexistence: %v\n", tot)
	time.Sleep(time.Second * 10)

	tot = 0
	for _, appendRequest := range v.AppendRequests {
		auditorResponse, err := auditorClient.GetEpochUpdate(ctx)
		if err != nil {
			fmt.Println("Could not query auditor server, skipping this verify cycle")
			return
		}
		var digest = new(core.Digest)
		marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
		json.Unmarshal(marshaledDigest, &digest)
		_, _, s := v.VerifyAppendForSize(ctx, appendRequest, digest)
		tot = tot + s
	}
	fmt.Printf("Existence: %v\n", tot)
	time.Sleep(time.Second * 10)

	tot = 0
	for _, lookUpRequest := range v.LookupRequests {
		if lookUpRequest.ProofRequest.GetIsMasterKey() {
			auditorResponse, err := auditorClient.GetEpochUpdate(ctx)
			if err != nil {
				fmt.Println("Could not query auditor server, skipping this verify cycle")
				return
			}
			var digest = new(core.Digest)
			marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
			json.Unmarshal(marshaledDigest, &digest)
			_, _, s := v.VerifyLookUpForSize(ctx, lookUpRequest, digest)
			tot = tot + s
		}
	}
	fmt.Printf("LookupMK: %v\n", tot)
	time.Sleep(time.Second * 10)

	tot = 0
	for _, lookUpRequest := range v.LookupRequests {
		if !lookUpRequest.ProofRequest.GetIsMasterKey() {
			auditorResponse, err := auditorClient.GetEpochUpdate(ctx)
			if err != nil {
				fmt.Println("Could not query auditor server, skipping this verify cycle")
				return
			}
			var digest = new(core.Digest)
			marshaledDigest := auditorResponse.GetCkPoint().GetMarshaledDigest()
			json.Unmarshal(marshaledDigest, &digest)
			_, _, s := v.VerifyLookUpForSize(ctx, lookUpRequest, digest)
			tot = tot + s
		}
	}
	fmt.Printf("LookupPK: %v\n", tot)
	time.Sleep(time.Second * 10)

	// }

}

func main() {
	E2EVerifier()
}
