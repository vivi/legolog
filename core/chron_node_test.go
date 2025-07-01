package core

import (
	"fmt"
	"testing"

	crypto "github.com/huyuncong/MerkleSquare/lib/crypto"
)

func TestChronTreeAppendLots(t *testing.T) {
	tree := NewChronTree()
	key := makePrefixFromKey([]byte{0b01})
	valueHash := crypto.Hash([]byte{0b1})
	sigHash := crypto.Hash([]byte{0b1})
	for i := 0; i < 200; i++ {
		t.Log(i)
		tree.Append(key, valueHash, sigHash)
	}
}

func TestChronTreeBasicGenerateConsistencyProof(t *testing.T) {
	tree := NewChronTree()
	key := makePrefixFromKey([]byte{0b01})
	valueHash := crypto.Hash([]byte{0b1})
	sigHash := crypto.Hash([]byte{0b1})
	for i := 0; i < 1; i++ {
		tree.Append(key, valueHash, sigHash)
	}
	proof := tree.GenerateConsistencyProof(0, 1)
	t.Log(proof)
}

func TestChronTreeGenerateConsistencyProof(t *testing.T) {
	tree := NewChronTree()
	key := makePrefixFromKey([]byte{0b01})
	valueHash := crypto.Hash([]byte{0b1})
	sigHash := crypto.Hash([]byte{0b1})
	for i := 0; i < 5; i++ {
		tree.Append(key, valueHash, sigHash)
	}
	proof := tree.GenerateConsistencyProof(4, 5)
	t.Log(proof)
}

func TestVerifyConsistencyProof(t *testing.T) {
	testVerifyConsistencyProof(0, 1, t)
	testVerifyConsistencyProof(4, 5, t)
	testVerifyConsistencyProof(10, 11, t)
	testVerifyConsistencyProof(19, 21, t)
	testVerifyConsistencyProof(32, 64, t)

}

func testVerifyConsistencyProof(oldSize int, newSize int, t *testing.T) {
	tree := NewChronTree()
	key := makePrefixFromKey([]byte{0b01})
	valueHash := crypto.Hash([]byte{0b1})
	sigHash := crypto.Hash([]byte{0b1})
	for i := 0; i < oldSize; i++ {
		tree.Append(key, valueHash, sigHash)
		tree.PrintTree()
	}
	oldDigest := Digest{
		Roots: [][]byte{tree.GetRootHash()},
		Size:  uint32(oldSize),
	}
	for i := oldSize; i < newSize; i++ {
		tree.Append(key, valueHash, sigHash)
		tree.PrintTree()
	}
	newDigest := Digest{
		Roots: [][]byte{tree.root.getHash()},
		Size:  uint32(newSize),
	}
	proof := tree.GenerateConsistencyProof(uint32(oldSize), uint32(newSize))
	for _, p := range proof.PrefixHashes {
		fmt.Printf("%x\n", p[:4])
	}
	ok := VerifyConsistencyProof(&oldDigest, &newDigest, proof)
	if !ok {
		t.Error("VerifyConsistencyProof failed")
	}
}
