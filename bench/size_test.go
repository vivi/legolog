package bench

import (
	"fmt"
	"testing"

	"github.com/huyuncong/MerkleSquare/core"
)

func helperMerkleSquareSize(b *testing.B, depth uint32, numAppends uint32) {
	total := float64(0)
	count := float64(0)

	for i := 0; i < b.N; i++ {
		ms := core.NewMerkleSquare(depth)
		inserts := getKeyValueSigPairs(numAppends)

		for _, ins := range *inserts {
			ms.Append(ins.Key, ins.Value, ins.Signature)
		}

		total += float64(ms.GetMerkleSquareSize())
		count++
	}
	averageSize := total / count
	b.ReportMetric(averageSize, fmt.Sprintf("AvgForDepth%dNumAppends%d", depth, numAppends))
}

// for append benchmarking, can keep depth fixed at 31 or go

func BenchmarkMerkleSquareSize(b *testing.B) {
	helperMerkleSquareSize(b, maxTreeDepth, treeSize)
}
