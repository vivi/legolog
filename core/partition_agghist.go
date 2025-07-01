package core

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type AggHistPartition struct {
	cfg Config

	baseTreeForest *HistoryForest
	baseTree       *persistentPrefixTree

	queryUpdateLog         *ChronTree
	queryUpdatePrefixTrees []*prefixTree

	verifyUpdateLog         *ChronTree
	verifyUpdatePrefixTrees []*prefixTree

	currUpdatePeriodUpdates [][][]byte

	currVerifyPeriodUpdates [][][]byte

	epoch              uint32
	verificationPeriod uint64

	tmpdir string
}

var _ LegoLogPartition = (*AggHistPartition)(nil)

func NewAggHistPartition(cfg Config, tmpdir string) *AggHistPartition {
	ret := AggHistPartition{
		cfg:                     cfg,
		baseTree:                NewPersistentPrefixTree(),
		baseTreeForest:          NewHistoryForest(cfg.AggHistoryDepth),
		queryUpdateLog:          NewChronTree(),
		verifyUpdateLog:         NewChronTree(),
		epoch:                   0,
		currUpdatePeriodUpdates: [][][]byte{{}, {}},
		currVerifyPeriodUpdates: [][][]byte{{}, {}},
		tmpdir:                  tmpdir,
	}
	return &ret
}

func (p *AggHistPartition) Append(username []byte, identifier []byte, value []byte, signature []byte) {
	//fmt.Printf("partition_agghist.go: Append\n")
	id_hash := GetPrefixFromIdentifier(identifier)
	hashBytes := ConvertBitsToBytes(ComputeLeafNodeHash(identifier, value, signature, 0)) // todo: fix this position
	p.currUpdatePeriodUpdates[0] = append(p.currUpdatePeriodUpdates[0], id_hash)
	p.currUpdatePeriodUpdates[1] = append(p.currUpdatePeriodUpdates[1], hashBytes)
	p.currVerifyPeriodUpdates[0] = append(p.currVerifyPeriodUpdates[0], id_hash)
	p.currVerifyPeriodUpdates[1] = append(p.currVerifyPeriodUpdates[1], hashBytes)
	p.baseTree.Insert(id_hash, hashBytes, 0)
	//fmt.Printf("Leaving partition_agghist.go: Append\n")
}

func (p *AggHistPartition) IncrementUpdateEpoch() {
	// add Hash(id), Hash(id, val, pos) []
	// two arrays
	prefixTree := NewPrefixTree()
	for i, id_hash := range p.currUpdatePeriodUpdates[0] {
		prefixTree.PrefixAppend(id_hash, p.currUpdatePeriodUpdates[1][i], 0)
	}
	p.queryUpdatePrefixTrees = append(p.queryUpdatePrefixTrees, prefixTree)
	p.verifyUpdatePrefixTrees = append(p.verifyUpdatePrefixTrees, prefixTree)
	for _, id_hash := range p.currUpdatePeriodUpdates[0] {
		hasKey := prefixTree.getLeaf(id_hash) != nil
		if !hasKey {
			fmt.Println("[should not happen] server doesn't have key for hash ", id_hash)
		}
	}
	p.currUpdatePeriodUpdates = [][][]byte{{}, {}}
	p.epoch += 1
	p.queryUpdateLog.Append([]byte(strconv.Itoa(int(p.epoch))), prefixTree.getHash(), []byte(""))
	p.verifyUpdateLog.Append([]byte(strconv.Itoa(int(p.epoch))), prefixTree.getHash(), []byte(""))

	// fmt.Println("partition_agghist.go: IncrementUpdateEpoch")
	// fmt.Println(p.baseTreeForest.Roots, p.queryUpdateSetTrees, p.verifyUpdateSetTrees)
}

func (p *AggHistPartition) IncrementVerificationPeriod() error {
	p.verificationPeriod += 1
	p.baseTree.NextEpoch()

	// Stick the root prefix hash in forest
	if p.verificationPeriod >= 2 {
		p.baseTreeForest.Append(p.baseTree.getHash(p.verificationPeriod-2), p.verificationPeriod-2)
	}

	// Set the next query update log to the old verify update log
	p.currVerifyPeriodUpdates = [][][]byte{{}, {}}
	p.queryUpdateLog = p.verifyUpdateLog
	p.queryUpdatePrefixTrees = p.verifyUpdatePrefixTrees
	// Clear the verify update log
	p.verifyUpdateLog = NewChronTree()
	p.verifyUpdatePrefixTrees = []*prefixTree{}
	// fmt.Println("partition_agghist.go: IncrementVerificationPeriod")
	// fmt.Println(p.baseTreeForest.Roots, p.queryUpdateSetTrees, p.verifyUpdateSetTrees)
	//err = p.lastOffloadedBaseTree.OffloadToDisk()
	//p.lastOffloadedBaseTree = offloadedQueryBaseTree
	//if err != nil {
	//	fmt.Println(err)
	//	return err
	//}
	return nil
}

func (p *AggHistPartition) GenerateExistenceProof(identifier []byte, value []byte, signature []byte) *LegologExistenceProof {
	//fmt.Println(p.baseTreeForest.Roots, p.baseTreeMap, p.queryUpdateSetTrees, p.verifyUpdateSetTrees)
	//fmt.Println("partition_agghist.go: GenerateExistenceProof")
	proof := LegologExistenceProof{
		BaseTreeProofs:  nil,
		UpdateLogProofs: []*MembershipOrNonmembershipProof{},
	}

	id_hash := GetPrefixFromIdentifier(identifier)

	// fmt.Println("len(p.baseTreeForest.Roots)", len(p.baseTreeForest.Roots))

	for _, histNode := range p.baseTreeForest.Roots {
		startTime := time.Now()
		baseTreeProof := &MembershipOrNonmembershipProof{
			MembershipProof:    nil,
			NonMembershipProof: nil,
			ValueExists:        false,
			LeafValues:         []KeyHash{},
		}
		leaf := p.baseTree.getLeaf(id_hash, histNode.getVerificationPeriod())

		if leaf != nil {
			btProof, _ := p.baseTree.generateMembershipProof(id_hash, histNode.getVerificationPeriod())
			baseTreeProof.ValueExists = true
			baseTreeProof.MembershipProof = btProof
			baseTreeProof.LeafValues = leaf.getValues()
		} else {
			btProof := p.baseTree.generateNonMembershipProof(id_hash, histNode.getVerificationPeriod()) // TODO: verify
			baseTreeProof.ValueExists = false
			baseTreeProof.NonMembershipProof = btProof
			//fmt.Println("base tree proof", *btProof)
		}
		proof.BaseTreeProofs = append(proof.BaseTreeProofs, baseTreeProof)
		timeTaken := time.Since(startTime)
		_ = timeTaken
		// fmt.Printf("Time taken to generate base tree proof %d: %d ns\n", i, timeTaken.Nanoseconds())
	}

	/*
		for _, setTree := range append(p.queryUpdateSetTrees, p.verifyUpdateSetTrees...) {
			updateLogProof := MembershipOrNonmembershipProof{
				MembershipProof: nil, NonMembershipProof: nil,
			}
			// hashBytes := ConvertBitsToBytes(ComputeContentHash(identifier, value, signature, 0))

			//valueHash := ConvertBitsToBytes(ComputeLeafNodeHash(identifier, value, signature, 0))
			if setTree.HasKey(id_hash) {
				leaf := setTree.GetLeaf(id_hash)
				leafProof, _ := setTree.ProveExistence(id_hash)
				updateLogProof.MembershipProof = leafProof
				updateLogProof.ValueExists = true
				updateLogProof.LeafValues = leaf.getValues()
			} else {
				leafProof, _ := setTree.ProveNonExistence(id_hash)
				updateLogProof.MembershipProof = leafProof
				updateLogProof.ValueExists = false

			}
			proof.UpdateLogProofs = append(proof.UpdateLogProofs, &updateLogProof)
		}
	*/
	for _, prefixTree := range p.queryUpdatePrefixTrees {
		updateLogProof := MembershipOrNonmembershipProof{
			MembershipProof: nil, NonMembershipProof: nil,
		}
		// hashBytes := ConvertBitsToBytes(ComputeContentHash(identifier, value, signature, 0))

		//valueHash := ConvertBitsToBytes(ComputeLeafNodeHash(identifier, value, signature, 0))
		//if prefixTree.HasValue(id_hash, valueHash) {
		if prefixTree.getLeaf(id_hash) != nil {
			leafProof, leafValues := prefixTree.generateMembershipProof(id_hash) // prefixTree.ProveExistence(id_hash)
			updateLogProof.MembershipProof = leafProof
			updateLogProof.ValueExists = true
			updateLogProof.LeafValues = leafValues
		} else {
			leafProof := prefixTree.generateNonMembershipProof(id_hash)

			// ProveNonExistence(id_hash)
			updateLogProof.NonMembershipProof = leafProof
			updateLogProof.ValueExists = false
		}
		proof.UpdateLogProofs = append(proof.UpdateLogProofs, &updateLogProof)
	}

	// TODO: prove that each update prefix tree is in the update log. we can do this by just sending over the full tree.

	return &proof
}

func (p *AggHistPartition) GetDigest() *LegologDigest {
	/*
		updateSetRoots := make([][]byte, 0)
		for _, updateTree := range append(p.queryUpdateSetTrees, p.verifyUpdateSetTrees...) {
			updateSetRoots = append(updateSetRoots, updateTree.tree.root.hash)
		}
		baseTreeRoots := make([][]byte, 0)
		for _, baseTree := range p.baseTreeForest.Roots {
			baseTreeRoots = append(baseTreeRoots, baseTree.getNewestLeafHash())
		}
	*/
	updatePrefixTreeRoots := make([][]byte, 0)
	for _, updateTree := range append(p.queryUpdatePrefixTrees, p.verifyUpdatePrefixTrees...) {
		updatePrefixTreeRoots = append(updatePrefixTreeRoots, updateTree.root.hash)
	}
	baseTreeRoots := make([][]byte, 0)
	for _, baseTree := range p.baseTreeForest.Roots {
		baseTreeRoots = append(baseTreeRoots, baseTree.getNewestLeafHash())
	}
	return &LegologDigest{
		BaseTreeRoots:  baseTreeRoots,
		UpdateLogRoot:  p.verifyUpdateLog.GetRootHash(),
		UpdateSetRoots: updatePrefixTreeRoots,
	}
}

func (p *AggHistPartition) GetUpdateEpochConsistencyProof(oldSize uint32) *MerkleExtensionProof {
	return nil
}

type AggHistVerifier struct {
}

func (AggHistVerifier) ValidatePKProof(oldDigest *LegologDigest, proof *LegologExistenceProof, identifier []byte, value []byte, signature []byte, pos uint64, masterVK []byte) (bool, error) {
	var latestExistenceProof *MembershipOrNonmembershipProof = nil
	var latestTreeRoot []byte
	idx := 0
	valueExists := false
	isBaseTreeProof := true

	for i, baseTreeProof := range proof.BaseTreeProofs {
		if baseTreeProof.ValueExists {
			latestExistenceProof = baseTreeProof
			latestTreeRoot = oldDigest.BaseTreeRoots[i]
			idx = i
			valueExists = true

			success, err := validateExistenceProof(latestExistenceProof, identifier, value, signature, pos, masterVK, latestTreeRoot, false)

			if !success {
				fmt.Printf("(isBaseTree %t) update log/base tree %d: %s\n", isBaseTreeProof, idx, err.Error())
				// return false, fmt.Errorf("(isBaseTree %t) update log/base tree %d: %s", isBaseTreeProof, idx, err.Error())
			}
		} else {
			success := validateNonMembershipProof(baseTreeProof.NonMembershipProof, identifier, oldDigest.BaseTreeRoots[i])
			if !success {
				return false, errors.New("non membership proof does not go through")
			}
		}
		if valueExists && !baseTreeProof.ValueExists {
			return false, errors.New("membership proof does not exist in a later base tree")
		}
	}

	// TODO: do existence and non existence validation here instead, just to be sure!
	for i, updateSetProof := range proof.UpdateLogProofs { // TODO: I think there should only be one update log proof here
		if updateSetProof.ValueExists {
			latestExistenceProof = updateSetProof
			idx = i
			latestTreeRoot = oldDigest.UpdateSetRoots[i]
			isBaseTreeProof = false
		} else {
			nonMembershipProof := updateSetProof.NonMembershipProof
			success := validateNonMembershipProof(nonMembershipProof, identifier, oldDigest.UpdateSetRoots[i])
			if !success {
				return false, errors.New("non membership proof does not go through")
			}
		}
	}

	if latestExistenceProof == nil {
		return false, errors.New("unable to find any existence proofs")
	}
	success, err := validateExistenceProof(latestExistenceProof, identifier, value, signature, pos, masterVK, latestTreeRoot, false)

	if !success {
		return false, fmt.Errorf("(isBaseTree %t) update log/base tree %d: %s", isBaseTreeProof, idx, err.Error())
	}

	return true, nil
}

func (AggHistVerifier) ValidatePKProofMonitoring(oldDigest *LegologDigest, proof *LegologExistenceProof, identifier []byte, value []byte, signature []byte, pos uint64, masterVK []byte) (bool, error) {
	var latestExistenceProof *MembershipOrNonmembershipProof = nil
	var latestTreeRoot []byte
	idx := 0
	valueExists := false
	isBaseTreeProof := true

	// fmt.Println("\nlen(proof.BaseTreeProofs): ", len(proof.BaseTreeProofs))

	for i, baseTreeProof := range proof.BaseTreeProofs {
		if baseTreeProof.ValueExists {
			// fmt.Println("ValidatePKProofMonitoring: found base tree proof")
			startTime := time.Now()
			latestExistenceProof = baseTreeProof
			latestTreeRoot = oldDigest.BaseTreeRoots[i]
			idx = i
			valueExists = true

			// validate all existence proofs insteaed of just the latest one
			success, err := validateExistenceProof(latestExistenceProof, identifier, value, signature, pos, masterVK, latestTreeRoot, false)

			if !success {
				fmt.Printf("(isBaseTree %t) update log/base tree %d: %s\n", isBaseTreeProof, idx, err.Error())
				// return false, fmt.Errorf("(isBaseTree %t) update log/base tree %d: %s", isBaseTreeProof, idx, err.Error())
			}
			timeTaken := time.Since(startTime)
			_ = timeTaken
			// fmt.Printf("Time taken to validate base tree proof %d: %d ns\n\n", i, timeTaken.Nanoseconds())
		} else {
			success := validateNonMembershipProof(baseTreeProof.NonMembershipProof, identifier, oldDigest.BaseTreeRoots[i])
			if !success {
				return false, errors.New("non membership proof does not go through")
			}
		}
		if valueExists && !baseTreeProof.ValueExists {
			return false, errors.New("membership proof does not exist in a later base tree")
		}
	}
	// fmt.Println("ValidatePKProofMonitoring: finished validating base tree proofs")

	/*
		When monitoring, we only check the base trees, not the update logs.

		// TODO: do existence and non existence validation here instead, just to be sure!
		for i, updateSetProof := range proof.UpdateLogProofs { // TODO: I think there should only be one update log proof here
			if updateSetProof.ValueExists {
				latestExistenceProof = updateSetProof
				idx = i
				latestTreeRoot = oldDigest.UpdateSetRoots[i]
				isBaseTreeProof = false
			} else {
				nonMembershipProof := updateSetProof.NonMembershipProof
				success := validateNonMembershipProof(nonMembershipProof, identifier, oldDigest.UpdateSetRoots[i])
				if !success {
					return false, errors.New("non membership proof does not go through")
				}
			}
		}
	*/
	// fmt.Println("latest existence proof is ", latestExistenceProof)
	if latestExistenceProof == nil {
		return false, errors.New("unable to find any existence proofs")
	}

	return true, nil
}

func validateNonMembershipProof(proof *NonMembershipProof, id []byte, rootHash []byte) bool {
	if rootHash == nil {
		return true
	}
	prefix := makePrefixFromKey(id)
	computedRootHash := computeRootHashNonMembership(prefix, proof)
	//fmt.Println("computedRootHash", computedRootHash)
	//fmt.Println("expected:", rootHash)
	return bytes.Equal(computedRootHash, rootHash)
}

// Prove that this is the first value in the tree
func (AggHistVerifier) ValidateMKProof(oldDigest *LegologDigest, proof *LegologExistenceProof, username []byte, value []byte, signature []byte, pos uint64, masterVK []byte) (bool, error) {
	identifier := append(username, []byte("MK")...)
	for i, baseTreeProof := range proof.BaseTreeProofs {
		if baseTreeProof.ValueExists {
			return false, errors.New("MK should not exist in the tree yet")
		} else {
			ok := validateNonMembershipProof(baseTreeProof.NonMembershipProof, identifier, oldDigest.BaseTreeRoots[i])
			if !ok {
				return false, errors.New("non membership proof does not go through")
			}
		}
	}

	proved := false
	for i, updateSetProof := range proof.UpdateLogProofs { // TODO: I think there should only be one update log proof here
		if updateSetProof.ValueExists {
			if proved {
				return false, errors.New("multiple existence proofs for MK found")
			}
			existenceProof := updateSetProof
			success, err := validateExistenceProof(existenceProof, identifier, masterVK, signature, pos, masterVK, oldDigest.UpdateSetRoots[i], true)
			if !success {
				return false, err
			}
			proved = true
		} else {
			nonexistenceProof := updateSetProof
			success, err := validateNonExistenceProof(nonexistenceProof, identifier, oldDigest.UpdateSetRoots[i])
			if !success {
				return false, err
			}
		}
	}

	if !proved {
		return false, errors.New("no existence proof found")
	}
	return true, nil
}
