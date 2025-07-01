package core

func (tree *prefixTree) copyFast() (*prefixTree, error) {
	// print number of nodes in tree
	// fmt.Println("number of nodes in tree: ", tree.root.getNumNodes())
	ret := &prefixTree{
		root: &internalNode{
			parent:        nil,
			hash:          []byte(tree.root.hash),
			leftChild:     nil,
			rightChild:    nil,
			partialPrefix: []byte(tree.root.partialPrefix),
		},
		isComplete: tree.isComplete,
		appends:    make([]prefixAppend, len(tree.appends)),
	}
	for i := 0; i < len(tree.appends); i++ {
		ret.appends[i] = prefixAppend{
			Prefix: tree.appends[i].Prefix,
			Value:  tree.appends[i].Value,
			Pos:    tree.appends[i].Pos,
		}
	}

	leftChild, err := deepCopyInternalNodeIterative(ret.root, tree.root.leftChild)
	// weird issue where if we set ret.root.leftChild = nil, then it says
	// a data property inside that child to nil
	if leftChild != nil {
		ret.root.leftChild = leftChild
	}
	if err != nil {
		return nil, err
	}
	rightChild, err := deepCopyInternalNodeIterative(ret.root, tree.root.rightChild)
	if rightChild != nil {
		ret.root.rightChild = rightChild
	}
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func deepCopyInternalNode(parent prefixNode, node prefixNode) (prefixNode, error) {
	if node == nil {
		return nil, nil
	}
	if node.isLeafNode() {
		ret := &leafNode{
			parent:        parent,
			hash:          []byte(node.getHash()),
			values:        make([]KeyHash, 0, len(node.getValues())),
			partialPrefix: []byte(node.getPartialPrefix()),
		}
		// fmt.Printf("num values: %d, size of values hash: %d\n", len(node.getValues()), len(node.getValues()[0].Hash))
		ret.values = append(ret.values, node.getValues()...)
		for i := 0; i < len(ret.values); i++ {
			ret.values[i] = KeyHash{
				Hash: ret.values[i].Hash,
				Pos:  ret.values[i].Pos,
			}
		}
		return ret, nil
	}

	ret := &internalNode{
		parent:        parent,
		hash:          []byte(node.getHash()),
		leftChild:     nil,
		rightChild:    nil,
		partialPrefix: []byte(node.getPartialPrefix()),
	}

	leftChild, err := deepCopyInternalNode(ret, node.getLeftChild())
	if leftChild != nil {
		ret.leftChild = leftChild
	}
	if err != nil {
		return nil, err
	}
	rightChild, err := deepCopyInternalNode(ret, node.getRightChild())
	if rightChild != nil {
		ret.rightChild = rightChild
	}
	if err != nil {
		return nil, err
	}
	return ret, nil

}

func deepCopyInternalNodeIterative(parent prefixNode, node prefixNode) (prefixNode, error) {
	if node == nil {
		return nil, nil
	}
	if node.isLeafNode() {
		ret := &leafNode{
			parent:        parent,
			hash:          []byte(node.getHash()),
			values:        make([]KeyHash, 0, len(node.getValues())),
			partialPrefix: []byte(node.getPartialPrefix()),
		}
		ret.values = append(ret.values, node.getValues()...)
		for i := 0; i < len(ret.values); i++ {
			ret.values[i] = KeyHash{
				Hash: ret.values[i].Hash,
				Pos:  ret.values[i].Pos,
			}
		}
		return ret, nil
	}

	ret := &internalNode{
		parent:        parent,
		hash:          []byte(node.getHash()),
		leftChild:     nil,
		rightChild:    nil,
		partialPrefix: []byte(node.getPartialPrefix()),
	}
	clone := ret

	left := node
	height := 0
	for left != nil {
		left = left.getLeftChild()
		height += 1
	}
	// fmt.Println("height: ", height)

	for node != nil {
		//fmt.Println("node: ", node)
		// fmt.Println("node not nil yet")
		if node.getLeftChild() != nil && clone.getLeftChild() == nil {
			// create clone.left child
			if node.getLeftChild().isLeafNode() {
				//fmt.Println("at left leaf case")
				child := node.getLeftChild()
				clone.leftChild = cloneLeafNode(child, clone)

				// dummy internal node to mimic a leaf (hack)
				clone = &internalNode{
					leftChild:  nil,
					rightChild: nil,
					parent:     clone,
				}
				node = node.getLeftChild()
			} else {
				//fmt.Println("at left non-leaf case")
				child := node.getLeftChild()
				internalNode := &internalNode{
					parent:        clone,
					hash:          []byte(child.getHash()),
					leftChild:     nil,
					rightChild:    nil,
					partialPrefix: []byte(child.getPartialPrefix()),
				}
				clone.leftChild = internalNode

				clone = internalNode
				node = node.getLeftChild()
			}
		} else if node.getRightChild() != nil && clone.getRightChild() == nil {
			if node.getRightChild().isLeafNode() {
				//fmt.Println("at right leaf case")
				child := node.getRightChild()
				leafNode := cloneLeafNode(child, clone)
				clone.rightChild = leafNode

				// dummy internal node to mimic a leaf (hack)
				clone = &internalNode{
					leftChild:  nil,
					rightChild: nil,
					parent:     clone,
				}
				node = node.getRightChild()
			} else {
				//fmt.Println("at right non-leaf case")
				child := node.getRightChild()
				internalNode := &internalNode{
					parent:        clone,
					hash:          []byte(child.getHash()),
					leftChild:     nil,
					rightChild:    nil,
					partialPrefix: []byte(child.getPartialPrefix()),
				}
				clone.rightChild = internalNode

				clone = internalNode
				node = node.getRightChild()
			}
		} else {
			//fmt.Println("at parent case")
			node = node.getParent()
			//fmt.Println("node parent is ", node)
			//fmt.Println("clone parent is ", clone.getParent())
			if clone.getParent() == nil {
				clone = nil
			} else {
				clone = clone.getParent().(*internalNode)
			}
		}
	}

	/* ret := &internalNode{
		parent:        parent,
		hash:          []byte(node.getHash()),
		leftChild:     nil,
		rightChild:    nil,
		partialPrefix: []byte(node.getPartialPrefix()),
	} */
	return ret, nil

}

func cloneLeafNode(leaf prefixNode, parent prefixNode) *leafNode {
	newChild := &leafNode{
		parent:        parent,
		hash:          []byte(leaf.getHash()),
		values:        make([]KeyHash, 0, len(leaf.getValues())),
		partialPrefix: []byte(leaf.getPartialPrefix()),
	}
	newChild.values = append(newChild.values, leaf.getValues()...)
	return newChild
}
