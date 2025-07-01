package bench

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/huyuncong/MerkleSquare/core"
)

//*******************************
// TEST MAIN
//*******************************

var ms *core.MerkleSquare
var inserts *[]insert

var maxTreeDepth uint32
var treeSize uint32
var epochSize uint32
var offlineEpochs uint32

var batchedLookupKey []byte
var batchedLookupKeyPositions []uint32

var extensionN int

func skipTestMain(m *testing.M) {

	//maxTreeDepth = 20 // SET MAX TREE DEPTH FOR GENERATED TREES
	//treeSize := (1 << maxTreeDepth) // SET TREE SIZE FOR GENERATED TREES
	//epochSize = 256 // SET EPOCH SIZE

	for i := 1; i < 21; i++ {
		maxTreeDepth = uint32(i)
		treeSize = uint32(math.Pow(2, float64(i)))
		res := testing.Benchmark(BenchmarkMerkleSquareSize)
		err := writeBenchmarkResultToFile(&res, fmt.Sprintf("../bench/latest_results/merkleSquareSize/%d-%d", maxTreeDepth, treeSize))
		if err != nil {
			panic(err)
		}
	}
}

//*******************************
// UTILS TEST
//*******************************

func skipTestCreateAndStoreTrees(t *testing.T) {

	i := 3

	createAndStoreTree(i)

	for j := 1; j <= i; j++ {
		ms, inserts, err := loadTreeAndKeys(j)
		if err != nil {
			t.Error()
		}

		oldDigest := ms.GetOldDigest(1 << j)

		for p, insert := range *inserts {

			existenceProof := ms.GenerateExistenceProof(insert.Key, uint32(p), 0, 1<<j)
			nodeHash := core.ComputeLeafNodeHash(insert.Key, insert.Value, insert.Signature, uint32(p))
			otherHashes := []core.KeyHash{core.KeyHash{
				Hash: core.ComputeContentHash(insert.Key, insert.Value, insert.Signature, uint32(p)),
				Pos:  uint32(p),
			}}
			verify, _, _ := core.VerifyExistenceProof(oldDigest, nodeHash, insert.Key, uint32(p), 0, existenceProof, otherHashes)

			if !verify {
				t.Error()
			}
		}
	}
}

func skipTestSerializeInserts(t *testing.T) {

	a := []insert{
		insert{
			Key:       []byte("1"),
			Value:     []byte("2"),
			Signature: []byte("4"),
		},
		insert{
			Key:       []byte("10"),
			Value:     []byte("15"),
			Signature: []byte("20"),
		},
	}

	buf, err := serializeInserts(&a)

	if err != nil {
		t.Error(err)
	}

	res, err := deserializeInserts(buf)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(*res, a) {
		t.Error(res, a)
	}
}

func skipTestCreateRepeatedValues(t *testing.T) {

	inserts := getKeyValueSigPairs(10)

	for _, val := range *inserts {
		fmt.Println(val.Key[:10])
	}

	key, keyPositions := createRepeatedValues(inserts, 5)
	for _, val := range *inserts {
		fmt.Println(val.Key[:10])
	}

	fmt.Println(key[:10])
	fmt.Println(keyPositions)
}

func skipTestGetKeyHashFromInserts(t *testing.T) {

	inserts := getKeyValueSigPairs(10)

	_, keyPositions := createRepeatedValues(inserts, 5)

	res := getKeyHashesFromInserts(inserts, &keyPositions)

	fmt.Println(res)
}
