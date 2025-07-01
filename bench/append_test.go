package bench

import (
	"testing"

	"github.com/huyuncong/MerkleSquare/core"
)

func helperAppendMerkleSquare(b *testing.B, depth uint32, numAppends uint32) {
	b.StopTimer()
	inserts := getKeyValueSigPairs(numAppends)
	start := numAppends - 100

	for i := 0; i < b.N; i++ {
		ms := core.NewMerkleSquare(depth)
		for j, ins := range *inserts {
			if uint32(j) >= start {
				b.StartTimer()
				ms.Append(ins.Key, ins.Value, ins.Signature)
				b.StopTimer()
			} else {
				ms.Append(ins.Key, ins.Value, ins.Signature)
			}
		}
	}
}

// for append benchmarking, can keep depth fixed at 31 or go

func BenchmarkAppendMerkleSquare2(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 2)
}

func BenchmarkAppendMerkleSquare4(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 4)
}

func BenchmarkAppendMerkleSquare8(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 8)
}
func BenchmarkAppendMerkleSquare16(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 16)
}

func BenchmarkAppendMerkleSquare32(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 32)
}

func BenchmarkAppendMerkleSquare64(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 64)
}

func BenchmarkAppendMerkleSquare128(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 128)
}

func BenchmarkAppendMerkleSquare256(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 256)
}

func BenchmarkAppendMerkleSquare512(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 512)
}

func BenchmarkAppendMerkleSquare1024(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 1024)
}

func BenchmarkAppendMerkleSquare2048(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 2048)
}

func BenchmarkAppendMerkleSquare4096(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 4096)
}

func BenchmarkAppendMerkleSquare8192(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 8192)
}

func BenchmarkAppendMerkleSquare16384(b *testing.B) {
	helperAppendMerkleSquare(b, 20, 16384)
}

func BenchmarkAppendMerkleSquare32768(b *testing.B) {
	helperAppendMerkleSquare(b, 20, 32768)
}

func BenchmarkAppendMerkleSquare65536(b *testing.B) {
	helperAppendMerkleSquare(b, 20, 65536)
}

func BenchmarkAppendMerkleSquare131072(b *testing.B) {
	helperAppendMerkleSquare(b, 20, 131072)
}

func BenchmarkAppendMerkleSquare262144(b *testing.B) {
	helperAppendMerkleSquare(b, 20, 262144)
}

func BenchmarkAppendMerkleSquare524288(b *testing.B) {
	helperAppendMerkleSquare(b, 20, 524288)
}

func BenchmarkAppendMerkleSquare1048576(b *testing.B) {
	helperAppendMerkleSquare(b, 31, 1048576)
}
