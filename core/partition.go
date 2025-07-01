package core

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	libcrypto "github.com/huyuncong/MerkleSquare/lib/crypto"
	"github.com/immesys/bw2/crypto"
)

type LegoLogPartition interface {
	Append(username []byte, identifier []byte, value []byte, signature []byte)
	GenerateExistenceProof(identifier []byte, value []byte, signature []byte) *LegologExistenceProof
	IncrementUpdateEpoch()
	IncrementVerificationPeriod() error
	GetDigest() *LegologDigest
	GetUpdateEpochConsistencyProof(oldSize uint32) *MerkleExtensionProof
}

// assert that Partition implements LegoLogPartition
var _ LegoLogPartition = (*Partition)(nil)

/*
TODO:
1. fix set tree construction, both in incremenetupdateepoch and in the lookup done in GenerateExistenceProof
2. debug why the validate proof fails
	akshit: previously failed at having no existence proofs. it had no existence proofs because GenerateExistenceProof wasn't able
	to find anything in the set tree.
*/

type Partition struct {
	baseTree *persistentPrefixTree

	queryUpdateLog         *ChronTree
	queryUpdatePrefixTrees []*prefixTree

	verificationUpdatePrefixTrees []*prefixTree
	verificationUpdateLog         *ChronTree

	//  the updates from the current epoch that get rolled up into the set tree;
	//  latestUpdates[0] = list of H(identifier), latestUpdates[1] = list of H(id, value, signature, pos)
	latestUpdates     [][][]byte
	epoch             uint64
	verificationEpoch uint64
	/* 	PublishedDigest *LegologDigest
	 */
	pos       uint32
	hashChain []byte
}

// Digest struct for snapshots of the current state of MerkleSquare
type LegologDigest struct {
	/* 	Roots [][]byte
	   	Size  uint32 */
	BaseTreeRoots      [][]byte
	BaseTreeSize       uint32
	UpdateLogRoot      []byte
	UpdateSetRoots     [][]byte
	UpdateLogSize      uint32
	Epoch              uint64
	HashChain          []byte
	HistoryForestRoots [][]byte
}

type LegologExistenceProof struct {
	BaseTreeProofs  []*MembershipOrNonmembershipProof
	UpdateLogProofs []*MembershipOrNonmembershipProof
}

func (p *LegologExistenceProof) String() string {
	var buf bytes.Buffer
	buf.WriteString("Base Tree Proofs:\n")
	for i, proof := range p.BaseTreeProofs {
		buf.WriteString(fmt.Sprintf("\tProof %d: %s\n", i, proof))
	}
	buf.WriteString("Update Log Proofs:\n")
	for i, proof := range p.UpdateLogProofs {
		buf.WriteString(fmt.Sprintf("\tProof %d: %s\n", i, proof))
	}
	return buf.String()
}

func (p *MembershipOrNonmembershipProof) String() string {
	var buf bytes.Buffer
	if p.ValueExists {
		buf.WriteString(fmt.Sprintf("\tMembership: %s\n", p.MembershipProof))
	} else {
		buf.WriteString(fmt.Sprintf("\tNon-Membership: %s\n", p.NonMembershipProof))
	}
	buf.WriteString(fmt.Sprintf("\tLeaf Values: %v\n", p.LeafValues))
	return buf.String()
}

func (p *MembershipProof) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("\t\tLeaf Hash: %s\n", p.LeafPartialPrefix))
	// buf.WriteString(fmt.Sprintf("\t\tPath: %v\n", p.CopathNodes))
	return buf.String()
}

func (p *NonMembershipProof) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("\t\tLeaf partial prefix: %s\n", p.EndNodePartialPrefix))
	buf.WriteString(fmt.Sprintf("\t\tHash: %x\n", p.EndNodeHash))
	//buf.WriteString(fmt.Sprintf("\t\tPath: %v\n", p.CopathNodes))
	return buf.String()
}

func NewPartition() *Partition {

	// leafNode := &leafChronNode {
	// 	parent: nil,
	// 	hash:   []byte{},
	// }
	// leafNode.parent = leafNode

	// soleNode := &internalChronNode {
	// 	parent:     nil,
	// 	leftChild:  nil,
	// 	rightChild: nil,
	// 	hash:       []byte{},
	// 	depth:      0,
	// }

	p := &Partition{
		baseTree:              NewPersistentPrefixTree(),
		queryUpdateLog:        NewChronTree(),
		verificationUpdateLog: NewChronTree(),
		latestUpdates:         [][][]byte{{}, {}},
		verificationEpoch:     0,
	}
	p.IncrementVerificationPeriod()
	p.IncrementVerificationPeriod()
	return p
}

func (p *Partition) GetDigest() *LegologDigest {
	/*
		updateSetRoots := make([][]byte, 0)
		for _, updateTree := range p.queryUpdateSetTrees {
			updateSetRoots = append(updateSetRoots, updateTree.tree.root.hash)
		}
	*/
	updatePrefixTreeRoots := make([][]byte, 0)
	for _, updateTree := range p.queryUpdatePrefixTrees {
		updatePrefixTreeRoots = append(updatePrefixTreeRoots, updateTree.root.hash)
	}
	return &LegologDigest{
		BaseTreeRoots:  [][]byte{p.baseTree.getHash(p.verificationEpoch - 2), p.baseTree.getHash(p.verificationEpoch - 1)},
		BaseTreeSize:   uint32(p.baseTree.getSize(p.verificationEpoch - 2)), // TODO: @vivian is this the right tree?
		UpdateLogRoot:  p.queryUpdateLog.root.getHash(),
		UpdateSetRoots: updatePrefixTreeRoots,
		UpdateLogSize:  p.queryUpdateLog.numNodes, // TODO: @vivian is this the right log?
		Epoch:          p.epoch,
		HashChain:      p.hashChain,
	}
}

func GetPrefixFromIdentifier(identifier []byte) []byte {
	return ConvertBitsToBytes(libcrypto.Hash(identifier))
}

func (p *Partition) GetUpdateEpochConsistencyProof(oldSize uint32) *MerkleExtensionProof {
	newSize := p.verificationUpdateLog.numNodes
	//fmt.Println("oldSize", oldSize, "newSize", newSize)
	if newSize < oldSize {
		return &MerkleExtensionProof{}
	}
	if oldSize > newSize {
		panic("oldSize > newSize")
	}
	proof := p.verificationUpdateLog.GenerateConsistencyProof(oldSize, newSize)
	return proof
}

func (p *Partition) Append(username []byte, identifier []byte, value []byte, signature []byte) {
	// TODO
	// look at chron_node tree append fxn
	/* 	fmt.Printf("partition.go: Append\n")
	 */id_hash := GetPrefixFromIdentifier(identifier)
	hashBytes := ConvertBitsToBytes(ComputeLeafNodeHash(identifier, value, signature, 0)) // todo: fix this position
	// fmt.Printf("before append, length of latestUpdates is %d\n", len(p.latestUpdates[0]))
	p.latestUpdates[0] = append(p.latestUpdates[0], id_hash)
	p.latestUpdates[1] = append(p.latestUpdates[1], hashBytes)
	// fmt.Printf("appended to latestUpdates[0], is now of length %d\n", len(p.latestUpdates[0]))
	/* 	p.updateLog.Append(identifier, value, signature) */

	p.baseTree.Insert(id_hash, hashBytes, p.pos)
	//p.nextVerificationBaseTree.PrefixAppend(id_hash, hashBytes, p.pos)
	p.pos += 1
	/* 	fmt.Printf("Leaving partition.go: Append\n")
	 */ // marshalled, _ := json.Marshal(p.latestUpdates)
	// fmt.Println("Byte Array is ", marshalled)
	// fmt.Println("Byte Array without marshalling is ", p.latestUpdates)
	/* MERKLE ^2 Implementation
	if m.isFull() {
		return // throw error?
	}
	node := m.next.(*LeafNode)
	node.completeLeaf(key, value, signature, m.Size)
	m.appendToPrefixTrees(node, node.getPrefix(), node.getContentHash(), m.Size)
	m.Size++

	p := m.next

	for p.isRightChild() {
		p = p.getParent()
		p.complete()
		m.pop()
	}

	m.addRoot(p)

	// check to see if tree is full
	if m.isFull() {
		return
	}
	newNode := p.getParent().createRightChild()
	m.addNodeToMap(newNode)
	p = p.getParent().getRightChild()

	for p.getDepth() > 0 {
		newNode = p.createLeftChild()
		m.addNodeToMap(newNode)
		p = p.getLeftChild()
	}

	m.next = p
	*/
}

// MembershipProof or NonMembershipProof
type MembershipOrNonmembershipProof struct {
	MembershipProof    *MembershipProof
	NonMembershipProof *NonMembershipProof // ask about this type!
	ValueExists        bool
	LeafValues         []KeyHash
}

func (p *Partition) GenerateExistenceProof(identifier []byte, value []byte, signature []byte) *LegologExistenceProof {

	proof := LegologExistenceProof{
		BaseTreeProofs:  nil,
		UpdateLogProofs: []*MembershipOrNonmembershipProof{},
	}

	baseTreeProof := &MembershipOrNonmembershipProof{
		MembershipProof:    nil,
		NonMembershipProof: nil,
		ValueExists:        false,
		LeafValues:         []KeyHash{},
	}
	proof.BaseTreeProofs = []*MembershipOrNonmembershipProof{baseTreeProof}

	id_hash := GetPrefixFromIdentifier(identifier)
	//leaf := p.queryBaseTree.getLeaf(id_hash)
	leaf := p.baseTree.GetLeaf(id_hash, p.verificationEpoch-2)

	if leaf != nil {
		btProof, _ := p.baseTree.generateMembershipProof(id_hash, p.verificationEpoch-2)
		proof.BaseTreeProofs[0].ValueExists = true
		proof.BaseTreeProofs[0].MembershipProof = btProof
		proof.BaseTreeProofs[0].LeafValues = leaf.getValues()
	} else {
		btProof := p.baseTree.generateNonMembershipProof(id_hash, p.verificationEpoch-2) // TODO: verify
		proof.BaseTreeProofs[0].ValueExists = false
		proof.BaseTreeProofs[0].NonMembershipProof = btProof
	}

	/*
		use prefix trees instead of set trees
			for _, setTree := range p.queryUpdateSetTrees {

				updateLogProof := MembershipOrNonmembershipProof{
					MembershipProof: nil, NonMembershipProof: nil,
				}
				id_hash := ConvertBitsToBytes(libcrypto.Hash(identifier))
				// hashBytes := ConvertBitsToBytes(ComputeContentHash(identifier, value, signature, 0))

				valueHash := ConvertBitsToBytes(ComputeLeafNodeHash(identifier, value, signature, 0))
				if setTree.HasValue(id_hash, valueHash) {
					leaf = setTree.GetLeaf(id_hash)
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
	// fmt.Printf("lookin thru %d prefix trees\n", len(p.queryUpdatePrefixTrees))
	for _, prefixTree := range p.queryUpdatePrefixTrees {

		updateLogProof := MembershipOrNonmembershipProof{
			MembershipProof: nil, NonMembershipProof: nil,
		}
		id_hash := ConvertBitsToBytes(libcrypto.Hash(identifier))
		// hashBytes := ConvertBitsToBytes(ComputeContentHash(identifier, value, signature, 0))
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

func (p *Partition) IncrementUpdateEpoch() {
	// add Hash(id), Hash(id, val, pos) []
	// two arrays
	// fmt.Println("increment update epoch called")

	// print p.latestupdates[0] and p.latestUpdates[1] in a readable format
	// fmt.Printf("latestUpdates[0] of length %d: %v\n\n", len(p.latestUpdates[0]), p.latestUpdates[0])
	// fmt.Printf("latestUpdates[1] of length %d: %v\n\n", len(p.latestUpdates[1]), p.latestUpdates[1])
	prefixTree := NewPrefixTree()
	for i, id_hash := range p.latestUpdates[0] {
		prefixTree.PrefixAppend(id_hash, p.latestUpdates[1][i], 0)
	}
	p.queryUpdatePrefixTrees = append(p.queryUpdatePrefixTrees, prefixTree)
	p.verificationUpdatePrefixTrees = append(p.verificationUpdatePrefixTrees, prefixTree)
	for _, id_hash := range p.latestUpdates[0] {
		hasKey := prefixTree.getLeaf(id_hash) != nil
		if hasKey {
			// fmt.Println("server has key for hash ", id_hash)
		} else {
			fmt.Println("[should not happen] server doesn't have key for hash ", id_hash)
		}
	}
	p.latestUpdates = [][][]byte{{}, {}}
	p.epoch += 1
	p.verificationUpdateLog.Append([]byte(strconv.Itoa(int(p.epoch))), prefixTree.getHash(), []byte(""))
	p.queryUpdateLog.Append([]byte(strconv.Itoa(int(p.epoch))), prefixTree.getHash(), []byte(""))
}

func (p *Partition) IncrementVerificationPeriod() (err error) {
	p.baseTree.NextEpoch()
	p.verificationEpoch += 1

	p.queryUpdateLog = p.verificationUpdateLog
	p.queryUpdatePrefixTrees = p.verificationUpdatePrefixTrees
	p.verificationUpdateLog = NewChronTree()
	p.verificationUpdatePrefixTrees = []*prefixTree{}
	if p.verificationEpoch >= 2 {
		p.hashChain = libcrypto.Hash(p.baseTree.getHash(p.verificationEpoch-2), p.baseTree.getHash(p.verificationEpoch-1), []byte(strconv.FormatUint(p.epoch, 10)), p.hashChain)
	}
	return
}

func validateExistenceProof(proof *MembershipOrNonmembershipProof, identifier []byte, value []byte, signature []byte, pos uint64, masterVK []byte, reportedRoot []byte, isMK bool) (bool, error) {
	existenceProof := proof
	prefix := GetPrefixFromIdentifier(identifier)
	startTime := time.Now()
	computedRoot := computeRootHashMembership(prefix, existenceProof.MembershipProof, existenceProof.LeafValues)
	elapsedTime := time.Since(startTime)
	_ = elapsedTime
	// fmt.Printf("time to compute root hash membership: %d ns\n", elapsedTime.Nanoseconds())
	// fmt.Printf("length of copath nodes: %d\n", len(proof.MembershipProof.CopathNodes))
	// fmt.Printf("length of leaf values : %d\n", len(proof.LeafValues))
	// TODO: check that h(id, val) is is the same as what is in the actual tree; see
	// idk wtf to do with signature

	var foundMatch = false
	// TODO: fix position eventually
	expectedLeafNodeHash := ConvertBitsToBytes(ComputeLeafNodeHash(identifier, value, signature, 0))
	for _, leafNodeHash := range existenceProof.LeafValues {
		if bytes.Equal(leafNodeHash.Hash, expectedLeafNodeHash) {
			foundMatch = true
		}
	}

	if !foundMatch {
		return false, errors.New("unable to find leaf node hash in provided merkle tree")
	}

	signatureValue := append(value, []byte(strconv.Itoa(int(pos)))...)
	if isMK {
		signatureValue = value
	}
	if !crypto.VerifyBlob(masterVK, signature, signatureValue) {
		return false, errors.New("unable to verify signature on value")
	}

	if !bytes.Equal(computedRoot, reportedRoot) {
		return false, errors.New(fmt.Sprintf("existence: computed root doesn't match reported root. Expected %x, computed %x", reportedRoot, computedRoot))
	}

	return true, nil
}

type ProofVerifier interface {
	ValidatePKProof(oldDigest *LegologDigest, proof *LegologExistenceProof, identifier []byte, value []byte, signature []byte, pos uint64, masterVK []byte) (bool, error)
	ValidateMKProof(oldDigest *LegologDigest, proof *LegologExistenceProof, username []byte, value []byte, signature []byte, pos uint64, masterVK []byte) (bool, error)
}

/*
Non-existence proof validation outline
for each proof:
 1. check that the prefix in the nonExistenceProof is a prefix of GetPrefixFromIdentifier(identifier)
 2. compute the root using the specified values and check that it matches the digest
 3. check that the leafValues inside the leaf node contain the hash that signifies no values are here:
    not sure what, but something along the lines of Hash(identifier, value, 0) where value is ??
*/
func validateNonExistenceProof(proof *MembershipOrNonmembershipProof, identifier []byte, reportedRoot []byte) (bool, error) {
	nonexistenceProof := proof
	if nonexistenceProof.ValueExists {
		return false, errors.New("expected nonexistence proof, but got existence")
	}

	fullPrefix := GetPrefixFromIdentifier(identifier)
	frontierPrefix := append(getPrefix(nonexistenceProof.MembershipProof.CopathNodes), nonexistenceProof.MembershipProof.LeafPartialPrefix...)
	if !bytes.HasPrefix(fullPrefix, frontierPrefix) {
		return false, errors.New("prefix of nonexistence proof is not a prefix of the identifier")
	}

	if len(nonexistenceProof.LeafValues) != 1 {
		return false, errors.New("expected leaf node in non existence proof to only have one value")
	}

	leaf := nonexistenceProof.LeafValues[0]
	if leaf.Pos != 0 {
		return false, fmt.Errorf("expected frontier node position to be 0, but got %d", leaf.Pos)
	}

	if !bytes.Equal(leaf.Hash, frontierPrefix) {
		return false, fmt.Errorf("Expected value stored in frontier node to be equal to prefix. Instead, got prefix=%b and value=%b", frontierPrefix, leaf.Hash)
	}

	computedRoot := computeRootHashMembership(frontierPrefix, nonexistenceProof.MembershipProof, nonexistenceProof.LeafValues)

	if !bytes.Equal(computedRoot, reportedRoot) {
		return false, errors.New(fmt.Sprintf("non-existence: computed root doesn't match reported root. Expected %s, computed %s", reportedRoot, computedRoot))
	}
	return true, nil
}

func ValidatePKProof(oldDigest *LegologDigest, proof *LegologExistenceProof, identifier []byte, value []byte, signature []byte, pos uint64, masterVK []byte) (bool, error) {
	// this function is characteristically pretty similar to VerifyExistenceProof in core.

	var latestExistenceProof *MembershipOrNonmembershipProof = nil
	latestIndex := -1

	// TODO: assert that this is an existence proof, something with the else
	if proof.BaseTreeProofs[0].ValueExists {
		latestExistenceProof = proof.BaseTreeProofs[0]
		if proof.BaseTreeProofs[0].MembershipProof == nil {
			return false, errors.New("base tree membership proof is nil when it is expected")
		}
	} else {
		if proof.BaseTreeProofs[0].NonMembershipProof == nil {
			return false, errors.New("base tree non-membership proof is nil when it is expected")
		}
		// TODO: verify this proof
		// verify it doesn't exist in base tree
		// proof.BaseTreeProof.NonMembershipProof =
	}

	// TODO: do existence and non existence validation here instead, just to be sure!
	for i, updateSetProof := range proof.UpdateLogProofs { // TODO: I think there should only be one update log proof here
		if updateSetProof.ValueExists {
			latestExistenceProof = updateSetProof
			latestIndex = i
		} else {
			/*
				TODO: verify this proof
					nonexistenceProof := proof.UpdateLogProofs[i]
					success, err := validateNonExistenceProof(nonexistenceProof, identifier, oldDigest.UpdateSetRoots[i])
					if !success {
						return false, err
					}
			*/
		}
	}

	if latestExistenceProof == nil {
		return false, errors.New("unable to find any existence proofs")
	}
	success, err := validateExistenceProof(latestExistenceProof, identifier, value, signature, pos, masterVK, oldDigest.UpdateSetRoots[latestIndex], false)

	if !success {
		return false, fmt.Errorf("Update log %d: %s", latestIndex, err.Error())
	}

	return true, nil
}

func ValidateMKProof(oldDigest *LegologDigest, proof *LegologExistenceProof, username []byte, value []byte, signature []byte, pos uint64, masterVK []byte) (bool, error) {
	// TODO: assert that this is an existence proof, something with the else
	if proof.BaseTreeProofs[0].ValueExists {
		if proof.BaseTreeProofs[0].MembershipProof == nil {
			return false, errors.New("base tree membership proof is nil when it is expected")
		}
	} else {
		if proof.BaseTreeProofs[0].NonMembershipProof == nil {
			return false, errors.New("base tree non-membership proof is nil when it is expected")
		}
		// TODO: verify this proof
	}

	identifier := append(username, []byte("MK")...)
	for i, updateSetProof := range proof.UpdateLogProofs { // TODO: I think there should only be one update log proof here
		if updateSetProof.ValueExists {
			existenceProof := updateSetProof
			success, err := validateExistenceProof(existenceProof, identifier, masterVK, signature, pos, masterVK, oldDigest.UpdateSetRoots[i], true)
			if !success {
				return false, err
			}
			return true, nil
		} else {
			nonexistenceProof := updateSetProof
			success, err := validateNonExistenceProof(nonexistenceProof, identifier, oldDigest.UpdateSetRoots[i])
			if !success {
				return false, err
			}
		}
	}

	return false, errors.New("no existence proof found")
}

// ProveNonexistence provides a nonexistence proof for a given MSK/key
// TODO: fix
/*
func (p *Partition) ProveNonexistence(key []byte, mskPos uint32, currSize uint32) *NonExistenceProof {

	prefix := makePrefixFromKey(key)
	oldRoots := m.getOldRoots(mskPos)

	nonMemberShipProofs := []NonMembershipProof{}
	childHashes := [][]byte{}
	var leafHash LeafHash

	for _, root := range oldRoots {
		if root.isLeafNode() {
			leafHash = LeafHash{
				Prefix:          root.getPrefix(),
				NodeContentHash: root.getContentHash(),
			}
			break
		}

		childHashes = append(childHashes, root.getLeftChild().getHash())
		childHashes = append(childHashes, root.getRightChild().getHash())

		prefixTree := root.getPrefixTree()
		nonMemberShipProofs = append(nonMemberShipProofs, *prefixTree.generateNonMembershipProof(prefix))
	}

	extensionProof := m.GenerateExtensionProof(mskPos, currSize)

	return &NonExistenceProof{
		NonMembershipProofs: nonMemberShipProofs,
		ChildHashes:         childHashes,
		ExtensionProof:      *extensionProof,
		LeafHash:            leafHash,
	}
}

*/
