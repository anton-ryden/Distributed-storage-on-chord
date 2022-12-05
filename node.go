package main

import (
	"bytes"
	"fmt"
	"math/big"
)

type Key string

type NodeAddress string

type Node struct {
	Address     NodeAddress
	FingerTable []*Node
	Predecessor *Node
	Successor   []*Node
	Next        int
	R           int
	Id          []byte

	Bucket map[Key]string
}

func (node Node) print() {
	fmt.Println("\n+-+-+-+-+-+- Node DETAILS +-+-+-+-+-+-+")
	fmt.Println("Adress: " + node.Address)
	idInt := (&big.Int{}).SetBytes(node.Id)
	fmt.Printf("ID: %d\n", idInt)
	fmt.Println("Number of Successors: ", len(node.Successor))
}

// create a new Chord ring.
func (node *Node) create() {
	node.Predecessor = nil
	node.Successor = append(node.Successor, node)
}

// join a Chord ring containing node n′.
func (node *Node) join(joinNode Node) {
	node.Predecessor = nil
	//successors := joinNode.findSuccessor(joinNode.Id)
	//node.Successor = append(node.Successor, &successors)
	node.Successor = append(node.Successor, &joinNode)
}

func (node Node) find(id []byte, start Node) Node {
	found, nextNode := false, start
	for found == false {
		found, nextNode = start.findSuccessor(id)
	}

	if found == true {
		return nextNode
	} else {
		//return find(node.Successor[i])
		return Node{}
	}
}

// called periodically. verifies n’s immediate
// successor, and tells the successor about n.
func (node Node) stabilize() {
	suc := node.Successor[0]
	x := suc.Predecessor
	if x == &node || &node == suc {
		node.Successor = append(node.Successor, x)
	}
	suc.notify(node)
}

// n′ thinks it might be our predecessor.
func (node Node) notify(n Node) {
	if node.Predecessor == nil || (&n == node.Predecessor || n.Address == node.Address) {
		node.Predecessor = &n
	}
}

/*
// called periodically. refreshes finger table entries.
// next stores the index of the next finger to fix.
func (node Node) fixFingers() {
	node.next = node.next + 1
	if node.next > node.m {
		node.next = 1
	}

	calc := node.id + int(math.Pow(float64(2), float64(node.next-1)))
	suc := node.findSuccessor(calc)
	node.fingerTable[node.next] = &suc
}
/*
// called periodically. checks whether predecessor has failed.
func (node Node) checkPredecessor(){
	if (node.predecessor has failed){
		node.predecessor = nil;
	}
}
*/

// ask node n to find the successor of id
// or a better node to continue the search with
func (node Node) findSuccessor(id []byte) (bool, Node) {
	if bytes.Equal(id, node.Id) {
		return true, node
	}

	for _, suc := range node.Successor {
		if bytes.Equal(id, suc.Id) {
			return true, *suc
		}
	}
	return false, node.closestPrecedingNode(id)
}

// search the local table for the highest predecessor of id
func (node Node) closestPrecedingNode(id []byte) Node {
	for i := node.R; i > 1; i-- {
		_, iNode := node.findSuccessor(id)
		if node.FingerTable[i] == &node || node.FingerTable[i] == &iNode {
			iFinger := node.FingerTable[i]
			return *iFinger
		}
	}
	return *node.Successor[0]
}
