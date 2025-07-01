package bench

// new code
import (
	"MerkleSquare/constants"
	"MerkleSquare/legolog/auditor/auditorsrv"
	legolog "MerkleSquare/legolog/server"
	"context"
	"encoding/csv"
	"io/ioutil"
	"sort"

	// "math"
	"path/filepath"

	"crypto/rand"
	"fmt"

	// math_rand "math/rand"
	"os"
	"time"

	"github.com/huyuncong/MerkleSquare/core"

	"github.com/huyuncong/MerkleSquare/lib/storage"
	"github.com/immesys/bw2/crypto"
)

const ServerAddr = "localhost" + constants.ServerPort

const ServerPort = constants.ServerPort

func helperAuditorMicrobench(numAppendsPerUpdatePeriod int, numPartitions int, maxVerificationTime float64, updatePeriod float64, dataInterval int, writer *csv.Writer, numTrials int) {
	if updatePeriod != 1.0 {
		panic("update period must be 1.0")
	}
	if dataInterval != 1 {
		panic("data interval must be 1")
	}

	auditingTimes := make([]float64, int(maxVerificationTime))
	for i := range auditingTimes {
		auditingTimes[i] = 0
	}

	tmpdir, err := os.MkdirTemp("", "legologbench")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tmpdir)

	// each trial is a verification period, effectively
	for trial := 0; trial < numTrials; trial++ {
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
			AggHistory:         false,
			AggHistoryDepth:    21,
		}

		serv := legolog.NewServer(storage.NewMapStorage(), TestingEpochDuration, numPartitions, false, cfg, tmpdir)
		auditor_serv, err := auditorsrv.NewAuditorWithManualConfig("localhost"+ServerPort, 0, 0, uint64(numPartitions), false)

		numAppends := numAppendsPerUpdatePeriod * int(maxVerificationTime)

		// preload 1 million appends to the server
		//numPreloadAppends := 1000000
		numPreloadAppends := 10
		for i := 0; i < numPreloadAppends; i++ {
			id := make([]byte, 32)
			_, err := rand.Read(id)
			if err != nil {
				panic(err)
			}
			SK, VK := crypto.GenerateKeypair()
			signature := make([]byte, 64)
			ins := id
			crypto.SignBlob(SK, VK, signature,
				append(ins, []byte("0")...))
			partitionServer := serv.GetPartitionForIdentifier(ins)
			partitionServer.Partition.Append(ins, ins, ins, signature)
			partitionServer.LastPos += 1
		}
		// inc verification period twice to move preloaded data to query copy
		serv.IncrementVerificationPeriod()
		serv.IncrementVerificationPeriod()

		// precompute the ids we will insert, we will set values to be something random later
		ids := make([][]byte, numAppends)
		signatures := make([][]byte, numAppends)
		for i := 0; i < numAppends; i++ {
			ids[i] = make([]byte, 32)
			_, err := rand.Read(ids[i])
			if err != nil {
				panic(fmt.Errorf("error while generating random string: %s", err))
			}
		}
		totalTime := 0.0
		var thisTime float64

		for epoch := 0; epoch < int(maxVerificationTime); epoch++ {
			// if epoch > 0 && epoch%dataInterval == 0 {
			// 	// dataToWrite := []string{fmt.Sprintf("%d", epoch), fmt.Sprintf("%f", totalTime/float64(epoch))}
			// 	dataToWrite := []string{fmt.Sprintf("%d", epoch), fmt.Sprintf("%f", thisTime)}
			// 	err = writer.Write(dataToWrite)
			// 	if err != nil {
			// 		panic(err)
			// 	}
			// 	writer.Flush()
			// }

			for i := 0; i < numAppendsPerUpdatePeriod; i += 1 {
				// append from ids
				j := epoch*numAppendsPerUpdatePeriod + i
				ins := ids[j]
				SK, VK := crypto.GenerateKeypair()
				signature := make([]byte, 64)
				crypto.SignBlob(SK, VK, signature,
					append(ins, []byte("0")...))
				signatures[j] = signature
				partitionServer := serv.GetPartitionForIdentifier(ins)
				partitionServer.Partition.Append(ins, ins, ins, VK)
				partitionServer.LastPos += 1
				if err != nil {
					panic(fmt.Errorf("failed to append random key pair: %s", err.Error()))
				}
			}
			serv.IncrementUpdateEpoch()
			startTime := time.Now()
			auditor_serv.QueryServerUpdatePeriodWithoutClient(serv.PartitionServers)
			thisTime = float64(time.Since(startTime).Microseconds()) / 1e3
			totalTime += thisTime

			auditingTimes[epoch] += thisTime
		}
	}

	sort.Float64s(auditingTimes)
	dataToWrite := []string{fmt.Sprintf("%d", numPartitions), fmt.Sprintf("%f", auditingTimes[len(auditingTimes)/2]/float64(numTrials))}
	err = writer.Write(dataToWrite)
	if err != nil {
		panic(err)
	}
	writer.Flush()
	return
	for i := range auditingTimes {
		// dataToWrite := []string{fmt.Sprintf("%d", epoch), fmt.Sprintf("%f", totalTime/float64(epoch))}
		dataToWrite := []string{fmt.Sprintf("%d", i), fmt.Sprintf("%f", auditingTimes[i]/float64(numTrials))}
		err := writer.Write(dataToWrite)
		if err != nil {
			panic(err)
		}
		writer.Flush()
	}
}

func BenchmarkAuditorVaryingPartitions(partitionValues []int) {

	// os.RemoveAll("audit_bench_output")
	err := os.Mkdir("audit_bench_output", os.ModePerm)
	if err != nil {
		fmt.Println(err)
	}

	numAppendsPerUpdatePeriod := 200
	maxVerificationTime := 100.0
	dataInterval := 1
	numTrials := 5
	// maxVerificationTime := 20000.0

	fileName := "audit_times.csv"
	path := filepath.Join("audit_bench_output", fileName)
	newFilePath := filepath.FromSlash(path)
	file, err := os.Create(newFilePath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write([]string{"t_v/t_u", "time_taken (milliseconds)"})
	if err != nil {
		panic(err)
	}
	for _, partitions := range partitionValues {
		fmt.Println("Starting iteration with numPartitions: ", partitions)
		helperAuditorMicrobench(numAppendsPerUpdatePeriod, partitions, maxVerificationTime, 1.0, dataInterval, writer, numTrials)
		writer.Flush()
	}

	return
	for _, partitions := range partitionValues {
		fmt.Println("Starting iteration with numPartitions: ", partitions)

		fileName := fmt.Sprintf("%d-partitions_audit_times.csv", partitions)
		path := filepath.Join("audit_bench_output", fileName)
		newFilePath := filepath.FromSlash(path)
		file, err := os.Create(newFilePath)
		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		defer writer.Flush()

		err = writer.Write([]string{"t_v/t_u", "time_taken (microseconds)"})
		if err != nil {
			panic(err)
		}
		helperAuditorMicrobench(numAppendsPerUpdatePeriod, partitions, maxVerificationTime, 1.0, dataInterval, writer, numTrials)
		writer.Flush()
	}
}

// /*
// First run the server in a separate terminal (todo: write a script to run which will create this server).
// Then run this benchmarking code (currently the server runs at the default address so clients will know where
// to connect to without any additional arguments).
// */
// /*
// func helperLookups(b *testing.B, numPairs uint32, numLookups uint32, verificationTime float64) {
// 	b.StopTimer()
// 	ctx := context.Background()

// 	c, err := client.NewClient(ServerAddr, "", "")
// 	if err != nil {
// 		b.Error(errors.New("Failed to start client: " + err.Error()))
// 	}

// 	start := time.Now()

// 	appendsPerUpdatePeriod := 10
// 	numUpdatePeriods := 20
// 	numPairs = uint32(numUpdatePeriods * appendsPerUpdatePeriod)
// 	updateEpoch := time.Second

// 	// precompute the ids we will insert, we will set values to be the same
// 	ids := make([][]byte, numPairs)
// 	for i := uint32(0); i < numPairs; i++ {
// 		ids[i] = make([]byte, 32)
// 		_, err := rand.Read(ids[i])
// 		if err != nil {
// 			b.Fatalf("error while generating random string: %s", err)
// 		}
// 	}

// 	for epoch := 0; epoch < numUpdatePeriods; epoch++ {
// 		for i := 0; i < appendsPerUpdatePeriod; i += 1 {
// 			// append from ids
// 			for _, ins := range ids {
// 				_, VK := crypto.GenerateKeypair()
// 				pos, _, err := c.Append(ctx, ins, ins, VK)
// 				if err != nil {
// 					b.Error(fmt.Errorf("Failed to append random key pair %s at pos %d", err.Error(), pos))
// 				}
// 			}
// 		}

// 		// if < updateEpoch, sleep until updateEpoch
// 		elapsed := time.Since(start)
// 		if elapsed < updateEpoch {
// 			time.Sleep(updateEpoch - elapsed)
// 		}

// 		// benchmark lookup
// 		// randomly generated index between 0 and numPairs
// 		lookupIndex := math_rand.Intn(int(numPairs))

// 		if err != nil {
// 			b.Error(fmt.Errorf("Error while generating random index: %s", err.Error()))
// 		}
// 		b.StartTimer()
// 		c.LookUpPKVerify(ctx, ids[lookupIndex], ids[lookupIndex])
// 		b.StopTimer()
// 	}

// }
// */
// // everything below is from the old lookups benchmarking stuff

// /*
// First run the server in a separate terminal (todo: write a script to run which will create this server).
// Then run this benchmarking code (currently the server runs at the default address so clients will know where
// to connect to without any additional arguments).
// */

// func helperLookupsMicrobenchOld(b *testing.B, numAppends int, numLookups int, verificationTime float64, updatePeriod float64) {
// 	b.StopTimer()
// 	ctx := context.Background()

// 	dir, err := ioutil.TempDir("", "teststore")
// 	db := storage.OpenFile(dir)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer db.Close(ctx)
// 	var TestingEpochDuration time.Duration = 0 // time.Second
// 	TestingNumPartitions := 4

// 	serv := legolog.NewServer(storage.NewMapStorage(), TestingEpochDuration, TestingNumPartitions, false, nil)

// 	numUpdatePeriods := int(math.Ceil(verificationTime / updatePeriod))
// 	numAppendsPerUpdatePeriod := numAppends / numUpdatePeriods
// 	numLookupsPerUpdatePeriod := numLookups / numUpdatePeriods

// 	// precompute the ids we will insert, we will set values to be something random later
// 	ids := make([][]byte, numAppends)
// 	signatures := make([][]byte, numAppends)
// 	for i := 0; i < numAppends; i++ {
// 		ids[i] = make([]byte, 32)
// 		// signatures[i] = make([]byte, 64)
// 		_, err := rand.Read(ids[i])
// 		if err != nil {
// 			b.Fatalf("error while generating random string: %s", err)
// 		}
// 	}

// 	for epoch := 0; epoch < numUpdatePeriods; epoch++ {
// 		for i := 0; i < numAppendsPerUpdatePeriod; i += 1 {
// 			// append from ids
// 			for j, ins := range ids {
// 				SK, VK := crypto.GenerateKeypair()
// 				signature := make([]byte, 64)
// 				crypto.SignBlob(SK, VK, signature,
// 					append(ins, 0))
// 				signatures[j] = signature
// 				serv.GetPartitionForIdentifier(ins).Partition.Append(ins, ins, ins, VK)
// 				// pos, _, err := c.Append(ctx, ins, ins, VK)
// 				if err != nil {
// 					b.Error(fmt.Errorf("Failed to append random key pair: %s", err.Error()))
// 				}
// 			}
// 		}
// 		for j := 0; j < numLookupsPerUpdatePeriod; j++ {
// 			// benchmark lookup
// 			// randomly generated index between 0 and numPairs
// 			lookupIndex := math_rand.Intn(int(numAppends))

// 			if err != nil {
// 				b.Error(fmt.Errorf("Error while generating random index: %s", err.Error()))
// 			}
// 			partitionServer := serv.GetPartitionForIdentifier(ids[lookupIndex])
// 			identifier, value, signature := ids[lookupIndex], ids[lookupIndex], signatures[lookupIndex]
// 			b.StartTimer()
// 			partitionServer.Partition.GenerateExistenceProof(identifier, value, signature)
// 			b.StopTimer()
// 		}
// 		serv.IncrementUpdateEpoch()
// 	}
// }

// /*
// 	func BenchmarkLookupsSmall1(b *testing.B) {
// 		helperLookupsMicrobench(b, NumAppends, NumLookups, 1, UpdatePeriod)
// 	}

// 	func BenchmarkLookupsSmall2(b *testing.B) {
// 		helperLookupsMicrobench(b, NumAppends, NumLookups, 10, UpdatePeriod)
// 	}
// 	func BenchmarkLookupsSmall3(b *testing.B) {
// 	helperLookupsMicrobenchOld(b, NumAppends, NumLookups, 100, UpdatePeriod)
// }
// */

// // func BenchmarkLookupVsVerificationPeriod(b *testing.B) {
// // 	/* 	updateEpoch := 500
// // 	   	for k := 1; k < 10; k++ {
// // 	   		verificationPeriod := updateEpoch * k

// // 	   	} */

// // }

// // graph to generate is verification period vs lookup time
// // iterate over # of partitions p = {1, 2, 4, etc.} [ask vivian]
// // we can just set update epoch = c, verification period = c * k for k integer
// // do fixed number of appends per update period
// // do lookups at even intervals within verification period, for a few verification periods
// // q: what do we look up?
