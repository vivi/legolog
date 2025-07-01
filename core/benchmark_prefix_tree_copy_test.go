package core

import (
	"testing"
)

func BenchmarkPrefixTreeCopy(b *testing.B) {
	/* TODO: fix me if we ever use prefix tree again.
	b.ResetTimer()
	partition := NewPartition()
	numAppends := 20000
	startTime := time.Now()
	for i := 0; i < numAppends; i++ {
		partition.Append([]byte{byte(i)}, []byte{byte(i)}, []byte{byte(i)}, []byte{byte(i)})
	}
	appendTime := time.Since(startTime)
	startTime = time.Now()
	newTree, err := partition.nextVerificationBaseTree.copy()
	if err != nil {
		b.Errorf("Error copying base tree: %v", err)
	}
	_ = newTree
	baselineCopyTime := time.Since(startTime)
	startTime = time.Now()
	newTree, err = partition.nextVerificationBaseTree.copyFast()
	if err != nil {
		b.Errorf("Error copying base tree: %v", err)
	}
	_ = newTree
	fastCopyTime := time.Since(startTime)

	fmt.Printf("Time to append %d elements to tree: %v\n", numAppends, appendTime)
	fmt.Printf("Baseline copy time: %v\n", baselineCopyTime)
	fmt.Printf("Fast copy time: %v\n", fastCopyTime)
	*/
}