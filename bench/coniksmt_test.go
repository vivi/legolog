package bench

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	coniksmt "github.com/huyuncong/MerkleSquare/coniks/coniks-mt"
)

func getLookupPairs(inserts []Coniks_insert, numTot uint32, numPairs uint32) []Coniks_insert {

	res := []Coniks_insert{}

	for i := uint32(0); i < numPairs; i++ {
		id := rand.Intn(int(numTot))
		ins := inserts[id]
		res = append(res, ins)
	}

	return res
}

func getLookupProofs(lookups []Coniks_insert, testM *coniksmt.MerkleTree) []*coniksmt.AuthenticationPath {

	res := make([]*coniksmt.AuthenticationPath, 0)

	totSize := 0

	for _, ins := range lookups {
		proof := testM.Get(ins.Index)

		// fmt.Println(proof.Size())
		totSize = totSize + proof.Size()

		res = append(res, proof)
	}

	fmt.Println()
	fmt.Printf("avg lookup proof size: %f\n", float64(totSize)/float64(len(lookups)))

	return res
}

func helperNewCONIKSMT(b *testing.B, inserts []Coniks_insert) *coniksmt.MerkleTree {
	testM, err := coniksmt.NewMerkleTree()
	if err != nil {
		b.Fatal(err)
	}

	for _, ins := range inserts {
		if err := testM.Set(ins.Index, string(ins.Key), ins.Value); err != nil {
			b.Fatal(err)
		}
	}
	// testM.recomputeHash()

	return testM
}

func helperAppendCONIKS(b *testing.B, numAppends uint32, numNew uint32) {
	b.StopTimer()

	inserts := GetIndexKeyValuePairs(numAppends)
	test_inserts := GetIndexKeyValuePairs(numNew)

	initM := helperNewCONIKSMT(b, inserts)

	for i := 0; i < b.N; i++ {
		testM := initM.Clone()

		b.StartTimer()
		for _, ins := range test_inserts {
			if err := testM.Set(ins.Index, string(ins.Key), ins.Value); err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()
	}
}

func helperLookUpVerificationCONIKS(b *testing.B, numAppends uint32, numLookup uint32) {
	b.StopTimer()

	rand.Seed(time.Now().Unix())
	inserts := GetIndexKeyValuePairs(numAppends)
	lookups := getLookupPairs(inserts, numAppends, numLookup)

	initM := helperNewCONIKSMT(b, inserts)
	testM := initM.CloneDiff()

	lookupProofs := getLookupProofs(lookups, testM)

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		for idx, proof := range lookupProofs {
			// err := proof.Verify(lookups[idx].key, lookups[idx].value, testM.RootHash())
			// if err != nil {
			// 	b.Fatal(err)
			// }
			proof.Verify(lookups[idx].Key, lookups[idx].Value, testM.RootHash())
		}
		b.StopTimer()
	}
}

// =================================================================================

func BenchmarkAppendCONIKS128(b *testing.B) {
	helperAppendCONIKS(b, 128, 100)
}

func BenchmarkAppendCONIKS100(b *testing.B) {
	helperAppendCONIKS(b, 100, 100)
}

func BenchmarkAppendCONIKS256(b *testing.B) {
	helperAppendCONIKS(b, 256, 100)
}

func BenchmarkAppendCONIKS512(b *testing.B) {
	helperAppendCONIKS(b, 512, 100)
}

func BenchmarkAppendCONIKS1000(b *testing.B) {
	helperAppendCONIKS(b, 1000, 100)
}

func BenchmarkAppendCONIKS1024(b *testing.B) {
	helperAppendCONIKS(b, 1024, 100)
}

func BenchmarkAppendCONIKS2048(b *testing.B) {
	helperAppendCONIKS(b, 2048, 100)
}

func BenchmarkAppendCONIKS4096(b *testing.B) {
	helperAppendCONIKS(b, 4096, 100)
}

func BenchmarkAppendCONIKS8192(b *testing.B) {
	helperAppendCONIKS(b, 8192, 100)
}

func BenchmarkAppendCONIKS10000(b *testing.B) {
	helperAppendCONIKS(b, 10000, 100)
}

func BenchmarkAppendCONIKS16384(b *testing.B) {
	helperAppendCONIKS(b, 16384, 100)
}

func BenchmarkAppendCONIKS100000(b *testing.B) {
	helperAppendCONIKS(b, 100000, 100)
}

func BenchmarkAppendCONIKS131072(b *testing.B) {
	helperAppendCONIKS(b, 131072, 100)
}

func BenchmarkAppendCONIKS1000000(b *testing.B) {
	helperAppendCONIKS(b, 1000000, 100)
}

func BenchmarkAppendCONIKS1048576(b *testing.B) {
	helperAppendCONIKS(b, 1048576, 100)
}

//==============================================================================

func BenchmarkLookUpCONIKS100(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 100, 100)
}

func BenchmarkLookUpCONIKS128(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 128, 100)
}

func BenchmarkLookUpCONIKS256(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 256, 100)
}

func BenchmarkLookUpCONIKS512(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 512, 100)
}

func BenchmarkLookUpCONIKS1000(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 1000, 100)
}

func BenchmarkLookUpCONIKS1024(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 1024, 100)
}

func BenchmarkLookUpCONIKS2048(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 2048, 100)
}

func BenchmarkLookUpCONIKS4096(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 4096, 100)
}

func BenchmarkLookUpCONIKS8192(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 8192, 100)
}

func BenchmarkLookUpCONIKS10000(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 10000, 100)
}

func BenchmarkLookUpCONIKS16384(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 16384, 100)
}

func BenchmarkLookUpCONIKS32768(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 32768, 100)
}

func BenchmarkLookUpCONIKS65536(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 65536, 100)
}

func BenchmarkLookUpCONIKS100000(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 100000, 100)
}

func BenchmarkLookUpCONIKS131072(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 131072, 100)
}

func BenchmarkLookUpCONIKS262144(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 262144, 100)
}

func BenchmarkLookUpCONIKS524288(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 524288, 100)
}

func BenchmarkLookUpCONIKS1000000(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 1000000, 100)
}

func BenchmarkLookUpCONIKS1048576(b *testing.B) {
	helperLookUpVerificationCONIKS(b, 1048576, 100)
}
