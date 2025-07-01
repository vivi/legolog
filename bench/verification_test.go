package bench

import (
	"fmt"
	"math"
	"testing"

	"github.com/huyuncong/MerkleSquare/core"
)

func helperVerifyExistenceProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert) {

	oldDigest := ms.GetOldDigest(ms.Size)

	b.StopTimer()
	for i := 0; i < b.N; i++ {

		for j, ins := range *inserts {

			proof := ms.GenerateExistenceProof(ins.Key, uint32(j), 0, ms.Size)
			nodeHash := core.ComputeLeafNodeHash(ins.Key, ins.Value, ins.Signature, uint32(j))
			oldKeys := []core.KeyHash{core.KeyHash{
				Hash: core.ComputeContentHash(ins.Key, ins.Value, ins.Signature, uint32(j)),
				Pos:  uint32(j),
			}}

			b.StartTimer()
			core.VerifyExistenceProof(oldDigest, nodeHash, ins.Key, uint32(j), 0, proof, oldKeys)
			b.StopTimer()
		}
	}

	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperVerifyExtensionProof(b *testing.B, ms *core.MerkleSquare, n int) {

	b.StopTimer()

	proof := ms.GenerateExtensionProof(uint32(math.Pow10(n)), uint32(math.Pow10(n+1))) // CHANGE THIS FOR BASE 2: ms.GenerateExtensionProof((1<<j)-1, (1<<(j+1))-1)
	oldDigest := ms.GetOldDigest(uint32(math.Pow10(n)))
	newDigest := ms.GetOldDigest(uint32(math.Pow10(n + 1)))
	// proof := ms.GenerateExtensionProof((1<<n)-1, (1<<(n+1))-1)
	// oldDigest := ms.GetOldDigest((1 << n) - 1)
	// newDigest := ms.GetOldDigest((1 << (n + 1)) - 1)

	fmt.Println("tree size", ms.Size, "old digest size", oldDigest.Size, "new digest size", newDigest.Size)

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		core.VerifyExtensionProof(oldDigest, newDigest, proof)
		b.StopTimer()
	}

	b.ReportMetric(float64(n), "n")
	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperVerifyMKProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert) {

	b.StopTimer()
	digest := ms.GetOldDigest(ms.Size)
	for i := 0; i < b.N; i++ {
		for j, ins := range *inserts {
			proof := ms.ProveFirst(ins.Key, ins.Value, uint32(j), ms.Size)

			b.StartTimer()
			core.VerifyMKProof(digest, ins.Key, ins.Value, ins.Signature, uint32(j), proof)
			b.StopTimer()
		}
	}

	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperVerifyLatestPKProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert) {

	b.StopTimer()
	for i := 0; i < b.N; i++ {

		oldDigest := ms.GetOldDigest(ms.Size)
		for j, ins := range *inserts {
			proof := ms.ProveLatest(ins.Key, ins.Value, uint32(j), ms.Size)

			b.StartTimer()
			core.VerifyPKProof(oldDigest, ins.Key, ins.Value, ins.Signature, uint32(j), proof)
			b.StopTimer()
		}
	}

	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperVerifyNonexistenceProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert) {

	b.StopTimer()
	for i := 0; i < b.N; i++ {

		for j, ins := range *inserts {
			proof := ms.ProveNonexistence(ins.Key, uint32(j), uint32(j+1))
			digest := ms.GetOldDigest(uint32(j + 1))

			b.StartTimer()
			core.VerifyNonexistenceProof(ins.Key, uint32(j), digest, proof)
			b.StopTimer()
		}
	}

	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperVerifyBatchedLookupProof(b *testing.B, ms *core.MerkleSquare, inserts *[]insert, key []byte, keyPositions []uint32) {
	b.StopTimer()

	proof, err := ms.GenerateBatchedLookupProof(0, ms.Size, key, keyPositions)
	if err != nil {
		panic(err)
	}
	digest := ms.GetOldDigest(ms.Size)
	keyHashes := getKeyHashesFromInserts(inserts, &keyPositions)

	for i := 0; i < b.N; i++ {
		b.StartTimer()
		core.VerifyBatchedLookupProof(0, key, *keyHashes, digest, proof)
		b.StopTimer()
	}

	b.ReportMetric(float64(ms.Size), "treesize")
}

func helperVerifyExistenceProofWithCaching(b *testing.B, ms *core.MerkleSquare, inserts *[]insert, epochSize uint32, offlineEpochs uint32) {
	b.StopTimer()

	timeOffline := offlineEpochs * epochSize
	oldDigest := ms.GetOldDigest(ms.Size)

	for i := 0; i < b.N; i++ {

		for k, ins := range *inserts {
			if uint32(k) >= epochSize {
				break
			}

			leafHash := core.ComputeLeafNodeHash(ins.Key, ins.Value, ins.Signature, uint32(k))
			oldKeys := []core.KeyHash{core.KeyHash{
				Hash: core.ComputeContentHash(ins.Key, ins.Value, ins.Signature, uint32(k)),
				Pos:  uint32(k),
			}}

			var height uint32
			var nodeHash []byte
			if timeOffline <= ms.Size {
				height = 0
				nodeHash = leafHash
			} else {
				height = core.GetOldDepth(uint32(k), ms.Size-timeOffline)

				oldProof := ms.GenerateExistenceProof(ins.Key, uint32(k), 0, ms.Size-timeOffline)
				otherDigest := ms.GetOldDigest(ms.Size - timeOffline)
				_, nodeHash, _ = core.VerifyExistenceProof(otherDigest, leafHash, ins.Key, uint32(k), 0, oldProof, oldKeys)
			}

			proof := ms.GenerateExistenceProof(ins.Key, uint32(k), height, ms.Size)

			b.StartTimer()
			core.VerifyExistenceProof(oldDigest, nodeHash, ins.Key, uint32(k), height, proof, oldKeys)
			b.StopTimer()
		}
	}

	b.ReportMetric(float64(offlineEpochs), "offlineEpochs")
	b.ReportMetric(float64(epochSize), "epochSize")
	b.ReportMetric(float64(ms.Size), "treesize")
}

func BenchmarkVerifyExistenceProof(b *testing.B) {
	helperVerifyExistenceProof(b, ms, inserts)
}

func BenchmarkVerifyExtensionProof(b *testing.B) {
	helperVerifyExtensionProof(b, ms, extensionN) // CHANGE THIS FOR BASE 2: 6 should be maxTreeDepth
}

func BenchmarkVerifyMKProof(b *testing.B) {
	helperVerifyMKProof(b, ms, inserts)
}

func BenchmarkVerifyPKProof(b *testing.B) {
	helperVerifyLatestPKProof(b, ms, inserts)
}

func BenchmarkVerifyNonexistenceProof(b *testing.B) {
	helperVerifyNonexistenceProof(b, ms, inserts)
}

func BenchmarkVerifyBatchedLookupProof(b *testing.B) {
	helperVerifyBatchedLookupProof(b, ms, inserts, batchedLookupKey, batchedLookupKeyPositions)
}

func BenchmarkVerifyExistenceWithCaching(b *testing.B) {
	helperVerifyExistenceProofWithCaching(b, ms, inserts, epochSize, offlineEpochs)
}
