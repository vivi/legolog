package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"

	crypto "github.com/huyuncong/MerkleSquare/lib/crypto"
)

// internal node in (chronological) Merkle Tree
type internalChronNode struct {
	parent     ChronNode
	leftChild  ChronNode
	rightChild ChronNode
	hash       []byte
	depth      uint32
	id         uint32
}

// Leaf node in chronological MT
type leafChronNode struct {
	parent ChronNode
	hash   []byte
	id     uint32
}

// ChronNode interface for leaf/internal nodes
type ChronNode interface {
	isLeafNode() bool
	getHash() []byte
	updateHash()
	getParent() ChronNode
	setParent(ChronNode)
	getLeftChild() ChronNode
	getRightChild() ChronNode
	setRightChild(ChronNode)
	// TODO: implement serialize() ([]byte, error)
	getDepth() uint32
	getSize() int
	getId() uint32
}

// **********************************
// INTERFACE METHODS
// **********************************
func (node *internalChronNode) isLeafNode() bool {
	return false
}

func (node *internalChronNode) getHash() []byte {
	return node.hash
}

func (node *internalChronNode) updateHash() {
	var leftHash, rightHash []byte
	if node.leftChild != nil {
		leftHash = node.leftChild.getHash()
	}
	if node.rightChild != nil {
		rightHash = node.rightChild.getHash()
	}
	node.hash = crypto.Hash(leftHash, rightHash)
}

func (node *internalChronNode) getParent() ChronNode {
	return node.parent
}

func (node *internalChronNode) setParent(parent ChronNode) {
	node.parent = parent
}

func (node *internalChronNode) getLeftChild() ChronNode {
	return node.leftChild
}

func (node *internalChronNode) getRightChild() ChronNode {
	return node.rightChild
}

func (node *internalChronNode) setRightChild(c ChronNode) {
	node.rightChild = c
}

func (node *internalChronNode) getDepth() uint32 {
	return node.depth
}

func (node *internalChronNode) getSize() int {
	total := pointerSizeInBytes
	total += binary.Size(node.getHash())
	return total
}

func (node *internalChronNode) getId() uint32 {
	return node.id
}

func (node *leafChronNode) isLeafNode() bool {
	return true
}

func (node *leafChronNode) getHash() []byte {
	return node.hash
}

func (node *leafChronNode) updateHash() {}

func (node *leafChronNode) getParent() ChronNode {
	return node.parent
}

func (node *leafChronNode) setParent(parent ChronNode) {
	node.parent = parent
}

func (node *leafChronNode) getLeftChild() ChronNode {
	return nil
}

func (node *leafChronNode) getRightChild() ChronNode {
	return nil
}

func (node *leafChronNode) setRightChild(_ ChronNode) {}

func (node *leafChronNode) getDepth() uint32 {
	return 0
}

func (node *leafChronNode) getSize() int {
	total := pointerSizeInBytes
	total += binary.Size(node.getHash())
	return total
}

func (node *leafChronNode) getId() uint32 {
	return node.id
}

type ChronTree struct {
	root     ChronNode
	last     ChronNode
	numNodes uint32
	maxNodes uint32
}

func NewChronTree() *ChronTree {
	res := &ChronTree{
		root:     nil,
		last:     nil,
		numNodes: 0,
		maxNodes: 1,
	}
	return res
}

func (c *ChronTree) GetRootHash() []byte {
	if c.root == nil {
		return crypto.Hash()
	}
	return c.root.getHash()
}

func (c *ChronTree) Append(key []byte, value []byte, signature []byte) {
	contentHash := ComputeContentHash(key, value, signature, c.numNodes)
	hashVal := crypto.Hash(makePrefixFromKey(key), contentHash)
	leaf := leafChronNode{
		hash: hashVal,
		id:   c.numNodes + 1,
	}

	// First insert
	if c.numNodes == 0 {
		c.root = &leaf
		c.last = &leaf
		c.numNodes += 1
		return
	}

	// Tree is full, need a new root
	if c.maxNodes == c.numNodes {
		newRoot := internalChronNode{
			leftChild:  c.root,
			rightChild: &leaf,
			depth:      c.root.getDepth() + 1,
		}
		c.root.setParent(&newRoot)
		leaf.setParent(&newRoot)
		c.root = &newRoot
		c.maxNodes *= 2
		c.root.updateHash()
		c.last = &leaf
		c.numNodes += 1
		return
	}

	// Find the first parent node not at a proper depth to push down
	curr := c.last

	for curr != c.root && curr.getDepth()+1 == curr.getParent().getDepth() {
		curr = curr.getParent()
	}

	// Complete the tree
	if curr == c.root {
		parent := c.last.getParent()
		if parent.getRightChild() == nil {
			parent.setRightChild(&leaf)
			leaf.setParent(parent)
			c.last = &leaf
			c.numNodes += 1
		}
		// Propagate new hashes up the path
		for curr != nil {
			curr.updateHash()
			curr = curr.getParent()
		}
		return
	}

	// Insert
	grandparent := curr.getParent()
	parent := internalChronNode{
		leftChild:  curr,
		rightChild: &leaf,
		depth:      curr.getDepth() + 1,
	}
	grandparent.setRightChild(&parent)
	curr.setParent(&parent)
	leaf.setParent(&parent)
	parent.setParent(grandparent)

	curr = &parent

	// Propagate new hashes up the path
	for curr != nil {
		curr.updateHash()
		curr = curr.getParent()
	}

	c.last = &leaf
	c.numNodes += 1
}

func (c *ChronTree) GenerateConsistencyProof(oldSize uint32, requestedSize uint32) *MerkleExtensionProof {
	proof := MerkleExtensionProof{}
	c.generateConsistencyProof(oldSize, requestedSize, c.root, true, &proof)
	return &proof
}

func (c *ChronTree) generateConsistencyProof(m uint32, n uint32, currNode ChronNode, isComplete bool, proof *MerkleExtensionProof) {
	if m == 0 { // Anything can extend an empty tree, no proof needed.
		return
	}
	if m == n {
		if !isComplete {
			proof.PrefixHashes = append(proof.PrefixHashes, currNode.getHash())
		}
		return
	}

	k := uint32(math.Pow(2, math.Floor(math.Log2(float64(n)))))
	if k == n {
		k /= 2
	}
	if m <= k {
		c.generateConsistencyProof(m, k, currNode.getLeftChild(), isComplete, proof)
		proof.PrefixHashes = append(proof.PrefixHashes, currNode.getRightChild().getHash())
	}
	if m > k {
		c.generateConsistencyProof(m-k, n-k, currNode.getRightChild(), false, proof) // TODO: Is n-k right here?
		proof.PrefixHashes = append(proof.PrefixHashes, currNode.getLeftChild().getHash())
	}
}

func VerifyConsistencyProof(oldDigest *Digest, newDigest *Digest, proof *MerkleExtensionProof) bool {
	if oldDigest.Size == 0 { // Anything can extend an empty tree
		return true
	}
	if oldDigest.Size == newDigest.Size {
		return bytes.Equal(oldDigest.Roots[0], newDigest.Roots[0])
	}
	if len(proof.PrefixHashes) == 0 {
		return false
	}

	oldSize := oldDigest.Size
	oldHash := oldDigest.Roots[0]
	newSize := newDigest.Size
	newHash := newDigest.Roots[0]

	// If oldSize is a power of 2
	if (oldSize & (oldSize - 1)) == 0 {
		proof.PrefixHashes = append([][]byte{oldHash}, proof.PrefixHashes...)
	}

	fn := oldSize - 1
	sn := newSize - 1

	for (fn & 1) == 1 {
		fn = fn >> 1
		sn = sn >> 1
	}

	fr := proof.PrefixHashes[0]
	sr := proof.PrefixHashes[0]

	for _, c := range proof.PrefixHashes[1:] {
		if sn == 0 {
			return false
		}
		if (fn&1) == 1 || fn == sn {
			fr = crypto.Hash(c, fr)
			sr = crypto.Hash(c, sr)
			for (fn&1) == 0 && fn != 0 {
				fn = fn >> 1
				sn = sn >> 1
			}
		} else {
			sr = crypto.Hash(sr, c)
		}
		fn = fn >> 1
		sn = sn >> 1
	}

	return bytes.Equal(fr, oldHash) && bytes.Equal(sr, newHash) && sn == 0
}

func (c *ChronTree) PrintTree() {
	currLevel := []ChronNode{c.root}
	for i := 0; i <= int(c.root.getDepth()); i++ {
		nextLevel := []ChronNode{}
		depthStr := ""
		for _, node := range currLevel {
			if node.isLeafNode() {
				depthStr += fmt.Sprintf("%d\t", node.getId())
			} else {
				depthStr += fmt.Sprintf("%x\t", node.getHash()[:4])
			}
			if node.getLeftChild() != nil {
				nextLevel = append(nextLevel, node.getLeftChild())
			}
			if node.getRightChild() != nil {
				nextLevel = append(nextLevel, node.getRightChild())
			}
		}
		fmt.Println(depthStr)
		currLevel = nextLevel
	}
}
