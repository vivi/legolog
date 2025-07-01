// Package auditorsrv contains server side implementations for auditor API.
package auditorsrv

import (
	legolog "MerkleSquare/legolog/client"
	"context"
	"encoding/json"
	"fmt"
	"time"

	server "MerkleSquare/legolog/server"

	"github.com/huyuncong/MerkleSquare/core"
	legolog_grpcint "github.com/huyuncong/MerkleSquare/legolog/legolog-grpcint"
)

type Auditor struct {
	client legolog.BasicClient

	epochs                  uint64
	UpdateCheckpoints       []*legolog_grpcint.CheckPoint
	UpdateDigests           []*core.LegologDigest
	VerificationCheckpoints []*legolog_grpcint.CheckPoint

	config  core.Config
	stopper chan struct{}
}

func NewAuditorWithManualConfig(serverAddr string, verificationPeriod time.Duration, updatePeriod time.Duration, partitions uint64, createClient bool) (
	*Auditor, error) {
	config :=
		core.Config{
			VerificationPeriod: verificationPeriod,
			UpdatePeriod:       updatePeriod,
			Partitions:         partitions,
			Verifier:           "verifier",
			AggHistory:         false,
			AggHistoryDepth:    0,
		}

	Auditor := &Auditor{
		config:  config,
		stopper: make(chan struct{}),
	}
	var err error = nil
	if createClient {
		Auditor.client, err = legolog.NewLegologClient(serverAddr)
		if err != nil {
			return nil, err
		}
	}

	Auditor.initializeCheckpoints(config.Partitions)
	if updatePeriod != 0 {
		go Auditor.QueryLoop()
	}

	return Auditor, nil
}

func (a *Auditor) initializeCheckpoints(numPartitions uint64) {
	a.UpdateCheckpoints = make([]*legolog_grpcint.CheckPoint, numPartitions)
	a.VerificationCheckpoints = make([]*legolog_grpcint.CheckPoint, numPartitions)
	a.UpdateDigests = make([]*core.LegologDigest, numPartitions)

	for i := uint64(0); i < numPartitions; i++ {
		a.UpdateCheckpoints[i] = nil
		a.VerificationCheckpoints[i] = nil
		a.UpdateDigests[i] = nil
	}

}

func NewAuditor(serverAddr string, configFile string) (
	*Auditor, error) {
	config, err := core.ParseConfig(configFile)
	if err != nil {
		return nil, err
	}
	Auditor := &Auditor{
		config:  config,
		stopper: make(chan struct{}),
	}
	Auditor.client, err = legolog.NewLegologClient(serverAddr)
	if err != nil {
		return nil, err
	}

	// initialize the update checkpoints array with partition checkpoints
	Auditor.initializeCheckpoints(config.Partitions)
	go Auditor.QueryLoop()

	return Auditor, nil
}

func (a *Auditor) QueryLoop() {
	updateTimer := time.NewTimer(a.config.UpdatePeriod)
	defer updateTimer.Stop()
	var updateTicker *time.Ticker
	untilNextUpdateEpoch := updateTimer.C

	verificationTimer := time.NewTimer(a.config.VerificationPeriod)
	defer verificationTimer.Stop()
	var verificationTicker *time.Ticker
	untilNextVerificationPeriod := verificationTimer.C

queryLoop:
	for {
		select {
		case <-untilNextUpdateEpoch:
			//NOTE: Debug
			fmt.Println("Querying Auditor for Update Period", time.Now())
			if updateTicker == nil {
				fmt.Println("Starting update Ticker")
				updateTimer.Stop()
				updateTicker = time.NewTicker(a.config.UpdatePeriod)
				untilNextUpdateEpoch = updateTicker.C
			}
			a.QueryServerUpdatePeriod()

		case <-untilNextVerificationPeriod:
			fmt.Println("Querying Auditor for Verification Period", time.Now())
			if verificationTicker == nil {
				fmt.Println("Starting verification Ticker")
				verificationTimer.Stop()
				verificationTicker = time.NewTicker(a.config.VerificationPeriod)
				untilNextVerificationPeriod = verificationTicker.C
			}
			a.QueryServerVerificationPeriod()

		case <-a.stopper:
			fmt.Println("Stopping query loop!")
			break queryLoop
		}
	}
	if updateTicker != nil {
		updateTicker.Stop()
	}
	if verificationTicker != nil {
		verificationTicker.Stop()
	}
}

func (a *Auditor) QueryServerUpdatePeriod() {
	responses, err := a.queryServerUpdatePeriod()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	digests, proofs := a.getContentsFromCheckpointResponses(responses)

	proven := a.isConsistencyProofValid(digests, proofs)
	if !proven {
		var prevEpoch uint64 = 0
		if a.UpdateDigests[0] != nil {
			prevEpoch = a.UpdateDigests[0].Epoch
		}
		fmt.Printf("Could not prove epoch %v is an extension of epoch %v\n",
			responses[0].Checkpoint.GetEpoch(), prevEpoch)
	}

	var updateCheckpoints []*legolog_grpcint.CheckPoint
	for _, r := range responses {
		updateCheckpoints = append(updateCheckpoints, r.Checkpoint)
	}
	a.UpdateCheckpoints = updateCheckpoints
	a.UpdateDigests = digests
}

func (a *Auditor) QueryServerVerificationPeriod() {
	response, err := a.queryServerUpdatePeriod()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// TODO: EXTENSION PROOF FOR VERIFICATION PERIOD HISTORY FORESTS
	if a.config.AggHistory {

	}

	var verificationCheckpoints []*legolog_grpcint.CheckPoint
	for _, r := range response {
		verificationCheckpoints = append(verificationCheckpoints, r.Checkpoint)
	}
	a.VerificationCheckpoints = verificationCheckpoints
}

/* TODO: Implement if needed in eval
// QueryServerForSize queries server for new checkpoint and an extension proof,
// verifies the proof and returns the size of the response returned by the server.
// proof. This function is used for tests measuring size of the message.
func (a *Auditor) QueryServerForSize() int {
	response, err := a.queryServer()
	if err != nil {
		fmt.Printf("%v\n", err)
		return 0
	}
	proven := a.isExtensionProofValid(response)
	if !proven {
		fmt.Printf("Could not prove epoch %v is an extension of epoch %v\n",
			response.Checkpoint.GetEpoch(), a.Checkpoint.GetEpoch())
	}

	return proto.Size(response)
}
*/

func (a *Auditor) queryServerUpdatePeriod() ([]*legolog_grpcint.GetNewCheckPointResponse, error) {
	var res []*legolog_grpcint.GetNewCheckPointResponse
	for i := uint64(0); i < a.config.Partitions; i++ {
		var oldSize uint64 = 0
		if a.UpdateDigests[i] != nil {
			// used to use a.UpdateCheckpoints[i].NumLeaves(), but I'm porting everything away from UpdateCheckpoints for ease of use
			oldSize = uint64(a.UpdateDigests[i].BaseTreeSize)
		}
		checkpointRequest := &legolog_grpcint.GetNewCheckPointRequest{
			OldSize:        oldSize,
			PartitionIndex: i,
		}
		response, err := a.client.GetNewUpdateCheckPoint(
			context.Background(), checkpointRequest)
		if err != nil {
			return nil, err
		}
		res = append(res, response)
	}
	return res, nil
}

/*
hacky function used for benchmarking that accomplishes the functionality of `QueryServerUpdatePeriod`
without using a client (directly uses partitionServers instead, passed in from the test file)
*/
func (a *Auditor) QueryServerUpdatePeriodWithoutClient(partitionServers []*server.PartitionServer) {
	// initialize digest and proof array
	var digests []*core.LegologDigest
	var proofs []*core.MerkleExtensionProof
	for i := uint64(0); i < a.config.Partitions; i++ {
		var oldSize uint64 = 0
		if a.UpdateDigests[i] != nil {
			// used to use a.UpdateCheckpoints[i].NumLeaves(), but I'm porting everything away from UpdateCheckpoints for ease of use
			oldSize = uint64(a.UpdateDigests[i].UpdateLogSize)
		}
		/*
			replicate GetNewUpdateCheckPoint
		*/
		digest := partitionServers[i].PublishedDigest
		// proof := new(core.MerkleExtensionProof) // TODO: copy logic for getting the proof
		proof := partitionServers[i].Partition.GetUpdateEpochConsistencyProof(uint32(oldSize))
		_ = oldSize // old size would be needed to generate this proof, but for now its not i guess?
		digests = append(digests, &digest)
		proofs = append(proofs, proof)
	}

	proven := a.isConsistencyProofValid(digests, proofs)
	// fmt.Printf("%t\n", proven)
	var prevEpoch uint64 = 0
	if a.UpdateDigests[0] != nil {
		prevEpoch = a.UpdateDigests[0].Epoch
	}
	_ = prevEpoch
	if !proven {
		fmt.Printf("Could not prove epoch %v is an extension of epoch %v\n", digests[0].Epoch, prevEpoch)
	} else {
		// fmt.Printf("Proved epoch %v is an extension of epoch %v\n", digests[0].Epoch, prevEpoch)
	}
	/* 	a.isConsistencyProofValid(digests, proofs) */
	a.UpdateDigests = digests
}

func (a *Auditor) queryServerVerificationPeriod() ([]*legolog_grpcint.GetNewCheckPointResponse, error) {
	var res []*legolog_grpcint.GetNewCheckPointResponse
	for i := uint64(0); i < a.config.Partitions; i++ {
		checkpointRequest := &legolog_grpcint.GetNewCheckPointRequest{
			OldSize:        a.VerificationCheckpoints[i].GetNumLeaves(),
			PartitionIndex: i,
		}
		response, err := a.client.GetNewVerifyCheckPoint(
			context.Background(), checkpointRequest)
		if err != nil {
			return nil, err
		}
		res = append(res, response)
	}
	return res, nil
}

func (a *Auditor) isConsistencyProofValid(digests []*core.LegologDigest, proofs []*core.MerkleExtensionProof) bool {
	allVerified := true

	for i, newDigest := range digests {
		oldDigest := a.UpdateDigests[i]

		extensionProof := proofs[i]

		var oldRoots [][]byte = nil
		var oldSize uint32 = 0
		if oldDigest != nil {
			oldRoots = [][]byte{oldDigest.UpdateLogRoot}
			oldSize = oldDigest.UpdateLogSize
		}
		oldDigestForProof := &core.Digest{
			Roots: oldRoots,
			Size:  oldSize, // TODO: correct this size
		}
		newDigestForProof := &core.Digest{
			Roots: [][]byte{newDigest.UpdateLogRoot},
			Size:  newDigest.UpdateLogSize, // TODO: correct this size
		}

		allVerified = allVerified && core.VerifyConsistencyProof(oldDigestForProof, newDigestForProof, extensionProof)
	}
	return allVerified
}

/*
extract the proofs and digests from checkpoint responses into two arrays. unmarshall the digests and proofs before
inserting them into the array. return the two arrays.
*/
func (a *Auditor) getContentsFromCheckpointResponses(responses []*legolog_grpcint.GetNewCheckPointResponse) ([]*core.LegologDigest, []*core.MerkleExtensionProof) {
	var digests []*core.LegologDigest
	var proofs []*core.MerkleExtensionProof
	for _, response := range responses {
		marshaledNewDigest := response.Checkpoint.MarshaledDigest
		var newDigest = new(core.LegologDigest)
		json.Unmarshal(marshaledNewDigest, &newDigest)
		digests = append(digests, newDigest)

		marshaledProof := response.Proof
		var extensionProof = new(core.MerkleExtensionProof)
		json.Unmarshal(marshaledProof, &extensionProof)
		proofs = append(proofs, extensionProof)
	}
	return digests, proofs
}

/*
func (a *Auditor) isConsistencyProofValid(
	responses []*legolog_grpcint.GetNewCheckPointResponse) bool {
	var oldDigest = new(core.Digest)
	var newDigest = new(core.Digest)
	var extensionProof = new(core.MerkleExtensionProof)
	allVerified := true

	for i, response := range responses {
		marshaledOldDigest := a.UpdateCheckpoints[i].GetMarshaledDigest()
		json.Unmarshal(marshaledOldDigest, &oldDigest)

		marshaledNewDigest := response.Checkpoint.GetMarshaledDigest()
		json.Unmarshal(marshaledNewDigest, &newDigest)

		marshaledProof := response.Proof
		json.Unmarshal(marshaledProof, &extensionProof)
		allVerified = allVerified && core.VerifyConsistencyProof(oldDigest, newDigest, extensionProof)
	}
	return allVerified
}
*/

// Stop ends the epoch loop. This is useful if you need to free all resources
// associated with a Auditor.
func (a *Auditor) Stop() {
	close(a.stopper)
}
