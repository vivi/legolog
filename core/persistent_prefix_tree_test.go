package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	math_rand "math/rand"

	crypto "github.com/huyuncong/MerkleSquare/lib/crypto"
)

func TestPrintTree(t *testing.T) {
	fmt.Println("Starting TestPrintTree")
	tree := NewPersistentPrefixTree()
	prefix := []byte{0b1, 0b1, 0b1}
	valueHash := crypto.Hash([]byte{0b1})
	tree.Insert(prefix, valueHash, 0)
	//fmt.Printf("%v\n", tree)
	tree.NextEpoch()
	prefix = []byte{0b1, 0b0}
	tree.Insert(prefix, valueHash, 1)
	tree.NextEpoch()
	prefix = []byte{0b1, 0b1, 0b0}
	tree.Insert(prefix, valueHash, 2)

	res, err := tree.LookupPath([]byte{0b1, 0b1, 0b0}, 2)
	fmt.Println(res, err)
	res, err = tree.LookupPath([]byte{0b1, 0b1, 0b0}, 1)
	fmt.Println(err)
	fmt.Printf("%s\n", tree.StringAtEpoch(0))
	fmt.Printf("%s\n", tree.StringAtEpoch(1))
	fmt.Printf("%s\n", tree.StringAtEpoch(2))
}

func TestSimpleMembershipProof(t *testing.T) {
	tree := NewPersistentPrefixTree()
	prefix := []byte{0b0}
	valueHash := crypto.Hash([]byte{0b0})
	tree.Insert(prefix, valueHash, 0)
	tree.NextEpoch()

	proof, values := tree.generateMembershipProof([]byte{0b0}, tree.currEpoch)
	rootHash := computeRootHashMembership([]byte{0b0}, proof, values)
	if !bytes.Equal(rootHash, tree.currRoot.hash) {
		t.Error("Membership proof failed")
	}
}

func TestMembershipProof(t *testing.T) {
	tree := NewPersistentPrefixTree()
	prefix := []byte{0b1, 0b1, 0b1}
	valueHash := crypto.Hash([]byte{0b1})
	tree.Insert(prefix, valueHash, 0)
	tree.NextEpoch()
	prefix = []byte{0b1, 0b0}
	tree.Insert(prefix, valueHash, 1)
	tree.NextEpoch()
	prefix = []byte{0b1, 0b1, 0b0}
	tree.Insert(prefix, valueHash, 2)

	proof, values := tree.generateMembershipProof([]byte{0b1, 0b1, 0b0}, tree.currEpoch)
	rootHash := computeRootHashMembership([]byte{0b1, 0b1, 0b0}, proof, values)
	if !bytes.Equal(rootHash, tree.currRoot.hash) {
		t.Error("Membership proof failed")
	}
}

func TestSimpleNonMbershipProof(t *testing.T) {
	tree := NewPersistentPrefixTree()
	prefix := []byte{0b1, 0b1, 0b1}
	valueHash := crypto.Hash([]byte{0b1})
	tree.Insert(prefix, valueHash, 0)
	tree.NextEpoch()

	proof := tree.generateNonMembershipProof([]byte{0b1, 0b0, 0b0}, tree.currEpoch)
	rootHash := computeRootHashNonMembership([]byte{0b1, 0b0, 0b0}, proof)
	if !bytes.Equal(rootHash, tree.currRoot.hash) {
		t.Error("NonMembership proof failed")
	}
}

func TestNonMembershipProof(t *testing.T) {
	tree := NewPersistentPrefixTree()
	prefix := []byte{0b1, 0b1, 0b1}
	valueHash := crypto.Hash([]byte{0b1})
	tree.Insert(prefix, valueHash, 0)
	tree.NextEpoch()
	prefix = []byte{0b1, 0b0, 1}
	tree.Insert(prefix, valueHash, 1)
	tree.NextEpoch()
	prefix = []byte{0b1, 0b1, 0b0}
	tree.Insert(prefix, valueHash, 2)
	tree.NextEpoch()

	proof := tree.generateNonMembershipProof([]byte{0b1, 0b0, 0b0}, tree.currEpoch)
	rootHash := computeRootHashNonMembership([]byte{0b1, 0b0, 0b0}, proof)
	if !bytes.Equal(rootHash, tree.currRoot.hash) {
		t.Error("NonMembership proof failed")
	}
}

func TestNonMembershipProofInPast(t *testing.T) {
	tree := NewPersistentPrefixTree()
	prefix := []byte{0b1, 0b1, 0b1}
	valueHash := crypto.Hash([]byte{0b1})
	tree.Insert(prefix, valueHash, 0)
	tree.NextEpoch()
	prefix = []byte{0b1, 0b0}
	tree.Insert(prefix, valueHash, 1)
	tree.NextEpoch()
	prefix = []byte{0b1, 0b1, 0b0}
	tree.Insert(prefix, valueHash, 2)

	proof := tree.generateNonMembershipProof([]byte{0b1, 0b1, 0b0}, tree.currEpoch-1)
	rootHash := computeRootHashNonMembership([]byte{0b1, 0b1, 0b0}, proof)
	if !bytes.Equal(rootHash, tree.getRootAtEpoch(tree.currEpoch-1).hash) {
		// TODO: check that this is how we check proofs against roots in the past
		t.Error("NonMembership proof failed")
	}
}

func TestBigMembershipProof(t *testing.T) {
	tree := NewPersistentPrefixTree()
	prefix := []byte{0b1, 0b1, 0b1}
	valueHash := crypto.Hash([]byte{0b1})
	tree.Insert(prefix, valueHash, 0)
	tree.NextEpoch()

	proof := tree.generateNonMembershipProof([]byte{0b1, 0b0, 0b0}, tree.currEpoch)
	rootHash := computeRootHashNonMembership([]byte{0b1, 0b0, 0b0}, proof)
	if !bytes.Equal(rootHash, tree.currRoot.hash) {
		t.Error("NonMembership proof failed")
	}
}

func TestNonExistenceEmptyTree(t *testing.T) {
	tree := NewPersistentPrefixTree()
	tree.NextEpoch()
	tree.NextEpoch()
	tree.Insert(makePrefixFromKey([]byte{0b1}), nil, 0)
	tree.NextEpoch()
	tree.Insert(makePrefixFromKey([]byte{0b1, 0b0}), nil, 0)
	tree.NextEpoch()
	tree.Insert(makePrefixFromKey([]byte{0b1, 0b1}), nil, 0)
	proof := tree.generateNonMembershipProof(makePrefixFromKey([]byte{0b1, 0b1}), 2)
	success := validateNonMembershipProof(proof, []byte{0b1, 0b1}, tree.getHash(2))
	if !success {
		t.Error("Non-membership proof failed")
	}
}

func TestNonMembershipGoThrough(t *testing.T) {
	tree := NewPersistentPrefixTree()
	tree.NextEpoch()
	tree.NextEpoch()
	tree.Insert(makePrefixFromKey([]byte{0x9, 0x31, 0xa, 0x70, 0xf0, 0x84, 0xdb, 0x4d, 0x7d, 0xb7, 0xae, 0xf7, 0xc3, 0xab, 0xe9, 0x7d, 0xb8, 0x76, 0x80, 0xa3, 0x8f, 0x40, 0xcb, 0x1d, 0x51, 0x61, 0x81, 0x96, 0xb2, 0xc9, 0x1b, 0x4e}), nil, 0)
	tree.NextEpoch()
	tree.Insert(makePrefixFromKey([]byte{0x9e, 0x83, 0xe, 0x78, 0x11, 0x31, 0xe0, 0x67, 0xb7, 0x99, 0x25, 0x7d, 0x9e, 0x22, 0x9e, 0x8f, 0xf8, 0x4c, 0xcf, 0x19, 0x94, 0x9f, 0x68, 0x86, 0x78, 0x6b, 0x7b, 0x2d, 0x15, 0xf5, 0x32, 0x87}), nil, 0)
	tree.NextEpoch()
	tree.Insert(makePrefixFromKey([]byte{0xa9, 0x2, 0x17, 0x27, 0x86, 0x1, 0xe7, 0xe, 0xb4, 0xcb, 0x32, 0x80, 0xa0, 0xe9, 0x27, 0x8f, 0xb8, 0x7, 0xcb, 0xe6, 0x56, 0x44, 0x39, 0x43, 0x56, 0x2d, 0x8b, 0x3, 0xf8, 0xc7, 0x3b, 0x26}), nil, 0)
	proof := tree.generateNonMembershipProof(makePrefixFromKey([]byte{0xa9, 0x2, 0x17, 0x27, 0x86, 0x1, 0xe7, 0xe, 0xb4, 0xcb, 0x32, 0x80, 0xa0, 0xe9, 0x27, 0x8f, 0xb8, 0x7, 0xcb, 0xe6, 0x56, 0x44, 0x39, 0x43, 0x56, 0x2d, 0x8b, 0x3, 0xf8, 0xc7, 0x3b, 0x26}), 2)
	success := validateNonMembershipProof(proof, []byte{0xa9, 0x2, 0x17, 0x27, 0x86, 0x1, 0xe7, 0xe, 0xb4, 0xcb, 0x32, 0x80, 0xa0, 0xe9, 0x27, 0x8f, 0xb8, 0x7, 0xcb, 0xe6, 0x56, 0x44, 0x39, 0x43, 0x56, 0x2d, 0x8b, 0x3, 0xf8, 0xc7, 0x3b, 0x26}, tree.getHash(2))
	if !success {
		t.Error("Non-membership proof failed")
	}
}

func TestSimpleRootHashComputation(t *testing.T) {
	prefixTree := NewPrefixTree()

	tree := NewPersistentPrefixTree()

	prefixTree.PrefixAppend(makePrefixFromKey([]byte{0x9, 0x31, 0xa, 0x70, 0xf0, 0x84, 0xdb, 0x4d, 0x7d, 0xb7, 0xae, 0xf7, 0xc3, 0xab, 0xe9, 0x7d, 0xb8, 0x76, 0x80, 0xa3, 0x8f, 0x40, 0xcb, 0x1d, 0x51, 0x61, 0x81, 0x96, 0xb2, 0xc9, 0x1b, 0x4e}), nil, 0)
	tree.Insert(makePrefixFromKey([]byte{0x9, 0x31, 0xa, 0x70, 0xf0, 0x84, 0xdb, 0x4d, 0x7d, 0xb7, 0xae, 0xf7, 0xc3, 0xab, 0xe9, 0x7d, 0xb8, 0x76, 0x80, 0xa3, 0x8f, 0x40, 0xcb, 0x1d, 0x51, 0x61, 0x81, 0x96, 0xb2, 0xc9, 0x1b, 0x4e}), nil, 0)

	prefixTree.PrefixAppend(makePrefixFromKey([]byte{0x9e, 0x83, 0xe, 0x78, 0x11, 0x31, 0xe0, 0x67, 0xb7, 0x99, 0x25, 0x7d, 0x9e, 0x22, 0x9e, 0x8f, 0xf8, 0x4c, 0xcf, 0x19, 0x94, 0x9f, 0x68, 0x86, 0x78, 0x6b, 0x7b, 0x2d, 0x15, 0xf5, 0x32, 0x87}), nil, 1)
	tree.Insert(makePrefixFromKey([]byte{0x9e, 0x83, 0xe, 0x78, 0x11, 0x31, 0xe0, 0x67, 0xb7, 0x99, 0x25, 0x7d, 0x9e, 0x22, 0x9e, 0x8f, 0xf8, 0x4c, 0xcf, 0x19, 0x94, 0x9f, 0x68, 0x86, 0x78, 0x6b, 0x7b, 0x2d, 0x15, 0xf5, 0x32, 0x87}), nil, 1)

	fmt.Printf("prefix tree: %x\n", prefixTree.getHash())
	fmt.Printf("persistent prefix tree: %x\n", tree.getHash(tree.currEpoch))
}

func TestRandomRootHashComputation(t *testing.T) {
	math_rand.Seed(1)

	prefixTree := NewPrefixTree()
	tree := NewPersistentPrefixTree()

	for i := 0; i < 17; i++ {
		id := make([]byte, 32)
		binary.LittleEndian.PutUint32(id, uint32(math_rand.Intn(100)))

		prefixTree.PrefixAppend(makePrefixFromKey(id), nil, uint32(i))
		tree.Insert(makePrefixFromKey(id), nil, uint32(i))
		// tree.NextEpoch()
		fmt.Println("=========================")
		fmt.Println(i)
		fmt.Println(prefixTree)
		fmt.Println(tree.StringAtEpoch(tree.currEpoch))
		fmt.Println("=========================")
	}

	id := make([]byte, 32)
	binary.LittleEndian.PutUint32(id, uint32(math_rand.Intn(100)))

	prefixTree.PrefixAppend(makePrefixFromKey(id), nil, uint32(17))
	tree.Insert(makePrefixFromKey(id), nil, uint32(17))

	fmt.Printf("prefix tree: %x\n", prefixTree.getHash())
	fmt.Printf("persistent prefix tree: %x\n", tree.getHash(tree.currEpoch))
	fmt.Println(prefixTree)
	fmt.Println(tree.StringAtEpoch(tree.currEpoch))
}

func TestLookupMultipleEpochs(t *testing.T) {
	pt := NewPersistentPrefixTree()
	pt.InsertAndPrefixify([]byte{0x5b, 0x4, 0xf8, 0xe, 0x50, 0xfd, 0x1, 0xb1, 0xf, 0xdb}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x22, 0xdc, 0x53, 0x44, 0x49, 0x84, 0x92, 0xc2, 0x1b, 0xaa}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xea, 0xeb, 0x89, 0x79, 0x76, 0xb8, 0xe4, 0xc2, 0x2c, 0x96}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x82, 0xdd, 0xdd, 0xe0, 0x4f, 0xae, 0xa9, 0x60, 0xcc, 0xf5}, nil, 0)
	pt.NextEpoch()
	pt.InsertAndPrefixify([]byte{0x46, 0x76, 0x0, 0x86, 0xde, 0x18, 0xd2, 0x51, 0x13, 0xfd}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x6e, 0xb9, 0xf8, 0xa3, 0x14, 0xc9, 0x90, 0xf3, 0xa1, 0x56}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x42, 0xa9, 0x6e, 0xb4, 0x17, 0x52, 0x56, 0x33, 0x55, 0x3c}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xf8, 0x57, 0x41, 0x99, 0x4a, 0x6, 0xf3, 0xa2, 0xe9, 0x7}, nil, 0)
	pt.NextEpoch()
	pt.InsertAndPrefixify([]byte{0x5d, 0xac, 0x4c, 0x83, 0x16, 0xc1, 0x31, 0xd8, 0x40, 0x11}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x9a, 0x9a, 0x68, 0x6e, 0xaf, 0xd7, 0x7c, 0x62, 0xc5, 0x7f}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x4c, 0x3, 0xcf, 0xf4, 0xc7, 0x2c, 0x4a, 0x7b, 0x44, 0x4}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x99, 0x64, 0x9f, 0xc2, 0xea, 0xb1, 0x33, 0x9c, 0xe1, 0x51}, nil, 0)
	pt.NextEpoch()
	lookupPrefix := makePrefixFromKey([]byte{0x6e, 0xb9, 0xf8, 0xa3, 0x14, 0xc9, 0x90, 0xf3, 0xa1, 0x56})
	proof, values := pt.generateMembershipProof(
		lookupPrefix,
		pt.currEpoch,
	)
	rootHash := computeRootHashMembership(lookupPrefix, proof, values)
	if !bytes.Equal(rootHash, pt.currRoot.hash) {
		t.Error("Membership proof failed")
	}
}

func TestLookupMultipleTrace2(t *testing.T) {
	pt := NewPersistentPrefixTree()

	pt.InsertAndPrefixify([]byte{0xb6, 0x1f, 0xdd, 0xf3, 0x47, 0xdf, 0x9e, 0x63, 0x37, 0x25}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x1e, 0xd5, 0x5c, 0x23, 0x41, 0x45, 0xc5, 0xaf, 0xfd, 0x88}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x36, 0xc6, 0x79, 0x62, 0xbf, 0x44, 0xc6, 0xe0, 0xea, 0x55}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xe9, 0xaf, 0x4f, 0x28, 0xb7, 0xeb, 0x2a, 0x63, 0xb5, 0x22}, nil, 0)
	pt.NextEpoch()
	pt.InsertAndPrefixify([]byte{0xde, 0xee, 0xb8, 0xa7, 0x57, 0x3e, 0xd8, 0x6a, 0x16, 0xb2}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x91, 0x84, 0xde, 0xd6, 0xf6, 0x76, 0x8f, 0x2e, 0x1d, 0x3a}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x73, 0x7e, 0x4b, 0xed, 0xe0, 0x9a, 0xa4, 0xa4, 0x36, 0x36}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xdd, 0x24, 0x79, 0x5f, 0xb4, 0xec, 0x3, 0xc, 0x5c, 0x8a}, nil, 0)
	pt.NextEpoch()
	pt.InsertAndPrefixify([]byte{0x21, 0xd7, 0xd3, 0x80, 0xc4, 0xd2, 0xb9, 0xcc, 0xf3, 0xdf}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x26, 0xeb, 0x4b, 0x58, 0xea, 0xee, 0x39, 0x29, 0x80, 0xd7}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xc, 0x94, 0x28, 0x6f, 0x12, 0x82, 0x30, 0x29, 0x8f, 0x90}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xab, 0x9a, 0x6f, 0x88, 0x7b, 0xef, 0xb8, 0x28, 0x94, 0xab}, nil, 0)
	pt.NextEpoch()
	pt.InsertAndPrefixify([]byte{0xe3, 0x50, 0x30, 0xc0, 0xfc, 0xab, 0x54, 0xfe, 0xf5, 0xa2}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x46, 0xec, 0xf1, 0xfa, 0x97, 0x73, 0x6, 0xdb, 0xf8, 0xdd}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x56, 0x25, 0x48, 0x2f, 0xcb, 0x90, 0x12, 0x1d, 0x77, 0xf1}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x89, 0x1e, 0x64, 0xba, 0x20, 0x88, 0x7a, 0x73, 0xb, 0xda}, nil, 0)
	pt.NextEpoch()
	pt.InsertAndPrefixify([]byte{0x74, 0x12, 0x8b, 0x18, 0x69, 0x67, 0xe2, 0xc4, 0xb8, 0x55}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xb1, 0xba, 0x39, 0xa7, 0x3e, 0xd5, 0xbe, 0xd9, 0x10, 0xbf}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xe6, 0xbc, 0x2b, 0x43, 0x5f, 0x50, 0x8c, 0xd2, 0x3c, 0x86}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x17, 0x57, 0xbd, 0xb5, 0x9c, 0xda, 0x41, 0xe, 0x3e, 0xa4}, nil, 0)
	pt.NextEpoch()
	pt.InsertAndPrefixify([]byte{0xf1, 0x74, 0x63, 0xcb, 0xb6, 0xe8, 0xe7, 0x41, 0x9, 0x24}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x96, 0x43, 0x50, 0x79, 0x0, 0xf4, 0x46, 0x47, 0x7d, 0xb8}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x1a, 0xb, 0x37, 0xad, 0x5c, 0xa1, 0xdd, 0xe3, 0xba, 0x17}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xfb, 0x92, 0xe0, 0xa2, 0x71, 0x4c, 0xc8, 0xe6, 0x5d, 0x2a}, nil, 0)
	pt.NextEpoch()
	pt.InsertAndPrefixify([]byte{0xf0, 0xb4, 0xbc, 0x1c, 0x83, 0x33, 0x7, 0x38, 0xc5, 0x1}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x7, 0x51, 0x51, 0x47, 0xb3, 0x25, 0x97, 0xf8, 0xcf, 0x3a}, nil, 0)
	pt.InsertAndPrefixify([]byte{0x56, 0x8e, 0x16, 0xae, 0xa4, 0xc1, 0x97, 0xb9, 0xa7, 0x60}, nil, 0)
	pt.InsertAndPrefixify([]byte{0xc, 0x88, 0x3f, 0x6c, 0x22, 0x23, 0x1d, 0xd3, 0xb9, 0x1b}, nil, 0)
	pt.NextEpoch()

	lookupPrefix := makePrefixFromKey([]byte{0xde, 0xee, 0xb8, 0xa7, 0x57, 0x3e, 0xd8, 0x6a, 0x16, 0xb2})
	proof, values := pt.generateMembershipProof(
		lookupPrefix,
		pt.currEpoch,
	)
	rootHash := computeRootHashMembership(lookupPrefix, proof, values)
	if !bytes.Equal(rootHash, pt.currRoot.hash) {
		t.Error("Membership proof failed")
	}
}

func TestConsistentTrees(t *testing.T) {
	math_rand.Seed(1)

	prefixTree := NewPrefixTree()
	tree := NewPersistentPrefixTree()

	ins1 := toBits("010")
	tree.Insert(ins1, nil, 0)
	prefixTree.PrefixAppend(ins1, nil, 0)

	fmt.Println("=========================")
	fmt.Println(0)
	fmt.Println(prefixTree)
	fmt.Println(tree.StringAtEpoch(tree.currEpoch))
	fmt.Println("=========================")

	ins2 := toBits("011")
	tree.Insert(ins2, nil, 0)
	prefixTree.PrefixAppend(ins2, nil, 0)

	fmt.Println("=========================")
	fmt.Println(1)
	fmt.Println(prefixTree)
	fmt.Println(tree.StringAtEpoch(tree.currEpoch))
	fmt.Println("=========================")

	ins3 := toBits("001")
	tree.Insert(ins3, nil, 0)
	prefixTree.PrefixAppend(ins3, nil, 0)

	fmt.Println("=========================")
	fmt.Println(2)
	fmt.Println(prefixTree)
	fmt.Println(tree.StringAtEpoch(tree.currEpoch))
	fmt.Println("=========================")

	// ins4 := toBits("000")
	// tree.Insert(ins4, nil, 0)
	// prefixTree.PrefixAppend(ins4, nil, 0)

	// fmt.Println("=========================")
	// fmt.Println(3)
	// fmt.Println(prefixTree)
	// fmt.Println(tree.StringAtEpoch(tree.currEpoch))
	// fmt.Println("=========================")

	fmt.Printf("prefix tree: %x\n", prefixTree.getHash())
	fmt.Printf("persistent prefix tree: %x\n", tree.getHash(tree.currEpoch))
	fmt.Println(prefixTree)
	fmt.Println(tree.StringAtEpoch(tree.currEpoch))
}

func toBits(s string) []byte {
	bs := []byte(s)
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		if bs[i] == '0' {
			result[i] = 0
		} else {
			result[i] = 1
		}
	}
	return result
}
