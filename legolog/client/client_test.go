package legolog

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"errors"
	"testing"

	"github.com/huyuncong/MerkleSquare/constants"
	"github.com/huyuncong/MerkleSquare/core"
	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"

	"github.com/immesys/bw2/crypto"
)

const ServerAddr = "localhost" + constants.ServerPort

// const AuditorAddr = "localhost" + constants.AuditorPort
// const VerifierAddr = "localhost" + constants.VerifierPort

func TestClient(t *testing.T) {
	ctx := context.Background()

	c, err := NewClient(ServerAddr, "", "")
	if err != nil {
		t.Error(errors.New("Failed to start client: " + err.Error()))
	}

	aliceUsername := []byte("alice")
	aliceIdentifier1 :=
		[]byte("alice_key")
	masterSK, masterVK := crypto.GenerateKeypair()
	_, aliceVK1 := crypto.GenerateKeypair()
	_, aliceVK3 := crypto.GenerateKeypair()
	_, aliceVK4 := crypto.GenerateKeypair()

	/*
		Register Alice.
	*/
	pos, err := c.Register(ctx, aliceUsername, masterSK, masterVK)
	if err != nil {
		t.Error(errors.New("Failed to register first user: " + err.Error()))
	}
	if pos != 0 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	/*
		Start adding several identifier-value pairs for Alice.
	*/
	aliceId1Pos, _, err := c.Append(ctx, aliceUsername, aliceIdentifier1, aliceVK1)
	if err != nil {
		t.Error(errors.New("Failed to append alice's PK: " + err.Error()))
	}
	if aliceId1Pos != 1 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	// extra appends below
	pos, _, err = c.Append(ctx, aliceUsername, []byte("alice_key1"), aliceVK1)
	if err != nil {
		t.Error(errors.New("Failed to append alice's PK: " + err.Error()))
	}
	if pos != 2 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	pos, _, err = c.Append(ctx, aliceUsername, []byte("alice_key2"), aliceVK1)
	if err != nil {
		t.Error(errors.New("Failed to append alice's PK: " + err.Error()))
	}
	if pos != 3 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	aliceIdentifier3 := []byte("alice_key3")
	aliceId3Pos, _, err := c.Append(ctx, aliceUsername, aliceIdentifier3, aliceVK3)
	if err != nil {
		t.Error(errors.New("Failed to append alice's PK: " + err.Error()))
	}
	if aliceId3Pos != 4 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}
	// end extra appends

	time.Sleep(time.Second * 2)

	/*
		Look up Alice's master key and PKs.
	*/
	key, pos, err := c.LookUpMK(ctx, aliceUsername)
	if err != nil {
		t.Error(errors.New("Failed to look up alice's MK: " + err.Error()))
	}
	if pos != 0 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}
	if !reflect.DeepEqual(key, masterVK) {
		t.Error(errors.New("Master key mismatch"))
	}

	key, pos, err = c.LookUpPK(ctx, aliceIdentifier1)
	if err != nil {
		t.Error(errors.New("Failed to look up alice's PK: " + err.Error()))
	}
	if !reflect.DeepEqual(key, aliceVK1) {
		t.Error(errors.New("Public key mismatch"))
	}

	if aliceId1Pos != 1 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	key, pos, err = c.LookUpPK(ctx, aliceIdentifier3)
	if err != nil {
		t.Error(errors.New("Failed to look up alice's PK: " + err.Error()))
	}
	if !reflect.DeepEqual(key, aliceVK3) {
		t.Error(errors.New("Public key mismatch"))
	}

	if pos != aliceId3Pos {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	time.Sleep(time.Second * 10)

	/*
		First PK Verify, for a key that was added in the first update epoch.
	*/

	pk, pos, signature, marshalledProof, err := c.LookUpPKVerify(ctx, aliceUsername, aliceIdentifier1)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("pk, pos: ", pk, pos)

	var proof core.LegologExistenceProof
	json.Unmarshal(marshalledProof, &proof)

	response, _ := c.legologClient.GetNewCheckPoint(ctx, &legolog_grpcint.GetNewCheckPointRequest{PartitionIndex: 0})
	var digest core.LegologDigest
	json.Unmarshal(response.Checkpoint.MarshaledDigest, &digest)
	valid, err := core.ValidatePKProof(&digest, &proof, aliceIdentifier1, aliceVK1, signature, pos, masterVK)
	if !valid {
		t.Error("Unable to validate PK proof: ", err.Error())
	}

	/*
		Lookup MK Verify
	*/

	mkValue, mkPos, mkSignature, mkMarshalledProof, err := c.LookUpMKVerify(ctx, aliceUsername)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("mk, pos: ", mkValue, mkPos)
	json.Unmarshal(mkMarshalledProof, &proof)

	response, _ = c.legologClient.GetNewCheckPoint(ctx, &legolog_grpcint.GetNewCheckPointRequest{PartitionIndex: 0})
	json.Unmarshal(response.Checkpoint.MarshaledDigest, &digest)
	valid, err = core.ValidateMKProof(&digest, &proof, aliceUsername, masterVK, mkSignature, pos, masterVK)
	if !valid {
		t.Error("Unable to validate MK proof: ", err.Error())
	}

	/*
		Late append, after 10 second sleep
	*/

	aliceIdentifier4 := []byte("alice_key4")
	aliceId4Pos, _, err := c.Append(ctx, aliceUsername, aliceIdentifier4, aliceVK4)
	if err != nil {
		t.Error(errors.New("Failed to append alice's PK: " + err.Error()))
	}
	if aliceId4Pos != 5 {
		t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	}

	time.Sleep(time.Second * 10)

	/*
		TODO: uncomment this and fix it! Non-existence is currently broken
		// Second PK Verify, for a key that was added later (probably the second update epoch)
		pk, pos, signature, marshalledProof, err = c.LookUpPKVerify(ctx, aliceUsername, aliceIdentifier4)
		if err != nil {
			t.Error(err)
		}

		fmt.Println("pk, pos: ", pk, pos)

		json.Unmarshal(marshalledProof, &proof)

		response, _ = c.legologClient.GetNewCheckPoint(ctx, &legolog_grpcint.GetNewCheckPointRequest{PartitionIndex: 0})
		json.Unmarshal(response.Checkpoint.MarshaledDigest, &digest)
		valid, err = core.ValidatePKProof(&digest, &proof, aliceIdentifier4, aliceVK4, signature, pos, masterVK)
		if !valid {
			t.Error("Unable to validate PK proof: ", err.Error())
		}

	*/

	/***
		Below code comes from Yuncong's repo and is not yet updated to work with legolog.
	***/

	// mk, pos, signature, marshalledProof, _ := c.LookUpMKVerify(ctx, aliceUsername)
	// fmt.Println("mk, pos: ", mk, pos)

	// json.Unmarshal(marshalledProof, &proof)
	// valid, err = core.ValidateProof(&digest, &proof, aliceIdentifier1, aliceVK1, signature, pos, masterVK)
	// if !valid {
	// 	t.Error("Unable to validate PK proof: ", err.Error())
	// }

	// isValid := core.ValidateProof(proof, aliceIdentifier1, aliceVK1)
	// _ = isValid

	// ignoring position for lookupPK
	// if pos != 1 {
	// 	t.Error(errors.New("Unexpected position: " + fmt.Sprint(pos)))
	// }

	// bobSK, bobVK := crypto.GenerateKeypair()
	// _, err = c.Register(ctx, []byte("bob"), bobSK, bobVK)
	// if err != nil {
	// 	t.Error(errors.New("Failed to register second user: " + err.Error()))
	// }
}
