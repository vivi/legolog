package bench

import (
	"fmt"
	"math"
	"testing"

	"github.com/huyuncong/MerkleSquare/core"
)

func helperExistenceProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert) {

	total := float64(0)
	count := float64(0)
	for i := 0; i < b.N; i++ {
		for j, ins := range *inserts {
			proof := ms.GenerateExistenceProof(ins.Key, uint32(j), 0, ms.Size)
			size := checkExistenceProofSize(proof)
			total += float64(size)
			count++
		}
	}

	avgProofSize := total / count
	b.ReportMetric(avgProofSize, "avgSize")
	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperExtensionProof(b *testing.B, ms *core.MerkleSquare, n int) {

	total := float64(0)
	count := float64(0)

	for i := 0; i < b.N; i++ {
		for j := 1; j < n; j++ {
			//proof := ms.GenerateExtensionProof((1<<j)-1, (1<<(j+1))-1)
			proof := ms.GenerateExtensionProof(uint32(math.Pow10(j)), uint32(math.Pow10(j+1))) // CHANGE THIS FOR BASE 2: ms.GenerateExtensionProof((1<<j)-1, (1<<(j+1))-1)
			size := checkExtensionProofSize(proof)

			b.ReportMetric(float64(size), fmt.Sprintf("iValue%d", j))
			total += float64(size)
			count++
		}
	}

	avgProofSize := total / count
	b.ReportMetric(avgProofSize, "avgSize")
	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperMKProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert) {

	total := float64(0)
	count := float64(0)

	for i := 0; i < b.N; i++ {

		for j, ins := range *inserts {
			proof := ms.ProveFirst(ins.Key, ins.Value, uint32(j), ms.Size)
			size := checkMKProofSize(proof)
			total += float64(size)
			count++
		}
	}

	avgProofSize := total / count
	b.ReportMetric(avgProofSize, "avgSize")
	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperLatestPKProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert) {

	total := float64(0)
	count := float64(0)

	for i := 0; i < b.N; i++ {
		for j, ins := range *inserts {
			proof := ms.ProveLatest(ins.Key, ins.Value, uint32(j), ms.Size)
			size := checkLatestPKProofSize(proof)
			total += float64(size)
			count++
		}
	}

	avgProofSize := total / count
	b.ReportMetric(avgProofSize, "avgSize")
	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperNonexistenceProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert) {

	total := float64(0)
	count := float64(0)

	for i := 0; i < b.N; i++ {
		for j, ins := range *inserts {
			proof := ms.ProveNonexistence(ins.Key, uint32(j), uint32(j+1))
			size := checkNonexistenceProofSize(proof)
			total += float64(size)
			count++
		}
	}

	avgProofSize := total / count
	b.ReportMetric(avgProofSize, "avgSize")
	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperBatchedLookupProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert, key []byte, keyPositions []uint32) {

	total := float64(0)
	count := float64(0)

	for i := 0; i < b.N; i++ {
		proof, err := ms.GenerateBatchedLookupProof(0, ms.Size, key, keyPositions)
		if err != nil {
			panic(err)
		}

		size := checkBatchedLookupProofSize(proof)
		total += float64(size)
		count++
	}

	avgProofSize := total / count
	b.ReportMetric(avgProofSize, "avgSize")
	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperExistenceProofWithCaching(b *testing.B, ms *core.MerkleSquare, inserts *[]insert, epochSize uint32) {

	for i := 0; i < b.N; i++ {
		for j := 1; uint32(j) <= (ms.Size / epochSize); j = j << 1 {

			timeOffline := uint32(j) * epochSize
			total := float64(0)
			count := float64(0)

			for k, ins := range *inserts {
				if uint32(k) >= epochSize {
					break
				}

				var height uint32
				if timeOffline <= ms.Size {
					height = 0
				} else {
					height = core.GetOldDepth(uint32(k), ms.Size-timeOffline)
				}
				pf := ms.GenerateExistenceProof(ins.Key, uint32(k), height, ms.Size)
				size := checkExistenceProofSize(pf)

				total += float64(size)
				count++
			}

			averageProofSize := total / count
			b.ReportMetric(averageProofSize, fmt.Sprintf("AvgForSleepTime%d", j))
		}
	}
	b.ReportMetric(float64(ms.Size), "treesize")
}

func BenchmarkExistenceProof(b *testing.B) {
	helperExistenceProof(b, ms, inserts)
}

func BenchmarkExtensionProof(b *testing.B) {
	helperExtensionProof(b, ms, extensionN) // CHANGE THIS FOR BASE 2: 6 should be maxTreeDepth
}

func BenchmarkMKProof(b *testing.B) {
	helperMKProof(b, ms, inserts)
}

func BenchmarkPKProof(b *testing.B) {
	helperLatestPKProof(b, ms, inserts)
}

func BenchmarkNonexistenceProof(b *testing.B) {
	helperNonexistenceProof(b, ms, inserts)
}

func BenchmarkBatchedLookupProof(b *testing.B) {
	helperBatchedLookupProof(b, ms, inserts, batchedLookupKey, batchedLookupKeyPositions)
}

func BenchmarkExistenceProofWithCaching(b *testing.B) {
	helperExistenceProofWithCaching(b, ms, inserts, epochSize)
}
