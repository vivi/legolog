package main

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/coniks-sys/coniks-go/crypto/sign"
	"github.com/coniks-sys/coniks-go/crypto/vrf"
	"github.com/huyuncong/MerkleSquare/bench"
	coniksmt "github.com/huyuncong/MerkleSquare/coniks/coniks-mt"
)

var signKey sign.PrivateKey
var vrfKey vrf.PrivateKey

const numAppend = 1048576
const appPerEpoch = 256
const pastEpoch = numAppend/appPerEpoch - 1

type TestAd struct {
	data string
}

func (t TestAd) Serialize() []byte {
	return []byte(t.data)
}

func helperNewPAD(inserts []bench.Coniks_insert) *coniksmt.PAD {
	var err error
	signKey, err = sign.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	vrfKey, err = vrf.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	testP, err := coniksmt.NewPAD(TestAd{""}, signKey, vrfKey, 5000)
	if err != nil {
		panic(err)
	}

	cnt := 0

	for _, ins := range inserts {
		cnt++
		if err := testP.Set(string(ins.Key), ins.Value); err != nil {
			panic(err)
		}

		if cnt == 256 {
			cnt = 0
			testP.Update(TestAd{""})
		}
	}
	// testM.recomputeHash()

	return testP
}

func main() {
	inserts := bench.GetIndexKeyValuePairs(numAppend)
	testP := helperNewPAD(inserts)

	monitoring := inserts[0]

	mp, err := testP.LookupInPastEpochs(string(monitoring.Key), testP.LatestEpoch(), pastEpoch, 0)
	if err != nil {
		panic(err)
	}

	roots := make([]*coniksmt.SignedTreeRoot, 0)
	for epoch := testP.LatestEpoch(); epoch >= testP.LatestEpoch()-pastEpoch; epoch-- {
		root := testP.GetSTR(epoch)
		roots = append(roots, root)
	}

	fmt.Printf("Tree Size: %d\n", numAppend)
	fmt.Printf("Appends per Epoch: %d\n", appPerEpoch)
	fmt.Println("Proof Size:")
	proofSize := binary.Size(mp.VrfProof) 
	// Remember to add root hashes
	for i, ap := range mp.APs {
		proofSize = proofSize + ap.Size()
		fmt.Printf("#epochs: %d; proof size: %d Bytes\n", i+1, proofSize)
	}

	fmt.Println()

	fmt.Println("Verification Time:")

	res := make([]int64, 5000)

	start := time.Now()
	for i, ap := range mp.APs {
		// err := ap.Verify(monitoring.Key, monitoring.Value, roots[i].TreeHash)
		// if err != nil {
		// 	panic(err)
		// }
		// bl := roots[i].VerifyHashChain(roots[i+1])
		// if !bl {
		// 	panic(errors.New("wrong roots"))
		// }

		ap.Verify(monitoring.Key, monitoring.Value, roots[i].TreeHash)
		roots[i].VerifyHashChain(roots[i+1])
		// fmt.Println("!")
		res[i] = time.Since(start).Nanoseconds()
		// fmt.Println(time.Since(start))
	}

	for i := 0; i < pastEpoch; i++ {
		r := res[i]
		fmt.Printf("#epochs: %d; monitoring time: %d ns\n", i+1, r)
	}
}
