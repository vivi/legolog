package bench

import (
	// "MerkleSquare/constants"
	// "MerkleSquare/legolog/auditor/auditorsrv"
	legolog "MerkleSquare/legolog/server"
	"bytes"
	"encoding/csv"
	"encoding/gob"
	"sort"

	// "encoding/json"
	"io"

	// "math"
	"path/filepath"

	"github.com/huyuncong/MerkleSquare/core"

	"crypto/rand"
	// "errors"
	"fmt"
	math_rand "math/rand"
	"time"

	// client "github.com/huyuncong/MerkleSquare/legolog/client"

	"os"

	"github.com/huyuncong/MerkleSquare/lib/storage"
	"github.com/immesys/bw2/crypto"
)

type CounterWr struct {
	io.Writer
	Count int
}

func (cw *CounterWr) Write(p []byte) (n int, err error) {
	n, err = cw.Writer.Write(p)
	cw.Count += n
	return
}

func realHelperLookupsMicrobench(incrementVerificationPeriod bool, numAppendsPerUpdatePeriod int, numLookupsPerUpdatePeriod int, maxUpdates int, updatePeriod float64, dataInterval int, numPartitions int, aggHistory bool, lookupWriter *csv.Writer, proofSizeWriter *csv.Writer, proofTimesWriter *csv.Writer) {
	math_rand.Seed(1) // TODO: FIXME
	tmpdir, err := os.MkdirTemp("", "legologbench")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)
	var TestingEpochDuration time.Duration = 0 // time.Second
	TestingNumPartitions := numPartitions

	config := &core.Config{
		VerificationPeriod: time.Duration(float64(maxUpdates) * float64(time.Second)),
		UpdatePeriod:       time.Duration(updatePeriod * float64(time.Second)),
		Partitions:         uint64(TestingNumPartitions),
		Verifier:           "local",
		AggHistory:         aggHistory,
		AggHistoryDepth:    uint32(AggHistoryDepthNum),
	}

	var serv *legolog.Server
	if !aggHistory {
		serv = legolog.NewServer(storage.NewMapStorage(), TestingEpochDuration, TestingNumPartitions, false, config, tmpdir)
	} else {
		serv = legolog.NewServer(storage.NewMapStorage(), TestingEpochDuration, TestingNumPartitions, true, config, tmpdir)
	}

	numAppends := numAppendsPerUpdatePeriod * maxUpdates

	// preload 1 million appends to the server
	// numPreloadAppends := 1000000
	partitionDistribution := make([]int, TestingNumPartitions)
	// numPreloadAppends := 100
	// for i := 0; i < numPreloadAppends; i++ {
	// 	id := make([]byte, 32)
	// 	_, err := rand.Read(id)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	SK, VK := crypto.GenerateKeypair()
	// 	signature := make([]byte, 64)
	// 	ins := id
	// 	crypto.SignBlob(SK, VK, signature,
	// 		append(ins, []byte("0")...))
	// 	partitionServer := serv.GetPartitionForIdentifier(ins)
	// 	partitionServer.Partition.Append(ins, ins, ins, signature)
	// 	partitionServer.LastPos += 1
	// 	partitionDistribution[partitionServer.Index] += 1
	// }
	// TODO: fixme uncomment preloads and print statement below

	// fmt.Printf("Finished loading %d appends\n", numPreloadAppends)

	for i, v := range partitionDistribution {
		fmt.Printf("Partition %d has %d appends\n", i, v)
	}

	// inc verification period twice to move preloaded data to query copy
	serv.IncrementVerificationPeriod()
	serv.IncrementVerificationPeriod()

	ids := make([][]byte, numAppends)
	signatures := make([][]byte, numAppends)
	SKs := make([][]byte, numAppends)
	VKs := make([][]byte, numAppends)
	for i := 0; i < numAppends; i++ {
		ids[i] = make([]byte, 32)
		// signatures[i] = make([]byte, 64)
		_, err := rand.Read(ids[i])
		if err != nil {
			panic(err)
		}
		SK, VK := crypto.GenerateKeypair()
		SKs[i] = SK
		VKs[i] = VK
	}

	numAppended := 0

	b := &bytes.Buffer{}
	enc := gob.NewEncoder(b)

	nextPow2 := 4
	_ = nextPow2
	partitionDistribution = make([]int, TestingNumPartitions)
	for epoch := 2; epoch < maxUpdates; epoch++ {
		fmt.Printf("Epoch %d\n", epoch)
		/*if epoch > 0 && epoch%dataInterval == 0 {
			// lookup time data
			avgLookupTime := totalLookupTime / float64(numLookedUp)
			dataToWrite := []string{fmt.Sprintf("%d", epoch), fmt.Sprintf("%f", avgLookupTime)}
			err := lookupWriter.Write(dataToWrite)
			if err != nil {
				panic(err)
			}
			lookupWriter.Flush()

			// proof size data
			avgProofSize := totalProofSize / float64(numProofSizes)
			dataToWrite = []string{fmt.Sprintf("%d", epoch), fmt.Sprintf("%f", avgProofSize)}
			err = proofSizeWriter.Write(dataToWrite)
			if err != nil {
				panic(err)
			}
			proofSizeWriter.Flush()

			// proof validation time data
			avgProofValidationTime := totalProofValidationTime / float64(numProofTimes)
			dataToWrite = []string{fmt.Sprintf("%d", epoch), fmt.Sprintf("%f", avgProofValidationTime)}
			err = proofTimesWriter.Write(dataToWrite)
			if err != nil {
				panic(err)
			}
			proofTimesWriter.Flush()
		}*/

		for i := 0; i < numAppendsPerUpdatePeriod; i += 1 {
			// append from ids
			signature := make([]byte, 64)
			ins := ids[numAppended]
			crypto.SignBlob(SKs[numAppended], VKs[numAppended], signature,
				append(ins, []byte("0")...))
			signatures[numAppended] = signature
			partitionServer := serv.GetPartitionForIdentifier(ins)
			partitionServer.Partition.Append(ins, ins, ins, signatures[numAppended])
			fmt.Printf("append: %#v\n", ins[:10])

			partitionServer.LastPos += 1
			numAppended += 1
			partitionDistribution[partitionServer.Index] += 1
		}

		serv.IncrementUpdateEpoch()
		if incrementVerificationPeriod {
			serv.IncrementVerificationPeriod()
		}

		if incrementVerificationPeriod && epoch != nextPow2 {
			continue
		} else {
			nextPow2 *= 2
		}

		if !incrementVerificationPeriod && epoch%25 != 0 { // TODO: CHANGE BACK TO MOD 25
			continue
		}

		//partitionDistributionLookups := make([]int, TestingNumPartitions)
		//for i, v := range partitionDistribution {
		//	fmt.Printf("[%d] Partition %d has %d appends\n", epoch, i, v)
		//}

		lookupTimesThisPeriod := make([]float64, numLookupsPerUpdatePeriod)
		lookupTimeThisPeriod := 0.0
		proofSizeThisPeriod := 0.0
		proofValidationTimesThisPeriod := make([]float64, numLookupsPerUpdatePeriod)
		proofValidationTimeThisPeriod := 0.0
		for j := 0; j < numLookupsPerUpdatePeriod; j++ {
			// random element from what has already been appended
			lookupIndex := math_rand.Intn(numAppended)
			if err != nil {
				panic(fmt.Errorf("error while generating random index: %s", err.Error()))
			}

			partitionServer := serv.GetPartitionForIdentifier(ids[lookupIndex])
			identifier, value, signature := ids[lookupIndex], ids[lookupIndex], signatures[lookupIndex]
			timeStart := time.Now()
			fmt.Printf("lookup: %#v\n", identifier[:10])
			proof := partitionServer.Partition.GenerateExistenceProof(identifier, value, signature)
			timeTaken := float64(time.Since(timeStart).Nanoseconds())
			lookupTimeThisPeriod += timeTaken
			lookupTimesThisPeriod[j] = timeTaken
			if proof == nil {
				panic("proof is nil")
			}

			enc.Encode(proof)
			thisProofSize := float64(b.Len())
			proofSizeThisPeriod += thisProofSize
			b.Reset()

			var v core.AggHistVerifier
			timeStart = time.Now()
			digest := partitionServer.Partition.GetDigest()
			// fmt.Printf("digest basetreeroot[0]: %x\n", digest.BaseTreeRoots[0])
			// fmt.Printf("digest basetreeroot[1]: %x\n", digest.BaseTreeRoots[1])
			ok, err := v.ValidatePKProof(digest, proof, identifier, value, signature, 0, VKs[lookupIndex])
			timeTaken = float64(time.Since(timeStart).Nanoseconds())
			proofValidationTimeThisPeriod += timeTaken
			proofValidationTimesThisPeriod[j] = timeTaken
			if err != nil {
				panic(err)
			}
			if !ok {
				panic("failed to validate proof")
			}

			//partitionDistributionLookups[partitionServer.Index] += 1
		}

		/*
			for i, v := range partitionDistributionLookups {
				fmt.Printf("Partition %d had %d lookups\n", i, v)
			}
		*/

		// lookup time data
		//avgLookupTime := lookupTimeThisPeriod / float64(numLookupsPerUpdatePeriod)
		sort.Float64s(lookupTimesThisPeriod)
		for _, v := range lookupTimesThisPeriod {
			fmt.Printf("%.2f,", v)
		}
		fmt.Printf("\n")
		medianLookUpTime := lookupTimesThisPeriod[numLookupsPerUpdatePeriod/2] / 1e6 // convert to ms
		//medianLookUpTime = lookupTimeThisPeriod / float64(numLookupsPerUpdatePeriod) / 1e6
		dataToWrite := []string{fmt.Sprintf("%d", epoch), fmt.Sprintf("%f", medianLookUpTime)}
		err := lookupWriter.Write(dataToWrite)
		if err != nil {
			panic(err)
		}
		lookupWriter.Flush()

		// proof size data
		avgProofSize := proofSizeThisPeriod / float64(numLookupsPerUpdatePeriod)
		dataToWrite = []string{fmt.Sprintf("%d", epoch), fmt.Sprintf("%f", avgProofSize)}
		err = proofSizeWriter.Write(dataToWrite)
		if err != nil {
			panic(err)
		}
		proofSizeWriter.Flush()

		// proof validation time data
		//avgProofValidationTime := proofValidationTimeThisPeriod / float64(numLookupsPerUpdatePeriod)
		sort.Float64s(proofValidationTimesThisPeriod)
		medianValidationTime := proofValidationTimesThisPeriod[numLookupsPerUpdatePeriod/2] / 1e6 // convert to ms
		dataToWrite = []string{fmt.Sprintf("%d", epoch), fmt.Sprintf("%f", medianValidationTime)}
		err = proofTimesWriter.Write(dataToWrite)
		if err != nil {
			panic(err)
		}
		proofTimesWriter.Flush()
	}
}

const numAppendsPerUpdatePeriod int = 4
const numLookupsPerUpdatePeriod int = 8
const UpdatePeriod float64 = 1.0

// const numAppendsPerUpdatePeriod int = 40
// const numLookupsPerUpdatePeriod int = 8
// const UpdatePeriod float64 = 1.0

// const maxVerificationTime int = 1025
const maxVerificationTime int = 400
const DataInterval int = 1

const AggHistoryDepthNum int = 31

func RealMeasureLookups(partitionList []int, AggHistoryOn bool) {
	var err error
	lookupsFolder := "lookup-time_bench_output"
	err = os.Mkdir(lookupsFolder, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	proofSizeFolder := "proof-size_bench_output"
	err = os.Mkdir(proofSizeFolder, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	proofValidationTimeFolder := "proof-time_bench_output"
	err = os.Mkdir(proofValidationTimeFolder, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Running with AggHistoryOn: ", AggHistoryOn)
	for _, numPartitions := range partitionList {
		fmt.Println("real_lookups_bench.go: Starting iteration with numPartitions: ", numPartitions)

		// file to write lookup times to
		lookupFileName := fmt.Sprintf("%d-partitions_agg-%t_lookup_times.csv", numPartitions, AggHistoryOn)
		path := filepath.Join(lookupsFolder, lookupFileName)
		newFilePath := filepath.FromSlash(path)
		lookupTimesFile, err := os.Create(newFilePath)
		if err != nil {
			fmt.Println(err)
		}
		defer lookupTimesFile.Close()

		// file to write proof sizes to
		proofSizeFilename := fmt.Sprintf("%d-partitions_agg-%t_proof_sizes.csv", numPartitions, AggHistoryOn)
		path = filepath.Join(proofSizeFolder, proofSizeFilename)
		newFilePath = filepath.FromSlash(path)
		proofSizesFile, err := os.Create(newFilePath)
		if err != nil {
			fmt.Println(err)
		}
		defer proofSizesFile.Close()

		// file to write proof validation times to
		proofValidationTimesFilename := fmt.Sprintf("%d-partitions_agg-%t_proof_validations.csv", numPartitions, AggHistoryOn)
		path = filepath.Join(proofValidationTimeFolder, proofValidationTimesFilename)
		newFilePath = filepath.FromSlash(path)
		proofValidationTimesFile, err := os.Create(newFilePath)
		if err != nil {
			fmt.Println(err)
		}
		defer proofValidationTimesFile.Close()

		writer1 := csv.NewWriter(lookupTimesFile)
		defer writer1.Flush()
		writer2 := csv.NewWriter(proofSizesFile)
		defer writer2.Flush()
		writer3 := csv.NewWriter(proofValidationTimesFile)
		defer writer3.Flush()

		writer1.Write([]string{"t_v/t_u", "lookupTime (ns)"})
		writer2.Write([]string{"t_v/t_u", "proofSize (bytes)"})
		writer3.Write([]string{"t_v/t_u", "proofValidationTime (µs)"})
		realHelperLookupsMicrobench(false, numAppendsPerUpdatePeriod, numLookupsPerUpdatePeriod, int(maxVerificationTime), UpdatePeriod, DataInterval, numPartitions, AggHistoryOn, writer1, writer2, writer3)
	}
}

func MeasureLookupsOverMultipleVerificationPeriods(partitionList []int, AggHistoryOn bool) {
	var err error
	lookupsFolder := "vp-lookup-time-vp_bench_output"
	err = os.Mkdir(lookupsFolder, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	proofSizeFolder := "vp-proof-size_bench_output"
	err = os.Mkdir(proofSizeFolder, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}
	proofValidationTimeFolder := "vp-proof-time_bench_output"
	err = os.Mkdir(proofValidationTimeFolder, os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Running with AggHistoryOn: ", AggHistoryOn)
	for _, numPartitions := range partitionList {
		fmt.Println("lookups_bench.go [vp]: Starting iteration with numPartitions: ", numPartitions)
		// file to write lookup times to
		lookupFileName := fmt.Sprintf("%d-partitions_agg-%t_lookup_times.csv", numPartitions, AggHistoryOn)
		path := filepath.Join(lookupsFolder, lookupFileName)
		newFilePath := filepath.FromSlash(path)
		lookupTimesFile, err := os.Create(newFilePath)
		if err != nil {
			fmt.Println(err)
		}
		defer lookupTimesFile.Close()

		// file to write proof sizes to
		proofSizeFilename := fmt.Sprintf("%d-partitions_agg-%t_proof_sizes.csv", numPartitions, AggHistoryOn)
		path = filepath.Join(proofSizeFolder, proofSizeFilename)
		newFilePath = filepath.FromSlash(path)
		proofSizesFile, err := os.Create(newFilePath)
		if err != nil {
			fmt.Println(err)
		}
		defer proofSizesFile.Close()

		// file to write proof validation times to
		proofValidationTimesFilename := fmt.Sprintf("%d-partitions_agg-%t_proof_validations.csv", numPartitions, AggHistoryOn)
		path = filepath.Join(proofValidationTimeFolder, proofValidationTimesFilename)
		newFilePath = filepath.FromSlash(path)
		proofValidationTimesFile, err := os.Create(newFilePath)
		if err != nil {
			fmt.Println(err)
		}
		defer proofValidationTimesFile.Close()

		writer1 := csv.NewWriter(lookupTimesFile)
		defer writer1.Flush()
		writer2 := csv.NewWriter(proofSizesFile)
		defer writer2.Flush()
		writer3 := csv.NewWriter(proofValidationTimesFile)
		defer writer3.Flush()

		writer1.Write([]string{"t_v/t_u", "lookupTime (ns)"})
		writer2.Write([]string{"t_v/t_u", "proofSize (bytes)"})
		writer3.Write([]string{"t_v/t_u", "proofValidationTime (µs)"})
		realHelperLookupsMicrobench(true, numAppendsPerUpdatePeriod, numLookupsPerUpdatePeriod*numPartitions, int(maxVerificationTime), UpdatePeriod, DataInterval, numPartitions, AggHistoryOn, writer1, writer2, writer3)
	}
}
