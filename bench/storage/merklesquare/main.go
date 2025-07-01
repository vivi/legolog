package main

import (
	"bufio"
	"fmt"
	"math"
	"os"

	"github.com/huyuncong/MerkleSquare/bench"
	"github.com/huyuncong/MerkleSquare/core"
)

func main() {

	var depth uint32
	depth = 31
	numAppends := uint32(math.Pow(2, float64(20)))

	ms := core.NewMerkleSquare(depth)

	var inserts *[]bench.Inserts
	inserts = bench.GetKeyValueSigPairs(numAppends)

	for _, ins := range *inserts {
		ms.Append(ins.Key, ins.Value, ins.Signature)
	}

	fmt.Println("finished!")
	fmt.Println("Press enter to stop.")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
}
