package bench

import (
	"testing"

	"github.com/huyuncong/MerkleSquare/core"
)

type prefix_insert struct {
	prefix    []byte
	valueHash []byte
}

const PREFIXBYTESIZE = 32
const VALUEHASHSIZE = 32

func helperAppendPrefix(b *testing.B, numAppends uint32) {
	b.StopTimer()
	inserts := getPrefixKeyValueSigPairs(numAppends)

	for i := 0; i < b.N; i++ {
		pTree := core.NewPrefixTree()
		var pos int
		var ins prefix_insert
		b.StartTimer()
		for pos, ins = range inserts {
			pTree.PrefixAppend(ins.prefix, ins.valueHash, uint32(pos))
		}
		b.StopTimer()
	}
}

func BenchmarkAppendPrefix16(b *testing.B) {
	helperAppendPrefix(b, 16)
}

func BenchmarkAppendPrefix128(b *testing.B) {
	helperAppendPrefix(b, 128)
}

func BenchmarkAppendPrefix1024(b *testing.B) {
	helperAppendPrefix(b, 1024)
}

func BenchmarkAppendPrefix8192(b *testing.B) {
	helperAppendPrefix(b, 8192)
}

func BenchmarkAppendPrefix16384(b *testing.B) {
	helperAppendPrefix(b, 16384)
}

func BenchmarkAppendPrefix131072(b *testing.B) {
	helperAppendPrefix(b, 131072)
}

func BenchmarkAppendPrefix1048576(b *testing.B) {
	helperAppendPrefix(b, 1048576)
}

//*******************************
// HELPER METHODS
//*******************************

func getPrefixKeyValueSigPairs(numPairs uint32) []prefix_insert {

	res := []prefix_insert{}

	for i := uint32(0); i < numPairs; i++ {

		prefix := core.ConvertBitsToBytes(GenerateRandomByteArray(PREFIXBYTESIZE))
		valueHash := GenerateRandomByteArray(VALUEHASHSIZE)

		ins := prefix_insert{
			prefix:    prefix,
			valueHash: valueHash,
		}

		res = append(res, ins)
	}

	return res
}
