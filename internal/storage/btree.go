package storage

import (
	"fmt"
)

type BTreeNode struct {
	IsLeaf    bool
	Keys      []string
	Values    [][]byte
	Children  []*BTreeNode
	Parent    *BTreeNode
	MinDegree int
}

type BTree struct {
	Root      *BTreeNode
	MinDegree int
	Size      int
}

func NewBTree(minDegree int) *BTree {
	return &BTree{
		MinDegree: minDegree,
		Size:      0,
	}
}

func (bt *BTree) Insert(key string, value []byte) {
	if bt.Root == nil {
		bt.Root = &BTreeNode{
			IsLeaf:    true,
			Keys:      []string{key},
			Values:    [][]byte{value},
			MinDegree: bt.MinDegree,
		}
		bt.Size = 1
		return
	}

	if len(bt.Root.Keys) == 2*bt.MinDegree-1 {
		newRoot := &BTreeNode{
			IsLeaf:    false,
			Keys:      []string{},
			Values:    [][]byte{},
			Children:  []*BTreeNode{bt.Root},
			MinDegree: bt.MinDegree,
		}
		bt.Root.Parent = newRoot
		bt.splitChild(newRoot, 0)
		bt.Root = newRoot
	}

	bt.insertNonFull(bt.Root, key, value)
	bt.Size++
}

func (bt *BTree) insertNonFull(node *BTreeNode, key string, value []byte) {
	i := len(node.Keys) - 1

	if node.IsLeaf {
		node.Keys = append(node.Keys, "")
		node.Values = append(node.Values, nil)
		for i >= 0 && node.Keys[i] > key {
			node.Keys[i+1] = node.Keys[i]
			node.Values[i+1] = node.Values[i]
			i--
		}
		node.Keys[i+1] = key
		node.Values[i+1] = value
	} else {
		for i >= 0 && node.Keys[i] > key {
			i--
		}
		i++
		if len(node.Children[i].Keys) == 2*bt.MinDegree-1 {
			bt.splitChild(node, i)
			if node.Keys[i] < key {
				i++
			}
		}
		bt.insertNonFull(node.Children[i], key, value)
	}
}

func (bt *BTree) splitChild(parent *BTreeNode, index int) {
	minDegree := bt.MinDegree
	child := parent.Children[index]
	newNode := &BTreeNode{
		IsLeaf:    child.IsLeaf,
		Keys:      make([]string, minDegree-1),
		Values:    make([][]byte, minDegree-1),
		MinDegree: minDegree,
		Parent:    parent,
	}
	for i := 0; i < minDegree-1; i++ {
		newNode.Keys[i] = child.Keys[i+minDegree]
		newNode.Values[i] = child.Values[i+minDegree]
	}
	if !child.IsLeaf {
		newNode.Children = make([]*BTreeNode, minDegree)
		for i := 0; i < minDegree; i++ {
			newNode.Children[i] = child.Children[i+minDegree]
			newNode.Children[i].Parent = newNode
		}
	}
	child.Keys = child.Keys[:minDegree-1]
	child.Values = child.Values[:minDegree-1]
	if !child.IsLeaf {
		child.Children = child.Children[:minDegree]
	}
	parent.Keys = append(parent.Keys, "")
	parent.Values = append(parent.Values, nil)
	parent.Children = append(parent.Children, nil)
	for i := len(parent.Keys) - 1; i > index; i-- {
		parent.Keys[i] = parent.Keys[i-1]
		parent.Values[i] = parent.Values[i-1]
		parent.Children[i+1] = parent.Children[i]
	}
	parent.Keys[index] = child.Keys[minDegree-1]
	parent.Values[index] = child.Values[minDegree-1]
	parent.Children[index+1] = newNode
}

func (bt *BTree) Search(key string) ([]byte, bool) {
	if bt.Root == nil {
		return nil, false
	}
	return bt.searchNode(bt.Root, key)
}

func (bt *BTree) searchNode(node *BTreeNode, key string) ([]byte, bool) {
	i := 0
	for i < len(node.Keys) && key > node.Keys[i] {
		i++
	}
	if i < len(node.Keys) && key == node.Keys[i] {
		return node.Values[i], true
	}
	if node.IsLeaf {
		return nil, false
	}
	return bt.searchNode(node.Children[i], key)
}

func (bt *BTree) Delete(key string) bool {
	if bt.Root == nil {
		return false
	}
	found := bt.deleteFromNode(bt.Root, key)
	if found {
		bt.Size--
		if len(bt.Root.Keys) == 0 && !bt.Root.IsLeaf {
			bt.Root = bt.Root.Children[0]
			bt.Root.Parent = nil
		}
	}
	return found
}

func (bt *BTree) deleteFromNode(node *BTreeNode, key string) bool {
	i := 0
	for i < len(node.Keys) && key > node.Keys[i] {
		i++
	}
	if i < len(node.Keys) && key == node.Keys[i] {
		if node.IsLeaf {
			bt.deleteFromLeaf(node, i)
		} else {
			bt.deleteFromInternal(node, i)
		}
		return true
	}
	if node.IsLeaf {
		return false
	}
	child := node.Children[i]
	if len(child.Keys) < bt.MinDegree {
		bt.fillChild(node, i)
	}
	if i > len(node.Keys) {
		return bt.deleteFromNode(node.Children[i-1], key)
	}
	return bt.deleteFromNode(node.Children[i], key)
}

func (bt *BTree) deleteFromLeaf(node *BTreeNode, index int) {
	copy(node.Keys[index:], node.Keys[index+1:])
	copy(node.Values[index:], node.Values[index+1:])
	node.Keys = node.Keys[:len(node.Keys)-1]
	node.Values = node.Values[:len(node.Values)-1]
}

func (bt *BTree) deleteFromInternal(node *BTreeNode, index int) {
	key := node.Keys[index]
	if len(node.Children[index].Keys) >= bt.MinDegree {
		pred := bt.getPredecessor(node.Children[index])
		node.Keys[index] = pred.Keys[len(pred.Keys)-1]
		node.Values[index] = pred.Values[len(pred.Values)-1]
		bt.deleteFromNode(node.Children[index], pred.Keys[len(pred.Keys)-1])
		return
	}
	if len(node.Children[index+1].Keys) >= bt.MinDegree {
		succ := bt.getSuccessor(node.Children[index+1])
		node.Keys[index] = succ.Keys[0]
		node.Values[index] = succ.Values[0]
		bt.deleteFromNode(node.Children[index+1], succ.Keys[0])
		return
	}
	bt.mergeChildren(node, index)
	bt.deleteFromNode(node.Children[index], key)
}

func (bt *BTree) getPredecessor(node *BTreeNode) *BTreeNode {
	if node.IsLeaf {
		return node
	}
	return bt.getPredecessor(node.Children[len(node.Children)-1])
}

func (bt *BTree) getSuccessor(node *BTreeNode) *BTreeNode {
	if node.IsLeaf {
		return node
	}
	return bt.getSuccessor(node.Children[0])
}

func (bt *BTree) fillChild(parent *BTreeNode, index int) {
	minDegree := bt.MinDegree
	if index > 0 && len(parent.Children[index-1].Keys) >= minDegree {
		bt.borrowFromLeft(parent, index)
		return
	}
	if index < len(parent.Children)-1 && len(parent.Children[index+1].Keys) >= minDegree {
		bt.borrowFromRight(parent, index)
		return
	}
	if index > 0 {
		bt.mergeChildren(parent, index-1)
	} else {
		bt.mergeChildren(parent, index)
	}
}

func (bt *BTree) borrowFromLeft(parent *BTreeNode, index int) {
	child := parent.Children[index]
	leftSibling := parent.Children[index-1]
	child.Keys = append([]string{""}, child.Keys...)
	child.Values = append([][]byte{nil}, child.Values...)
	child.Keys[0] = parent.Keys[index-1]
	child.Values[0] = parent.Values[index-1]
	parent.Keys[index-1] = leftSibling.Keys[len(leftSibling.Keys)-1]
	parent.Values[index-1] = leftSibling.Values[len(leftSibling.Values)-1]
	leftSibling.Keys = leftSibling.Keys[:len(leftSibling.Keys)-1]
	leftSibling.Values = leftSibling.Values[:len(leftSibling.Values)-1]
	if !child.IsLeaf {
		child.Children = append([]*BTreeNode{nil}, child.Children...)
		child.Children[0] = leftSibling.Children[len(leftSibling.Children)-1]
		child.Children[0].Parent = child
		leftSibling.Children = leftSibling.Children[:len(leftSibling.Children)-1]
	}
}

func (bt *BTree) borrowFromRight(parent *BTreeNode, index int) {
	child := parent.Children[index]
	rightSibling := parent.Children[index+1]
	child.Keys = append(child.Keys, parent.Keys[index])
	child.Values = append(child.Values, parent.Values[index])
	parent.Keys[index] = rightSibling.Keys[0]
	parent.Values[index] = rightSibling.Values[0]
	copy(rightSibling.Keys, rightSibling.Keys[1:])
	copy(rightSibling.Values, rightSibling.Values[1:])
	rightSibling.Keys = rightSibling.Keys[:len(rightSibling.Keys)-1]
	rightSibling.Values = rightSibling.Values[:len(rightSibling.Values)-1]
	if !child.IsLeaf {
		child.Children = append(child.Children, rightSibling.Children[0])
		child.Children[len(child.Children)-1].Parent = child
		copy(rightSibling.Children, rightSibling.Children[1:])
		rightSibling.Children = rightSibling.Children[:len(rightSibling.Children)-1]
	}
}

func (bt *BTree) mergeChildren(parent *BTreeNode, index int) {
	child := parent.Children[index]
	sibling := parent.Children[index+1]
	child.Keys = append(child.Keys, parent.Keys[index])
	child.Values = append(child.Values, parent.Values[index])
	child.Keys = append(child.Keys, sibling.Keys...)
	child.Values = append(child.Values, sibling.Values...)
	if !child.IsLeaf {
		for _, c := range sibling.Children {
			c.Parent = child
		}
		child.Children = append(child.Children, sibling.Children...)
	}
	copy(parent.Keys[index:], parent.Keys[index+1:])
	copy(parent.Values[index:], parent.Values[index+1:])
	copy(parent.Children[index+1:], parent.Children[index+2:])
	parent.Keys = parent.Keys[:len(parent.Keys)-1]
	parent.Values = parent.Values[:len(parent.Values)-1]
	parent.Children = parent.Children[:len(parent.Children)-1]
}

func (bt *BTree) Range(start, end string) []KeyValue {
	var result []KeyValue
	if bt.Root != nil {
		bt.rangeFromNode(bt.Root, start, end, &result)
	}
	return result
}

func (bt *BTree) rangeFromNode(node *BTreeNode, start, end string, result *[]KeyValue) {
	i := 0
	for i < len(node.Keys) && node.Keys[i] < start {
		i++
	}
	if !node.IsLeaf {
		for j := 0; j <= i; j++ {
			bt.rangeFromNode(node.Children[j], start, end, result)
		}
	}
	for i < len(node.Keys) && node.Keys[i] <= end {
		*result = append(*result, KeyValue{
			Key:   node.Keys[i],
			Value: node.Values[i],
		})
		i++
	}
	if !node.IsLeaf {
		for j := i; j < len(node.Children); j++ {
			bt.rangeFromNode(node.Children[j], start, end, result)
		}
	}
}

func (bt *BTree) GetAll() []KeyValue {
	var result []KeyValue
	if bt.Root != nil {
		bt.getAllFromNode(bt.Root, &result)
	}
	return result
}

func (bt *BTree) getAllFromNode(node *BTreeNode, result *[]KeyValue) {
	if !node.IsLeaf {
		for i, child := range node.Children {
			bt.getAllFromNode(child, result)
			if i < len(node.Keys) {
				*result = append(*result, KeyValue{
					Key:   node.Keys[i],
					Value: node.Values[i],
				})
			}
		}
	} else {
		for i, key := range node.Keys {
			*result = append(*result, KeyValue{
				Key:   key,
				Value: node.Values[i],
			})
		}
	}
}

type KeyValue struct {
	Key   string
	Value []byte
}

func (bt *BTree) String() string {
	if bt.Root == nil {
		return "Empty B-Tree"
	}
	return bt.nodeString(bt.Root, 0)
}

func (bt *BTree) nodeString(node *BTreeNode, depth int) string {
	result := ""
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	result += fmt.Sprintf("%sKeys: %v\n", indent, node.Keys)
	if !node.IsLeaf {
		for i, child := range node.Children {
			result += fmt.Sprintf("%sChild %d:\n", indent, i)
			result += bt.nodeString(child, depth+1)
		}
	}
	return result
}