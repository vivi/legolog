package core

import (
	"bytes"
	"errors"
	"fmt"

	crypto "github.com/huyuncong/MerkleSquare/lib/crypto"
)

type metadata struct {
	hash       []byte
	values     []KeyHash
	epoch      uint64
	prefix     []byte
	prev       *metadata
	next       *metadata
	leftChild  *metadata
	rightChild *metadata
	parent     *metadata // The newest parent.
}

// Copies over all fields in current metadata, changing pointers appropriately.
func (m *metadata) makeNextMetadata(epoch uint64) *metadata {
	ret := &metadata{
		hash:       m.hash,
		values:     m.values,
		epoch:      epoch,
		prefix:     m.prefix,
		prev:       m,
		next:       nil,
		leftChild:  m.leftChild,
		rightChild: m.rightChild,
		parent:     m.parent,
	}
	if m.parent != nil {
		metadataToEdit := m.parent
		if epoch > m.parent.epoch {
			metadataToEdit = m.parent.makeNextMetadata(epoch)
		}
		if m.prefix[0] == 0 {
			metadataToEdit.leftChild = ret
		} else {
			metadataToEdit.rightChild = ret
		}
		ret.parent = metadataToEdit
	}
	if m.leftChild != nil {
		m.leftChild.parent = ret
	}
	if m.rightChild != nil {
		m.rightChild.parent = ret
	}
	m.next = ret
	return ret
}

func (m *metadata) getValues() []KeyHash {
	return m.values
}

type persistentPrefixTree struct {
	roots        []*metadata
	currRoot     *metadata
	currEpoch    uint64
	sizesAtEpoch []uint32
}

func NewPersistentPrefixTree() *persistentPrefixTree {
	tree := &persistentPrefixTree{
		roots:        make([]*metadata, 0),
		currRoot:     &metadata{},
		sizesAtEpoch: []uint32{0},
		currEpoch:    0,
	}
	return tree
}

func (p *persistentPrefixTree) getHash(epoch uint64) []byte {
	return p.getRootAtEpoch(epoch).hash
}

func (p *persistentPrefixTree) getSize(epoch uint64) uint32 {
	return p.sizesAtEpoch[epoch]
}

func (p *persistentPrefixTree) getRootAtEpoch(epoch uint64) *metadata {
	if epoch == p.currEpoch {
		return p.currRoot
	}
	return p.roots[epoch]
}

func (p *persistentPrefixTree) InsertAndPrefixify(key []byte, valHash []byte, pos uint32) {
	prefix := makePrefixFromKey(key)
	p.Insert(prefix, valHash, pos)
}

// TODO: compressed prefix matching
func (p *persistentPrefixTree) Insert(prefix []byte, valHash []byte, pos uint32) {
	var prev *metadata
	var curr = p.currRoot

	i := 0
	for i < len(prefix) {
		prev = curr
		currBit := prefix[i]
		if currBit == 0 {
			curr = curr.leftChild
		} else {
			curr = curr.rightChild
		}

		if curr == nil {
			md := &metadata{
				hash:       nil,
				values:     []KeyHash{{valHash, pos}},
				epoch:      p.currEpoch,
				prefix:     prefix[i:],
				prev:       nil,
				next:       nil,
				leftChild:  nil,
				rightChild: nil,
				parent:     prev,
			}
			if currBit == 0 {
				prev.updateLeftChild(md, p.currEpoch)
			} else {
				prev.updateRightChild(md, p.currEpoch)
			}
			p.updateHashesFromLeaf(md)
			p.sizesAtEpoch[p.currEpoch] += 1
			return
		}

		j := uint32(0)
		for j < uint32(len(curr.prefix)) {
			if prefix[i] == curr.prefix[j] {
				i++
				j++
			} else {
				newParent := p.splitNode(curr, j)
				md := &metadata{
					hash:       nil,
					values:     []KeyHash{{valHash, pos}},
					epoch:      p.currEpoch,
					prefix:     prefix[i:],
					prev:       nil,
					next:       nil,
					leftChild:  nil,
					rightChild: nil,
					parent:     newParent,
				}
				if prefix[i] == 0 {
					newParent.leftChild = md
				} else {
					newParent.rightChild = md
				}
				p.updateHashesFromLeaf(md)
				p.sizesAtEpoch[p.currEpoch] += 1
				return
			}
		}
	}
	leaf := curr
	if leaf.epoch < p.currEpoch {
		m := leaf.makeNextMetadata(p.currEpoch)
		m.values = append(m.values, KeyHash{valHash, pos})
	} else {
		leaf.values = append(leaf.values, KeyHash{valHash, pos})
	}
	p.updateHashesFromLeaf(leaf)
}

func (p *persistentPrefixTree) splitNode(node *metadata, idx uint32) *metadata {
	parent := node.parent
	topHalf := &metadata{
		hash:       nil,
		values:     nil, // top half is never a leaf
		epoch:      p.currEpoch,
		prefix:     node.prefix[:idx],
		prev:       nil,
		next:       nil,
		leftChild:  nil,
		rightChild: nil,
		parent:     parent,
	}
	botHalf := &metadata{
		hash:       nil,
		values:     node.values, // in case bottom half was a leaf
		epoch:      p.currEpoch,
		prefix:     node.prefix[idx:],
		prev:       nil,
		next:       nil,
		leftChild:  node.leftChild,
		rightChild: node.rightChild,
		parent:     topHalf,
	}

	// update metadata parents for leftChild and rightChild
	if botHalf.leftChild != nil {
		if botHalf.leftChild.epoch < p.currEpoch {
			m := botHalf.leftChild.makeNextMetadata(p.currEpoch)
			m.parent = botHalf
			botHalf.leftChild = m
		} else {
			botHalf.leftChild.parent = botHalf
		}
	}
	if botHalf.rightChild != nil {
		if botHalf.rightChild.epoch < p.currEpoch {
			m := botHalf.rightChild.makeNextMetadata(p.currEpoch)
			m.parent = botHalf
			botHalf.rightChild = m
		} else {
			botHalf.rightChild.parent = botHalf
		}
	}

	botHalf.updateHash(p.currEpoch)
	if botHalf.prefix[0] == 0 {
		topHalf.leftChild = botHalf
	} else {
		topHalf.rightChild = botHalf
	}

	m := parent
	if parent.epoch < p.currEpoch {
		// m := &metadata{
		// 	hash:        parent.metadata.hash,
		// 	values:      parent.metadata.values,
		// 	epoch:       p.currEpoch,
		// 	prefix:      parent.metadata.prefix,
		// 	prev:        parent.metadata,
		// 	next:        nil,
		// 	leftChild:   parent.metadata.leftChild,
		// 	rightChild:  parent.metadata.rightChild,
		// 	nodePointer: parent.metadata.nodePointer,
		// }
		m = parent.makeNextMetadata(p.currEpoch)
	}
	if node.prefix[0] == 0 {
		m.leftChild = topHalf
	} else {
		m.rightChild = topHalf
	}
	topHalf.parent = m
	return topHalf
}

func (node *metadata) updateLeftChild(child *metadata, currEpoch uint64) {
	if currEpoch > node.epoch {
		m := node.makeNextMetadata(currEpoch)
		m.leftChild = child
	} else {
		node.leftChild = child
	}
}

func (node *metadata) updateRightChild(child *metadata, currEpoch uint64) {
	if currEpoch > node.epoch {
		m := node.makeNextMetadata(currEpoch)
		m.rightChild = child
	} else {
		node.rightChild = child
	}
}

func (p *persistentPrefixTree) updateHashesFromLeaf(leaf *metadata) {
	curr := leaf
	for curr.parent != nil {
		curr.updateHash(p.currEpoch)
		curr = curr.parent
	}
	curr.updateHash(p.currEpoch)
	//fmt.Printf("epoch %d, curr.epoch: %d, roothash: %x\n", p.currEpoch, curr.epoch, curr.hash[:10])
	p.currRoot = curr
	if curr.next != nil {
		p.currRoot = curr.next
	}
	//fmt.Printf("\troot:%x\n", p.currRoot.hash[:10])
}

func (node *metadata) updateHash(currEpoch uint64) {
	var leftHash, rightHash []byte
	if node.leftChild != nil {
		leftHash = node.leftChild.hash
	}
	if node.rightChild != nil {
		rightHash = node.rightChild.hash
	}
	var h []byte
	if node.leftChild == nil && node.rightChild == nil {
		h = leafHash(node.prefix, node.values) // update for leaf node
	} else {
		h = crypto.Hash(node.prefix, leftHash, rightHash) // update for internal node
	}
	if bytes.Compare(h, node.hash) != 0 {
		if node.epoch < currEpoch {
			m := node.makeNextMetadata(currEpoch)
			m.hash = h
		} else {
			node.hash = h
		}
	}
}

func (p *persistentPrefixTree) GetLeaf(prefix []byte, epoch uint64) *metadata {
	ret, err := p.LookupPath(prefix, epoch)
	if err != nil {
		return nil
	}
	return ret[len(ret)-1]
}

func (p *persistentPrefixTree) LookupPath(prefix []byte, epoch uint64) ([]*metadata, error) {
	if epoch > p.currEpoch {
		return nil, errors.New("epoch hasn't occurred yet")
	}

	var ret []*metadata = nil
	var curr *metadata = p.getRootAtEpoch(epoch)
	i := uint32(0)
	for i < uint32(len(prefix)) {
		if curr.epoch > epoch {
			panic("unreachable?")
		}
		if prefix[i] == 0 {
			if curr.leftChild == nil {
				return nil, fmt.Errorf("key doesn't exist in tree at epoch %d", epoch)
			}
			curr = curr.leftChild
		} else {
			if curr.rightChild == nil {
				return nil, fmt.Errorf("key doesn't exist in tree at epoch %d", epoch)
			}
			curr = curr.rightChild
		}
		if curr == nil {
			return nil, fmt.Errorf("key doesn't exist in tree at epoch %d", epoch)
		}

		partialPrefix := curr.prefix
		if bytes.Compare(prefix[i:i+uint32(len(partialPrefix))], partialPrefix) == 0 {
			ret = append(ret, curr)
			i += uint32(len(partialPrefix))
			continue
		} else {
			return nil, fmt.Errorf("key doesn't exist in tree at epoch %d", epoch)
		}
	}
	return ret, nil
}

func (p *persistentPrefixTree) NextEpoch() {
	// Apply the updates as one big path copy operation. -- maybe it's ok, we just only create new metadata when the epoch is different.
	p.sizesAtEpoch = append(p.sizesAtEpoch, p.sizesAtEpoch[p.currEpoch])
	p.roots = append(p.roots, p.currRoot)
	p.currEpoch += 1
}

func (p *persistentPrefixTree) StringAtEpoch(epoch uint64) string {
	currLevel := []*metadata{p.getRootAtEpoch(epoch)}
	currLevelParentPrefix := [][]byte{nil}
	ret := fmt.Sprintf("Tree at epoch %d:\n", epoch)
	level := 0
	for len(currLevel) > 0 {
		ret += fmt.Sprintf("Level %d: ", level)
		nextLevel := []*metadata{}
		nextLevelParentPrefix := [][]byte{}
		for i, node := range currLevel {
			m := node
			if m.parent != nil {
				//ret += fmt.Sprintf("%d<-", node.parent.fullPrefix(epoch))
				ret += fmt.Sprintf("%v<-", currLevelParentPrefix[i])
			}
			ret += fmt.Sprintf("%v\t", m)
			if m.leftChild != nil {
				nextLevel = append(nextLevel, m.leftChild)
				nextLevelParentPrefix = append(nextLevelParentPrefix, m.prefix)
			}
			if m.rightChild != nil {
				nextLevel = append(nextLevel, m.rightChild)
				nextLevelParentPrefix = append(nextLevelParentPrefix, m.prefix)
			}
		}
		ret += "\n"
		currLevel = nextLevel
		currLevelParentPrefix = nextLevelParentPrefix
		level += 1
	}
	return ret
}

func (m *metadata) String() string {
	return fmt.Sprintf("(%d, %x)", m.epoch, m.hash[:10])
	return fmt.Sprintf("(%v, %d, %x)", m.prefix, m.epoch, m.hash[:10])
}

// func (m *metadata) String() string {
// 	ret := ""
// 	curr := m
// 	for curr != nil {
// 		ret += fmt.Sprintf("(%v, %d, %v) -> ", curr.prefix, curr.epoch, curr.hash[:5])
// 		curr = curr.prev
// 	}
// 	return ret
// }

// PROOF GENERATION //

// type persistentMerkleProof struct {
// 	endNodeHash          []byte
// 	endNodePartialPrefix []byte
// 	copath               []forNodeOnCopath
// 	epoch 				 uint64
// }

// basically the same as LookupPath minus the ancestors of the leaf
func (tree *persistentPrefixTree) getLeaf(prefix []byte, epoch uint64) *metadata {
	var curr *metadata = tree.getRootAtEpoch(epoch)
	i := uint32(0)

	for i < uint32(len(prefix)) {
		if curr.epoch > epoch {
			panic("unreachable?")
		}
		if prefix[i] == 0 {
			if curr.leftChild == nil {
				return nil
			}
			curr = curr.leftChild
		} else {
			if curr.rightChild == nil {
				return nil
			}
			curr = curr.rightChild
		}
		if curr == nil {
			return nil
		}
		partialPrefix := curr.prefix
		//fmt.Printf("%x\n%x \n", prefix[i:i+uint32(len(partialPrefix))], partialPrefix)
		if bytes.Compare(prefix[i:i+uint32(len(partialPrefix))], partialPrefix) == 0 {
			i += uint32(len(partialPrefix))
			continue
		} else {
			return nil // leaf doesn't exist in tree at this epoch
		}
	}
	return curr

}

func (tree *persistentPrefixTree) generateMembershipProof(prefix []byte, epoch uint64) (proof *MembershipProof, leafValues []KeyHash) {
	var leaf *metadata = tree.getLeaf(prefix, epoch)
	if leaf == nil {
		return nil, nil
	}

	if leaf.leftChild != nil || leaf.rightChild != nil {
		panic("prefix path should end with a leaf, but does not")
	}
	copath := tree.buildCopathFromNodeFromRoot(leaf, prefix, epoch)
	return &MembershipProof{
		LeafPartialPrefix: leaf.prefix,
		CopathNodes:       copath,
	}, leaf.values
}

func (tree *persistentPrefixTree) generateNonMembershipProof(prefix []byte, epoch uint64) *NonMembershipProof {

	var prev *metadata
	var curr *metadata = tree.getRootAtEpoch(epoch)
	i := uint32(0)

	var conflictingPrefix []byte = nil

	var missingNodeIsLeftChild bool
	for i < uint32(len(prefix)) {
		if curr.epoch > epoch {
			panic("unreachable?")
		}

		prev = curr
		if prefix[i] == 0 {
			if curr.leftChild == nil {
				curr = nil
				missingNodeIsLeftChild = true
			} else {
				curr = curr.leftChild
			}
		} else {
			if curr.rightChild == nil {
				curr = nil
				missingNodeIsLeftChild = false
			} else {
				curr = curr.rightChild
			}
		}

		if curr == nil {
			if prev != tree.getRootAtEpoch(epoch) {
				panic("root should be the only internal node that can have <2 children")
			}
			missingMetadata := &metadata{
				hash:       nil,
				epoch:      epoch,
				prefix:     []byte{prefix[i]},
				prev:       nil,
				next:       nil,
				leftChild:  nil,
				rightChild: nil,
				parent:     tree.getRootAtEpoch(epoch),
			}
			return &NonMembershipProof{
				EndNodeHash:          nil,
				EndNodePartialPrefix: missingMetadata.prefix,
				CopathNodes:          tree.buildCopathFromMissingNode(missingMetadata, epoch, missingNodeIsLeftChild),
			}
		}
		partialPrefix := curr.prefix
		conflictingPrefix = append(conflictingPrefix, partialPrefix...)
		if bytes.Compare(prefix[i:i+uint32(len(partialPrefix))], partialPrefix) == 0 {
			i += uint32(len(partialPrefix))
			continue
		} else {
			// WARNING: Again, this seems to miss something.
			// At least, you should use "i" or "j" information and recalculate the hash
			// as if there is a splited node at "i" or "j".
			// TODO: discuss with Yuncong again, I don't believe faking a split is necessary
			return &NonMembershipProof{
				EndNodeHash:          curr.hash,
				EndNodePartialPrefix: curr.prefix,
				CopathNodes:          tree.buildCopathFromNodeFromRoot(curr, conflictingPrefix, epoch),
			}
		}
	}
	return nil //key exists
}

func (p *persistentPrefixTree) buildCopathFromMissingNode(startingNodeMeta *metadata, epoch uint64, startingNodeIsLeftChild bool) []forNodeOnCopath {
	copath := []forNodeOnCopath{}
	if startingNodeMeta.parent == nil {
		return copath
	}

	parentMeta := startingNodeMeta.parent
	var siblingHash []byte
	if startingNodeIsLeftChild {
		if parentMeta.rightChild == nil {
			siblingHash = nil
		} else {
			siblingHash = parentMeta.rightChild.hash
		}
	} else {
		if parentMeta.leftChild == nil {
			siblingHash = nil
		} else {
			siblingHash = parentMeta.leftChild.hash
		}
	}
	copath = append(copath,
		forNodeOnCopath{
			PartialPrefix:  parentMeta.prefix,
			OtherChildHash: siblingHash,
		})

	copath = append(copath, p.buildCopathFromNodeFromRoot(parentMeta, nil, epoch)...)
	return copath
}

func (p *persistentPrefixTree) getPath(leaf *metadata, fullPrefix []byte, epoch uint64) []*metadata {
	curr := p.getRootAtEpoch(epoch)
	i := 0
	ret := []*metadata{}
	for curr != leaf {
		ret = append(ret, curr)
		var leftPrefix, rightPrefix []byte
		if curr.leftChild != nil {
			leftPrefix = curr.leftChild.prefix
			if (len(fullPrefix)-i) >= len(leftPrefix) && bytes.Equal(fullPrefix[i:i+len(leftPrefix)], leftPrefix) {
				curr = curr.leftChild
				i += len(leftPrefix)
				continue
			}
		}
		if curr.rightChild != nil {
			rightPrefix = curr.rightChild.prefix
			if (len(fullPrefix)-i) >= len(rightPrefix) && bytes.Equal(fullPrefix[i:i+len(rightPrefix)], rightPrefix) {
				curr = curr.rightChild
				i += len(rightPrefix)
				continue
			}
		}
		// If we reached here, then there was no path.
		return nil
	}
	ret = append(ret, leaf)
	return ret
}

func (p *persistentPrefixTree) buildCopathFromNodeFromRoot(leaf *metadata, fullPrefix []byte, epoch uint64) []forNodeOnCopath {
	path := p.getPath(leaf, fullPrefix, epoch)
	copath := []forNodeOnCopath{}
	curr := p.getRootAtEpoch(epoch)
	for _, pathNode := range path[1:] {
		var siblingHash []byte = nil
		if pathNode == curr.leftChild {
			if curr.rightChild != nil {
				siblingHash = curr.rightChild.hash
			}
		} else {
			if curr.leftChild != nil {
				siblingHash = curr.leftChild.hash
			}
		}
		copath = append([]forNodeOnCopath{
			{
				PartialPrefix:  curr.prefix,
				OtherChildHash: siblingHash,
			},
		}, copath...) // Append in reverse order as it expects leaf --> root.
		curr = pathNode
	}
	if curr != leaf {
		panic("curr should end at leaf")
	}
	return copath
}

// PROOF VERIFICATION //

// should work with existing methods in prefix_tree.go
