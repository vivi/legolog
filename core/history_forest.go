package core

import (
	"bytes"
	"encoding/binary"
	"math/bits"
)

type HistoryForest struct {
	Roots []HistoryNode
	root  HistoryNode
	next  HistoryNode
	Size  uint32
	depth uint32
}

//*******************************
// CORE METHODS
//*******************************

func (m *HistoryForest) Append(prefixTreeHash []byte, verificationPeriod uint64) {
	if m.isFull() {
		return // throw error?
	}
	node := m.next.(*LeafHistoryNode)
	node.completeLeaf(verificationPeriod, prefixTreeHash, m.Size)
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
}

/*
// GenerateExistenceProof generates an existence proof for a given key/height pair
func (m *HistoryForest) GenerateExistenceProof(key []byte, pos uint32, height uint32, oldSize uint32) *MerkleExistenceProof {

	node := m.getNode(0, pos)
	siblings := []Sibling{}
	prefixProofs := []MembershipProof{}
	treeDepth := GetOldDepth(pos, oldSize)
	prefix := makePrefixFromKey(key)

	for node.getDepth() != treeDepth { // only thing that needs to change is m.depth to the depth of the tree that the key belongs to in new Size

		if height > 0 {
			height = height - 1
		} else {

			prefixTree, _ := node.getParent().getPrefixTree()
			proof, _ := prefixTree.generateMembershipProof(prefix)
			prefixProofs = append(prefixProofs, *proof)

			sibling := node.getSibling()
			siblings = append(siblings, sibling)
		}

		node = node.getParent()
	}

	return &MerkleExistenceProof{
		Siblings:     siblings,
		PrefixProofs: prefixProofs,
	}
}
*/

// GenerateExtensionProof generates an extension proof for a given digest
func (m *HistoryForest) GenerateExtensionProof(oldSize uint32, requestedSize uint32) *MerkleExtensionProof {

	roots := m.getOldRoots(requestedSize)
	oldDigestRoots := m.getOldRoots(oldSize)
	res := &MerkleExtensionProof{}

	for i, root := range oldDigestRoots {
		if !bytes.Equal(root.getHash(), roots[i].getHash()) { // if we want size param: switch m.roots for getRoots()

			lastNode := oldDigestRoots[len(oldDigestRoots)-1]

			generateHistoryForestExtensionProof(lastNode, res, roots[i].getDepth()) // if we want size param: need to pass in correct depth for the tree that the last element/root belongs to (olddigest.size) in proof version?
			break
		}
	}

	return res
}

func generateHistoryForestExtensionProof(node HistoryNode, proof *MerkleExtensionProof, depth uint32) {

	siblings := []Sibling{}
	prefixHashes := [][]byte{}

	for node.getDepth() != depth { // if we want size param: remove isComplete() and pass in correct depth

		if !node.isRightChild() {
			sibling := node.getSibling()
			siblings = append(siblings, sibling)
		}
		prefixHashes = append(prefixHashes, node.getParent().getNewestLeafHash())

		node = node.getParent()
	}

	proof.Siblings = siblings
	proof.PrefixHashes = prefixHashes
}

//*******************************
// HELPER METHODS
//*******************************

// Gets Roots of HistoryForest when it only contained oldSize keys
func (m *HistoryForest) getOldRoots(oldSize uint32) []HistoryNode {
	Roots := []HistoryNode{}
	var totalKeys uint32 = 0
	var mask uint32 = 1 << m.depth

	for mask > 0 {

		if bits.OnesCount32(mask&oldSize) == 1 {

			depth := bits.TrailingZeros32(mask)
			shift := totalKeys >> bits.TrailingZeros32(mask)

			Roots = append(Roots, m.getNode(uint32(depth), shift))

			totalKeys += mask
		}
		mask = mask >> 1
	}

	return Roots
}

// Adds a HistoryNode to the HistoryForest hashmap
func (m *HistoryForest) addNodeToMap(node HistoryNode) {

	// sync map version
	//m.nodeMap.Store(node.getIndex(), node)

	// old version
	//nodeMap[node.getIndex()] = node
}

// Gets a HistoryNode from an index
func (m *HistoryForest) getNodeFromIndex(index index) HistoryNode {

	node := m.root

	for node.getDepth() != index.depth {

		if isRightOf(node.getIndex(), index) {
			node = node.getRightChild()
		} else {
			node = node.getLeftChild()
		}

	}

	return node
}

func (m *HistoryForest) getNode(depth uint32, shift uint32) HistoryNode {
	index := index{
		depth: depth,
		shift: shift,
	}

	return m.getNodeFromIndex(index)
}

// Fetches a leaf node at a given position in the HistoryForest struct (log n)
func (m *HistoryForest) getLeafNode(pos uint32) HistoryNode {

	node := m.root

	for node.getDepth() > 0 {
		shift := node.getDepth() - 1

		if pos&(1<<shift)>>shift == 1 {
			node = node.getRightChild()
		} else {
			node = node.getLeftChild()
		}
	}

	return node
}

// GetDigest returns the most recent digest of a HistoryForest struct
func (m *HistoryForest) GetDigest() *Digest {
	Roots := [][]byte{}

	for _, root := range m.Roots {
		Roots = append(Roots, root.getHash())
	}

	return &Digest{
		Roots: Roots,
		Size:  m.Size,
	}
}

// GetOldDigest returns a digest of the a HistoryForest instance
// when it only contained oldSize keys.
func (m *HistoryForest) GetOldDigest(oldSize uint32) *Digest {
	Roots := [][]byte{}

	for _, root := range m.getOldRoots(oldSize) {
		Roots = append(Roots, root.getHash())
	}

	return &Digest{
		Roots: Roots,
		Size:  oldSize,
	}
}

// Pops an element from the forest and returns it
func (m *HistoryForest) pop() HistoryNode {

	numTrees := len(m.Roots)
	node := m.Roots[numTrees-1]
	m.Roots = m.Roots[:numTrees-1]

	return node
}

// Adds an element to the forest
func (m *HistoryForest) addRoot(node HistoryNode) {
	m.Roots = append(m.Roots, node)
}

func (m *HistoryForest) isFull() bool {
	return m.Size == 1<<m.depth
}

// NewHistoryForest is a factory method for constructing HistoryForest objects
func NewHistoryForest(depth uint32) *HistoryForest {
	m := &HistoryForest{
		Roots: []HistoryNode{},
		Size:  0,
		depth: depth,
	}

	next := createRootHistoryNode(depth)
	m.root = next
	m.addNodeToMap(next)

	for next.getDepth() > 0 {
		newChild := next.createLeftChild()
		m.addNodeToMap(newChild)
		next = next.getLeftChild()
	}

	m.next = next

	return m
}

// GetHistoryForestSize returns the # bytes HistoryForest object requires
func (m *HistoryForest) GetHistoryForestSize() int {

	// root + next pointers
	total := pointerSizeInBytes * 2

	// forest root pointers
	for range m.Roots {
		total += pointerSizeInBytes
	}

	// size and depth
	total += binary.Size(m.Size) + binary.Size(m.depth)

	// recursively find tree size
	total += m.root.getSize()

	return total
}
