package bench

import (
	"context"
	"io/ioutil"
	"os"
	"sort"
	"testing"

	"crypto/rand"
	"fmt"

	// math_rand "math/rand"

	"time"

	"github.com/huyuncong/MerkleSquare/core"

	"github.com/huyuncong/MerkleSquare/lib/storage"
	"github.com/immesys/bw2/crypto"
)

func monitorBenchmarkWrapper(b *testing.B, numPartitions int, numAppends int, numVerificationPeriodsOffline int, aggHistory bool) float64 {
	numTrials := 20
	return helperMonitorMicrobenchDebug(b, numPartitions, numAppends, numVerificationPeriodsOffline, aggHistory, numTrials)
}

/*
this function isn't used
func helperMonitorMicrobench(b *testing.B, numPartitions, numAppends int, numVerificationPeriodsOffline int, aggHistory bool) {
	b.StopTimer()
	ctx := context.Background()

	dir, err := ioutil.TempDir("", "teststore")
	db := storage.OpenFile(dir)
	if err != nil {
		panic(err)
	}
	defer db.Close(ctx)
	var TestingEpochDuration time.Duration = 0 // time.Second

	cfg := &core.Config{
		VerificationPeriod: time.Duration(0),
		UpdatePeriod:       time.Duration(0),
		Partitions:         uint64(numPartitions),
		Verifier:           "local",
		AggHistory:         aggHistory,
		AggHistoryDepth:    21,
	}

	serv := legolog.NewServer(storage.NewMapStorage(), TestingEpochDuration, numPartitions, true, cfg)
	// auditor_serv, err := auditorsrv.NewAuditorWithManualConfig("localhost"+ServerPort, 0, 0, uint64(numPartitions), false)

	numAppendsPerVerificationPeriod := numAppends / numVerificationPeriodsOffline

	// precompute the ids we will insert, we will set values to be something random later
	ids := make([][]byte, numAppends)
	monitoredKeyPair := make([][]byte, numVerificationPeriodsOffline)
	signatures := make([][]byte, numAppends)
	monitoredSignatures := make([][]byte, numVerificationPeriodsOffline)
	for i := 0; i < numAppends; i++ {
		ids[i] = make([]byte, 32)
		// signatures[i] = make([]byte, 64)
		_, err := rand.Read(ids[i])
		if err != nil {
			b.Fatalf("error while generating random string: %s", err)
		}
	}

	for verificationPeriod := 0; verificationPeriod < numVerificationPeriodsOffline; verificationPeriod++ {
		monitoredKeyPair[verificationPeriod] = make([]byte, 32)
		_, err = rand.Read(monitoredKeyPair[verificationPeriod])
		if err != nil {
			b.Fatalf("error while generating random string: %s", err)
		}
	}

	masterSK, masterVK := crypto.GenerateKeypair()
	monitoringKey := []byte("monitoring_key")

	for verificationPeriod := 0; verificationPeriod < numVerificationPeriodsOffline; verificationPeriod++ {
		// do numAppendsPerVerificationPeriod appends
		for i := 0; i < numAppendsPerVerificationPeriod; i += 1 {
			// append from ids
			j := verificationPeriod*numAppendsPerVerificationPeriod + i
			ins := ids[j]
			SK, VK := crypto.GenerateKeypair()
			signature := make([]byte, 64)
			crypto.SignBlob(SK, VK, signature,
				append(ins, []byte("1")...))
			signatures[j] = signature
			partitionServer := serv.GetPartitionForIdentifier(ins)
			partitionServer.Partition.Append(ins, ins, ins, VK)
			partitionServer.LastPos += 1
		}
		ins := monitoredKeyPair[verificationPeriod]
		signature := make([]byte, 64)
		crypto.SignBlob(masterSK, masterVK, signature,
			append(ins, []byte("1")...))
		monitoredSignatures[verificationPeriod] = signature
		partitionServer := serv.GetPartitionForIdentifier(monitoringKey)
		partitionServer.Partition.Append(monitoringKey, monitoringKey, ins, masterVK)
		partitionServer.LastPos += 1
		serv.IncrementUpdateEpoch()
		serv.IncrementVerificationPeriod()
	}

	if aggHistory {
		identifier := monitoringKey
		value := monitoredKeyPair[len(monitoredKeyPair)-1]
		signature := monitoredSignatures[len(monitoredSignatures)-1]
		partitionServer := serv.GetPartitionForIdentifier(monitoringKey)
		digest := partitionServer.Partition.GetDigest()
		b.StartTimer()
		proof := partitionServer.Partition.GenerateExistenceProof(identifier, value, signature)
		// partitionServer.Partition.validateExistenceProof
		verifier := core.AggHistVerifier{}
		ok, err := verifier.ValidatePKProof(digest, proof, identifier, value, signature, 1, masterVK)
		if err != nil {
			fmt.Println("error while validating agg history proof: ", err)
		}
		if !ok {
			fmt.Println("unable to validate agg history proof")
		}
		b.StopTimer()
	} else {
		// loop through verification periods
		for verificationPeriod := 0; verificationPeriod < numVerificationPeriodsOffline; verificationPeriod++ {
			// generate existence proof for the last monitored key pair

		}
	}

}
*/

func helperMonitorMicrobenchDebug(b *testing.B, numPartitions, numAppends int, numVerificationPeriodsOffline int, aggHistory bool, numTrials int) float64 {
	b.StopTimer()
	ctx := context.Background()

	dir, err := ioutil.TempDir("", "teststore")
	db := storage.OpenFile(dir)
	if err != nil {
		panic(err)
	}
	defer db.Close(ctx)
	//var TestingEpochDuration time.Duration = 0 // time.Second
	var testCfg core.Config = core.Config{
		AggHistoryDepth: 31,
	}

	tmpdir, err := os.MkdirTemp("", "legologbench")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	var partition core.LegoLogPartition = nil
	if aggHistory {
		partition = core.NewAggHistPartition(testCfg, tmpdir)
	} else {
		partition = core.NewPartition()
	}
	masterSK, masterVK := crypto.GenerateKeypair()

	numAppendsPerVerificationPeriod := 100
	ids := make([][]byte, numAppendsPerVerificationPeriod+1)
	signatures := make([][]byte, numAppendsPerVerificationPeriod+1)
	values := make([][]byte, numAppendsPerVerificationPeriod+1)
	for i := 0; i < numAppendsPerVerificationPeriod+1; i++ {
		ids[i] = make([]byte, 32)
		_, masterVK2 := crypto.GenerateKeypair()
		signature := make([]byte, 64)
		crypto.SignBlob(masterSK, masterVK, signature, append(masterVK2, []byte("1")...))
		signatures[i] = signature
		values[i] = masterVK2

		/* 		signatures[i] = make([]byte, 64) */
		_, err := rand.Read(ids[i])
		if err != nil {
			b.Fatalf("error while generating random string: %s", err)
		}
	}

	monitoringKey := ids[numAppendsPerVerificationPeriod]
	monitoringSignature := signatures[numAppendsPerVerificationPeriod]
	monitoringValue := values[numAppendsPerVerificationPeriod]

	for i := 0; i < numAppendsPerVerificationPeriod; i++ {
		partition.Append(ids[i], ids[i], values[i], signatures[i])
	}
	partition.Append(monitoringKey, monitoringKey, monitoringValue, monitoringSignature)
	partition.IncrementUpdateEpoch()
	partition.IncrementVerificationPeriod()

	var lastMasterVK []byte = nil
	var lastSignature []byte = nil
	for verificationPeriod := 0; verificationPeriod < numVerificationPeriodsOffline; verificationPeriod++ {
		/* 		_, masterVK2 := crypto.GenerateKeypair()
		   		signature := make([]byte, 64)
		   		crypto.SignBlob(masterSK, masterVK, signature, append(masterVK2, []byte("1")...)) */

		for i := 0; i < numAppendsPerVerificationPeriod; i++ {
			partition.Append(ids[i], ids[i], values[i], signatures[i])
		}
		partition.IncrementUpdateEpoch()
		partition.IncrementVerificationPeriod()
	}

	digest := partition.GetDigest()
	totalTime := 0.0
	times := make([]float64, numTrials)
	for i := 0; i < numTrials; i++ {
		totalTimeThisTrial := 0.0
		if aggHistory {
			var v core.AggHistVerifier

			// b.StartTimer()
			//startTime := time.Now()
			proof := partition.GenerateExistenceProof(monitoringKey, monitoringValue, monitoringSignature)
			// only count validation time
			//totalTime += float64(time.Since(startTime))
			startTime := time.Now()
			ok, err := v.ValidatePKProofMonitoring(digest, proof, monitoringKey, monitoringValue, monitoringSignature, 1, masterVK)
			totalTimeThisTrial += float64(time.Since(startTime).Microseconds())
			// b.StopTimer()
			if err != nil {
				b.Error(err)
			}
			if !ok {
				b.Error("Proof validation failed")
			}

			/*
				ok, _ = v.ValidatePKProof(digest, proof, ids[0], []byte("foo"), lastSignature, 1, masterVK)
				if ok {
					b.Error("Proof should not have validated ")
				} */
		} else {
			// loop through verification periods
			for verificationPeriod := 0; verificationPeriod < numVerificationPeriodsOffline; verificationPeriod++ {
				b.StartTimer()
				startTime := time.Now()
				proof := partition.GenerateExistenceProof(ids[0], lastMasterVK, lastSignature)
				core.ValidatePKProof(digest, proof, ids[1], lastMasterVK, lastSignature, 1, masterVK)
				totalTimeThisTrial += float64(time.Since(startTime).Microseconds())
				b.StopTimer()
			}
		}
		totalTime += totalTimeThisTrial
		times[i] = totalTimeThisTrial
	}
	sort.Float64s(times)
	// for i := 0; i < numTrials; i++ {
	// 	fmt.Printf("%.2f,", times[i])
	// }
	// fmt.Println()
	return totalTime / float64(numTrials)
	// fmt.Println("TOTAL TIME: ", totalTime)
}

func helperMonitorMicrobenchDebugOptimized(b *testing.B, numPartitions int, numAppends int, numVerificationPeriodsOfflineEnd int, numVerificationPeriodsOfflineInterval int, aggHistory bool) {
	b.StopTimer()
	ctx := context.Background()

	dir, err := ioutil.TempDir("", "teststore")
	db := storage.OpenFile(dir)
	if err != nil {
		panic(err)
	}
	defer db.Close(ctx)
	//var TestingEpochDuration time.Duration = 0 // time.Second
	var testCfg core.Config = core.Config{
		AggHistoryDepth: 31,
	}

	tmpdir, err := os.MkdirTemp("", "legologbench")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	var partition core.LegoLogPartition = nil
	if aggHistory {
		partition = core.NewAggHistPartition(testCfg, tmpdir)
	} else {
		partition = core.NewPartition()
	}
	masterSK, masterVK := crypto.GenerateKeypair()

	preloadAppends := 32
	ids := make([][]byte, preloadAppends)
	for i := 0; i < preloadAppends; i++ {
		ids[i] = make([]byte, 32)
		/* 		signatures[i] = make([]byte, 64) */
		_, err := rand.Read(ids[i])
		if err != nil {
			b.Fatalf("error while generating random string: %s", err)
		}
	}

	for i := 0; i < preloadAppends; i++ {
		partition.Append(ids[i], ids[i], masterVK, masterVK)
	}
	partition.IncrementUpdateEpoch()
	partition.IncrementVerificationPeriod()
	partition.IncrementVerificationPeriod()

	var lastMasterVK []byte = nil
	var lastSignature []byte = nil
	for verificationPeriod := 0; verificationPeriod < numVerificationPeriodsOfflineEnd; verificationPeriod++ {
		_, masterVK2 := crypto.GenerateKeypair()
		signature := make([]byte, 64)
		crypto.SignBlob(masterSK, masterVK, signature, append(masterVK2, []byte("1")...))

		for i := 0; i < 32; i++ {
			partition.Append(ids[i], ids[i], masterVK2, signature)
		}
		partition.IncrementUpdateEpoch()
		partition.IncrementVerificationPeriod()
		lastMasterVK = masterVK2
		lastSignature = signature

		if verificationPeriod%numVerificationPeriodsOfflineInterval == 0 {
			digest := partition.GetDigest()
			totalTime := 0.0
			samples := 5
			if aggHistory {
				// loop throug hsampels
				for sample := 0; sample < samples; sample++ {
					b.StartTimer()
					startTime := time.Now()
					proof := partition.GenerateExistenceProof(ids[sample], lastMasterVK, lastSignature)

					var v core.AggHistVerifier
					ok, err := v.ValidatePKProof(digest, proof, ids[sample], lastMasterVK, lastSignature, 1, masterVK)
					if err != nil {
						b.Error(err)
					}
					if !ok {
						b.Error("Proof validation failed")
					}
					totalTime += float64(time.Since(startTime))
					b.StopTimer()
				}
				totalTime /= float64(samples)
				fmt.Printf("agghistory=%t,numVerificationPeriodsOffline=%d, ns=%f\n", aggHistory, verificationPeriod+1, totalTime)
				/*
					ok, _ = v.ValidatePKProof(digest, proof, ids[0], []byte("foo"), lastSignature, 1, masterVK)
					if ok {
						b.Error("Proof should not have validated ")
					} */
			} else {
				// loop through verification periods
				for j := 0; j < verificationPeriod; j++ {
					b.StartTimer()
					startTime := time.Now()
					proof := partition.GenerateExistenceProof(ids[0], lastMasterVK, lastSignature)
					core.ValidatePKProof(digest, proof, ids[1], lastMasterVK, lastSignature, 1, masterVK)
					totalTime += float64(time.Since(startTime))
					b.StopTimer()
				}
			}
		}

	}
	// return totalTime
	// fmt.Println("TOTAL TIME: ", totalTime)
}

func BenchmarkMonitorVaryingVPs(b *testing.B) {
	// partitionValues := []int{1, 2, 5, 10, 20, 50, 100, 200, 400, 800, 1000}

	numAppends := 1024
	numPartitions := 4
	// monitorBenchmarkWrapper(b, numAppends, numPartitions, 3, true)
	// iterate through verification periods offline from 1 to 10
	aggHistory := false

	//vpsOfflineAmounts := []int{1, 2, 5, 13, 29, 61, 125, 253, 509, 1021, 2045}
	vpsOfflineAmounts := []int{1, 2, 5, 13, 29, 61, 125, 253, 509}
	for _, vpsOffline := range vpsOfflineAmounts {
		time := monitorBenchmarkWrapper(b, numAppends, numPartitions, vpsOffline, aggHistory)
		//fmt.Printf("agghistory=0,numVerificationPeriodsOffline=%d, ns=%f\n", vpsOffline, time/1e3)
		fmt.Printf("0,%d,%f\n", vpsOffline, time/1e3)
	}

	/*
		for numVerificationPeriodsOffline := 1; numVerificationPeriodsOffline < 1000; {
			// b.Run(fmt.Sprintf("agghistory=0,numVerificationPeriodsOffline=%d", numVerificationPeriodsOffline), func(b *testing.B) {
				monitorBenchmarkWrapper(b, numAppends, numPartitions, numVerificationPeriodsOffline, aggHistory)
			})
			time := monitorBenchmarkWrapper(b, numAppends, numPartitions, numVerificationPeriodsOffline, aggHistory)
			fmt.Printf("agghistory=0,numVerificationPeriodsOffline=%d, ns=%f\n", numVerificationPeriodsOffline, time)
			if numVerificationPeriodsOffline <= 200 {
				numVerificationPeriodsOffline += 5
			} else {
				numVerificationPeriodsOffline += 50
			}
		}
	*/
	aggHistory = true
	// helperMonitorMicrobenchDebugOptimized(b, numPartitions, numAppends, 100, 5, aggHistory)
	aggHistory = true
	// verificationPeriods := []int{50, 100}

	vpsOfflineAmounts = []int{1, 2, 5, 13, 29, 61, 125, 253, 509, 1021, 2045} // , 4093, 8191, 16381}
	// vpsOfflineAmounts = []int{1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048}

	for _, vpsOffline := range vpsOfflineAmounts {
		time := monitorBenchmarkWrapper(b, numAppends, numPartitions, vpsOffline, aggHistory)
		//fmt.Printf("agghistory=1,numVerificationPeriodsOffline=%d, us=%f\n", vpsOffline, time/1e3)
		fmt.Printf("1,%d,%f\n", vpsOffline, time/1e3)
	}
	/*
		for numVerificationPeriodsOffline := 1; numVerificationPeriodsOffline < 1000; {
			// for _, numVerificationPeriodsOffline := range verificationPeriods {
			// b.Run(fmt.Sprintf("agghistory=1,numVerificationPeriodsOffline=%d", numVerificationPeriodsOffline), func(b *testing.B) {monitorBenchmarkWrapper(b, numAppends, numPartitions, numVerificationPeriodsOffline, aggHistory)})
			time := monitorBenchmarkWrapper(b, numAppends, numPartitions, numVerificationPeriodsOffline, aggHistory)
			fmt.Printf("agghistory=1,numVerificationPeriodsOffline=%d, ns=%f\n", numVerificationPeriodsOffline, time)
			numVerificationPeriodsOffline += 1
			// numVerificationPeriodsOffline *= 2
			// numVerificationPeriodsOffline -= 1
		}
	*/
	/* for updatePeriods := 64; updatePeriods < 10000; updatePeriods *= 2 {
		for _, partitions := range partitionValues {
			b.Run(fmt.Sprintf("Partition=%d,UpdatePeriods=%d", partitions, updatePeriods), func(b *testing.B) {
				monitorBenchmarkWrapper(b, numAppends, partitions, UpdatePeriod*float64(updatePeriods), UpdatePeriod)
			})
			if partitions == 1 {
				partitions -= 1
			}
		}
	} */
}
