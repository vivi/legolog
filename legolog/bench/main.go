package main

import (
	// "fmt"

	// "github.com/huyuncong/MerkleSquare/core"
	"fmt"
	"sync"
	"time"

	bench "github.com/huyuncong/MerkleSquare/legolog/bench/microbench"
)

func main() {
	startTime := time.Now()

	// partitionList1 := []int{1, 8, 64, 128, 256, 512}
	// partitionList2 := []int{500, 1000, 2000}
	// partitionList3 := []int{4000, 8000}
	partitionList1 := []int{5}
	partitionList2 := []int{10}
	partitionList3 := []int{15}

	// bench.BenchmarkAuditorVaryingPartitions(partitionList1)
	// return
	bench.MeasureLookupsOverMultipleVerificationPeriods([]int{1}, true)
	return
	// bench.RealMeasureLookups([]int{5}, true)
	// return

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.RealMeasureLookups(partitionList1, false)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.RealMeasureLookups(partitionList1, true)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.RealMeasureLookups(partitionList2, false)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.RealMeasureLookups(partitionList2, true)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.RealMeasureLookups(partitionList3, false)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.RealMeasureLookups(partitionList3, true)
	}()

	wg.Wait()

	// DEBUGGING
	// bench.RealMeasureLookups([]int{20000}, true)

	endTime := time.Now()
	fmt.Printf("Total time: %f", endTime.Sub(startTime).Seconds())
}

/*func main() {
	startTime := time.Now()

	var wg sync.WaitGroup

	partitionList1 := []int{125}
	partitionList2 := []int{250}
	partitionList3 := []int{500}
	partitionList4 := []int{1000}
	partitionList5 := []int{2000}
	partitionList6 := []int{4000, 8000}

	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.BenchmarkAuditorVaryingPartitions(partitionList1)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.BenchmarkAuditorVaryingPartitions(partitionList2)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.BenchmarkAuditorVaryingPartitions(partitionList3)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.BenchmarkAuditorVaryingPartitions(partitionList4)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.BenchmarkAuditorVaryingPartitions(partitionList5)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		bench.BenchmarkAuditorVaryingPartitions(partitionList6)
	}()

	wg.Wait()

	// DEBUGGING
	// bench.BenchmarkAuditorVaryingPartitions([]int{1000})

	endTime := time.Now()

	fmt.Printf("Total time taken: %f\n", endTime.Sub(startTime).Seconds())
}
*/
