package bench

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/huyuncong/MerkleSquare/core"
)

func helperAppendToTree(ms *core.MerkleSquare, numAppends int) {
	for i := 0; i < numAppends; i++ {
		ms.Append([]byte(fmt.Sprintf("key%d", i)), []byte(fmt.Sprintf("value%d", i)), []byte(fmt.Sprintf("signature%d", i)))
	}
}

func helperRunProofs(ms *core.MerkleSquare, size int) {
	for i := 0; i < size; i += 3 {
		ms.GenerateExtensionProof(uint32(i), uint32(rand.Intn(size-i)+i))
	}
}

func helperConcurrentAppendProof(b *testing.B, depth uint32, numAppends int) {
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		ms := core.NewMerkleSquare(depth)

		b.StartTimer()
		helperAppendToTree(ms, numAppends)
		helperRunProofs(ms, numAppends)

		b.StopTimer()
	}
}

func BenchmarkMap2048(b *testing.B) {
	helperConcurrentAppendProof(b, 31, 2048)
}
