package core

import (
	// "fmt"
	"fmt"
	"testing"
	"time"

	"math/rand"

	// legolog "github.com/huyuncong/MerkleSquare/legolog/server"
	"github.com/immesys/bw2/crypto"
)

var testCfg Config = Config{
	AggHistoryDepth: 31,
}

func TestPartitionAggHistAppend(t *testing.T) {
	partition := NewAggHistPartition(testCfg, "")
	for i := 0; i < 32; i++ {
		partition.Append([]byte{byte(i)}, []byte{byte(i)}, []byte{byte(i)}, []byte{byte(i)})
	}
}

func TestPartitionIncrementUpdate(t *testing.T) {
	partition := NewAggHistPartition(testCfg, "")
	for i := 0; i < 32; i++ {
		partition.Append([]byte{byte(i)}, []byte{byte(i)}, []byte{byte(i)}, []byte{byte(i)})
	}
	partition.IncrementUpdateEpoch()
	partition.IncrementVerificationPeriod()
}

func TestPartitionValidateProof(t *testing.T) {
	partition := NewAggHistPartition(testCfg, "")
	masterSK, masterVK := crypto.GenerateKeypair()

	for i := 0; i < 32; i++ {
		partition.Append([]byte{byte(i)}, []byte{byte(i)}, masterVK, masterVK)
	}
	partition.IncrementUpdateEpoch()
	partition.IncrementVerificationPeriod()

	_, masterVK2 := crypto.GenerateKeypair()
	signature := make([]byte, 64)
	crypto.SignBlob(masterSK, masterVK, signature, append(masterVK2, []byte("1")...))

	for i := 32; i < 64; i++ {
		partition.Append([]byte{byte(i)}, []byte{byte(i)}, masterVK2, signature)
	}
	partition.IncrementUpdateEpoch()
	//partition.IncrementVerificationPeriod()
	digest := partition.GetDigest()
	proof := partition.GenerateExistenceProof([]byte{byte(32)}, masterVK2, signature)

	var v AggHistVerifier
	ok, err := v.ValidatePKProof(digest, proof, []byte{byte(32)}, masterVK2, signature, 1, masterVK)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Proof validation failed")
	}

	ok, _ = v.ValidatePKProof(digest, proof, []byte{byte(0)}, []byte("foo"), signature, 1, masterVK)
	if ok {
		t.Error("Proof should not have validated ")
	}
}

func TestPartitionNonAggValidateProof(t *testing.T) {
	var testCfg Config = Config{
		AggHistory:      false,
		AggHistoryDepth: 31,
	}
	_ = testCfg

	// partition := NewAggHistPartition(testCfg)
	partition := NewPartition()
	masterSK, masterVK := crypto.GenerateKeypair()

	for i := 0; i < 32; i++ {
		partition.Append([]byte{byte(i)}, []byte{byte(i)}, masterVK, masterVK)
	}
	partition.IncrementUpdateEpoch()
	// partition.IncrementVerificationPeriod()

	_, masterVK2 := crypto.GenerateKeypair()
	signature := make([]byte, 64)
	crypto.SignBlob(masterSK, masterVK, signature, append(masterVK2, []byte("1")...))

	for i := 0; i < 32; i++ {
		partition.Append([]byte{byte(i)}, []byte{byte(i)}, masterVK2, signature)
	}
	partition.IncrementUpdateEpoch()
	// partition.IncrementVerificationPeriod()
	digest := partition.GetDigest()
	proof := partition.GenerateExistenceProof([]byte{byte(0)}, masterVK2, signature)

	var v AggHistVerifier
	ok, err := v.ValidatePKProof(digest, proof, []byte{byte(0)}, masterVK2, signature, 1, masterVK)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Proof validation failed")
	}

	ok, _ = v.ValidatePKProof(digest, proof, []byte{byte(0)}, []byte("foo"), signature, 1, masterVK)
	if ok {
		t.Error("Proof should not have validated ")
	}
}

func TestInsanity(t *testing.T) {
	partition := NewPartition()
	_, masterVK := crypto.GenerateKeypair()
	identifier := []byte("\x1c\"\xeeO0\x9b\x9ci\xe6\xb9\x17Q\x96\x88\"١\x1aU\xd4<@\xb3`3\xa7>k¥\xd3\xd5")
	username := identifier
	value := identifier
	signature := []byte("\x18\x9ab`J\xd4B\xb0\x98\xb5j\v\xbf\xf1vsTw.\x84\x84'1\xca\xd3\te\xeb\xcc\xc35vei\x9f7\xa4\xfc#\b\x00>Wq\x84E\x99|(&|e~\f\xf9\xcb\xfc\x02\x90G\x1a\x9d\xdc\x01")
	partition.Append(username, identifier, value, signature)

	partition.IncrementUpdateEpoch()
	// partition.IncrementVerificationPeriod()
	digest := partition.GetDigest()
	proof := partition.GenerateExistenceProof(identifier, value, signature)

	var v AggHistVerifier
	ok, err := v.ValidatePKProof(digest, proof, identifier, value, signature, 0, masterVK)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error("Proof validation failed")
	}
}

func TestBroken(t *testing.T) {
	partition := NewPartition()

	for i := 0; i < 64; i++ {
		id := make([]byte, 32)
		_, err := rand.Read(id)
		// fmt.Println("id", id)
		if err != nil {
			panic(err)
		}
		SK, VK := crypto.GenerateKeypair()

		signature := make([]byte, 64)
		ins := id
		crypto.SignBlob(SK, VK, signature,
			append(ins, []byte("0")...))
		// TODO: remove this line
		partition.Append(ins, ins, ins, signature)

		partition.IncrementUpdateEpoch()

		// LOOKUP THE STUFF //
		identifier, value, signature := ins, ins, signature
		proof := partition.GenerateExistenceProof(identifier, value, signature)
		if proof == nil {
			panic("proof is nil")
		}

		// CHECK THE STUFF //
		var v AggHistVerifier
		ok, err := v.ValidatePKProof(partition.GetDigest(), proof, identifier, value, signature, 0, VK)
		if err != nil {
			panic(err)
		}
		if !ok {
			panic("failed to validate proof")
		}
	}
}

func TestPreloading(t *testing.T) {

	partition := NewPartition()

	// preload 1 million appends to the server
	numPreloadAppends := 1000000
	start := time.Now()
	for i := 0; i < numPreloadAppends; i++ {
		id := make([]byte, 32)
		_, err := rand.Read(id)
		if err != nil {
			panic(err)
		}
		SK, VK := crypto.GenerateKeypair()
		signature := make([]byte, 64)
		ins := id
		crypto.SignBlob(SK, VK, signature,
			append(ins, []byte("0")...))
		partition.Append(ins, ins, ins, VK)
	}
	// inc verification period twice to move preloaded data to query copy
	partition.IncrementVerificationPeriod()
	partition.IncrementVerificationPeriod()
	end := time.Now()
	fmt.Printf("Preloading %d appends to 1 partition took %f seconds", numPreloadAppends, end.Sub(start).Seconds())
}
