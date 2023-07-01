package redblacktree

import (
	"fmt"
	"log"
	"strings"

	"golang.org/x/exp/constraints"
)

type color int

const (
	red color = iota
	black
)

type node[T constraints.Ordered] struct {
	parent   *node[T]
	children [2]*node[T]

	color color

	data T
}

func (nd *node[T]) getChildID() int {
	// Only defined for non-nil nodes.

	p := nd.parent

	if p == nil {
		return -1
	}

	if p.children[0] == nd {
		return 0
	}
	return 1
}

func (nd *node[T]) sibling() *node[T] {
	// Only defined for non-nil nodes.

	p := nd.parent

	if p == nil {
		return nil
	}

	return p.children[1-nd.getChildID()]
}

func (nd *node[T]) validateSortInvariant() (T, T, error) {

	maxVal := nd.data
	minVal := nd.data

	if nd.children[0] != nil {
		maxValLChild, minValLChild, err := nd.children[0].validateSortInvariant()
		if err != nil {
			var nilVal T
			return nilVal, nilVal, err
		}

		if nd.data < maxValLChild {
			var nilVal T
			return nilVal, nilVal, fmt.Errorf("node %v has smaller value than some node %v in its left subtree", nd.data, maxValLChild)
		}

		minVal = minValLChild
	}

	if nd.children[1] != nil {
		maxValRChild, minValRChild, err := nd.children[1].validateSortInvariant()
		if err != nil {
			var nilVal T
			return nilVal, nilVal, err
		}

		if nd.data > minValRChild {
			var nilVal T
			return nilVal, nilVal, fmt.Errorf("node %v has larger value than some node %v in its right subtree", nd.data, minValRChild)
		}

		maxVal = maxValRChild
	}

	return maxVal, minVal, nil
}

// returns max black depth, min black depth, error in case of black invar
func (nd *node[T]) validateSubtreeRespectsBlackColorInvariant() (int, int, error) {
	if nd == nil {
		return 1, 1, nil
	}

	maxBDLChild, minBDLChild, err := nd.children[0].validateSubtreeRespectsBlackColorInvariant()
	if err != nil {
		return 0, 0, err
	}

	maxBDRChild, minBDRChild, err := nd.children[1].validateSubtreeRespectsBlackColorInvariant()
	if err != nil {
		return 0, 0, err
	}

	maxBD := maxBDLChild
	if maxBDRChild > maxBDLChild {
		maxBD = maxBDRChild
	}

	minBD := minBDLChild
	if minBDRChild < minBDLChild {
		minBD = minBDRChild
	}

	if nd.color == black {
		maxBD += 1
		minBD += 1
	}

	if maxBD != minBD {
		return 0, 0, fmt.Errorf("not all paths under node %v have the same black depth, max depth = %d, min depth = %d", maxBD, minBD)
	}

	return maxBD, minBD, nil
}

func (nd *node[T]) validateSubtreeRespectsRedColorInvariant() error {

	if nd == nil {
		return nil
	}

	if nd.color == red {
		if nd.children[0] != nil && nd.children[0].color == red {
			return fmt.Errorf("node %v and its child %v are both colored red", nd.data, nd.children[0].data)
		}

		if nd.children[1] != nil && nd.children[1].color == red {
			return fmt.Errorf("node %v and its child %v are both colored red", nd.data, nd.children[1].data)
		}
	}

	err := nd.children[0].validateSubtreeRespectsRedColorInvariant()
	if err != nil {
		return err
	}

	err = nd.children[1].validateSubtreeRespectsRedColorInvariant()
	if err != nil {
		return err
	}

	return nil
}

type Tree[T constraints.Ordered] struct {
	root *node[T]
}

func (t *Tree[T]) Insert(obj T) {
	log.Printf("Attempting to insert : %v", obj)

	p, nd := t.findParentAndNode(obj)

	// No op. Object already exists.
	if nd != nil {
		return
	}

	// Create new node.
	nd = &node[T]{
		parent:   p,
		children: [2]*node[T]{nil, nil},
		data:     obj,
	}

	// Tree empty. Make the new node the root. And set its color to black.
	if p == nil {
		log.Printf("Inserting : %v as root", obj)
		t.root = nd
		nd.color = black
		return
	}

	log.Printf("Inserting : %v", obj)

	nd.color = red

	if obj > p.data {
		p.children[1] = nd
	} else {
		// if obj < p.data
		p.children[0] = nd
	}

	t.rebalance(nd)
}

func (t *Tree[T]) rebalance(nd *node[T]) {
	log.Printf("Rebalancing : %v", nd.data)

	// Only need to rebalance when node is red.
	if nd.color != red {
		log.Printf("Color is not red. No need to rebalance : %v", nd.data)
		return
	}

	p := nd.parent

	// Node is root. Just make it black.
	// It's always safe to make the root node black.
	if p == nil {
		nd.color = black
		log.Printf("%v is at root. Marking as black", nd.data)
		return
	}

	// Node is red, and parent is black. No op.
	if p.color == black {
		log.Printf("Parent's color is not red. No need to rebalance : %v", nd.data)
		return
	}

	log.Printf("Have a double red problem at %v. Need to change tree structure to meet the red-black invariants", nd.data)

	// Otherwise, we have a double red problem. Need to change
	// tree structure to get rid of the double red problem while
	// maintaining the red-black invariants.

	gp := p.parent

	// If there is no grandparent, then parent is the root node. Just make it black.
	// It's always safe to make the root node black.
	if gp == nil {
		log.Printf("No grandparent for %v. Marking parent as black", nd.data)
		p.color = black
		return
	}

	uncl := p.sibling()

	if uncl == nil || uncl.color == black {
		// nd and p are red.
		// gp and uncl are black.

		if nd.getChildID() == p.getChildID() {
			// grandparent, parent and node are already in straight line.
			p.color = black
			gp.color = red
			log.Printf("Marking parent and grandparent for %v as black and red respectively", nd.data)
			t.rotate(p)
		} else {
			// grandparent, parent and node form a triangle.
			// rotate the node so that they get into a straight line.
			t.rotate(nd)
			t.rebalance(p)
		}

	} else {
		// nd, p and uncl are red.
		// gp is black.
		p.color = black
		uncl.color = black
		gp.color = red
		log.Printf("Marking parent, uncle and grandparent for %v as black, black and red respectively", nd.data)
		t.rebalance(gp)
	}
}

func (t *Tree[T]) rotate(nd *node[T]) {
	// We'll assume that nd is some node in the tree.
	log.Printf("Rotating : %v", nd.data)

	cid := nd.getChildID()
	if cid == -1 {
		// Node is the root. No op.
		return
	}

	p := nd.parent
	gp := p.parent

	pcid := p.getChildID()
	switch pcid {
	case -1:
		// Parent is the root of the tree.
		t.root = nd
		nd.parent = nil
	case 0:
		// Parent is a left child of grandparent.
		gp.children[0] = nd
		nd.parent = gp
	case 1:
		// Parent is a right child of grandparent.
		gp.children[1] = nd
		nd.parent = gp
	}

	// If nd is a left child, then ndChild will be its right child.
	// If nd is a right child, then ndChild will be its left child.
	ndChild := nd.children[1-cid]

	nd.children[1-cid] = p
	p.parent = nd

	// nd used to be the cid'th child of p. Now, ndChild will be that.
	p.children[cid] = ndChild
	if ndChild != nil {
		ndChild.parent = p
	}
}

func (t *Tree[T]) Contains(obj T) bool {
	_, nd := t.findParentAndNode(obj)
	return nd != nil
}

func (t *Tree[T]) findParentAndNode(obj T) (*node[T], *node[T]) {

	nd := t.root
	var p *node[T]
	for nd != nil {
		if obj < nd.data {
			p = nd
			nd = nd.children[0]
		} else if obj > nd.data {
			p = nd
			nd = nd.children[1]
		} else {
			return p, nd
		}
	}

	return p, nil
}

func (t *Tree[T]) Print() {
	t.print(t.root, 0)
}

func (t *Tree[T]) print(nd *node[T], indent int) {

	if nd == nil {
		fmt.Printf("%s|NIL\n", strings.Repeat(" ", indent))
		return
	}

	var clr string
	if nd.color == black {
		clr = "B"
	} else {
		clr = "R"
	}
	fmt.Printf("%s|%v(%s)\n", strings.Repeat(" ", indent), nd.data, clr)
	t.print(nd.children[0], indent+2)
	t.print(nd.children[1], indent+2)
}

func (t *Tree[T]) CheckInvariants() error {

	if t.root != nil {
		_, _, err := t.root.validateSortInvariant()
		if err != nil {
			return err
		}
	}

	err := t.checkColorInvariants()
	if err != nil {
		return err
	}

	return nil
}

func (t *Tree[T]) checkColorInvariants() error {

	if t.root == nil {
		return nil
	}

	if t.root.color != black {
		return fmt.Errorf("root is not colored black")
	}

	err := t.root.validateSubtreeRespectsRedColorInvariant()
	if err != nil {
		return err
	}

	err = t.root.validateSubtreeRespectsRedColorInvariant()
	if err != nil {
		return err
	}

	return nil
}
