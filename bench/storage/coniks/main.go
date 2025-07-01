package main

import (
	"bufio"
	"fmt"
	"math"
	"os"

	"github.com/coniks-sys/coniks-go/crypto/sign"
	"github.com/coniks-sys/coniks-go/crypto/vrf"
	"github.com/huyuncong/MerkleSquare/bench"
	coniksmt "github.com/huyuncong/MerkleSquare/coniks/coniks-mt"
)

var signKey sign.PrivateKey
var vrfKey vrf.PrivateKey

type TestAd struct {
	data string
}

func (t TestAd) Serialize() []byte {
	return []byte(t.data)
}

func helperNewPAD(inserts []bench.Coniks_insert) *coniksmt.PAD {
	var err error
	signKey, err := sign.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	vrfKey, err := vrf.GenerateKey(nil)
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

	var depth uint32
	depth = 20
	numAppends := uint32(math.Pow(2, float64(depth)))

	inserts := bench.GetIndexKeyValuePairs(numAppends)
	_ = helperNewPAD(inserts)

	fmt.Println("finished!")
	fmt.Println("Press enter to stop.")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}
