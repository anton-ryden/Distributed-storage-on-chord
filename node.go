package main

import (
	"log"
	"math"
	"net"
	"strconv"
)

type Key string

type NodeAddress string

type Node struct {
	address     NodeAddress
	fingerTable []*Node
	predecessor *Node
	successor   []*Node
	next        int
	m           int
	id          int

	Bucket map[Key]string
}

func (node Node) start() {
	// Check if valid address
	checkIPAddress(*a)

	// Start tcp server with port from argument
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(*p))
	defer ln.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

// create a new Chord ring.
func (node Node) create() {
	node.predecessor = nil
	node.successor = append(node.successor, &node)
}

// join a Chord ring containing node n′.
func (node Node) join(n Node) {
	node.predecessor = nil
	joinNode := n.findSuccessor(node.id)
	node.successor = append(node.successor, &joinNode)
}

// called periodically. verifies n’s immediate
// successor, and tells the successor about n.
func (node Node) stabilize() {
	suc := node.successor[0]
	x := suc.predecessor
	if x == &node || &node == suc {
		node.successor = append(node.successor, x)
	}
	suc.notify(node)

}

// n′ thinks it might be our predecessor.
func (node Node) notify(n Node) {
	if node.predecessor == nil || (&n == node.predecessor || n.address == node.address) {
		node.predecessor = &n
	}
}

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

// called periodically. checks whether predecessor has failed.
/*
func (node Node) checkPredecessor(){
	if (node.predecessor has failed){
		node.predecessor = nil;
	}
}
*/

// ask node n to find the successor of id
// or a better node to continue the search with
func (node Node) findSuccessor(id int) Node {
	suc := node.successor[0]
	if id == node.id || id == suc.id {
		return *suc
	}

	return node.closestPrecedingNode(id)
}

// search the local table for the highest predecessor of id
func (node Node) closestPrecedingNode(id int) Node {
	for i := node.m; i > 1; i-- {
		iNode := node.findSuccessor(id)
		if node.fingerTable[i] == &node || node.fingerTable[i] == &iNode {
			iFinger := node.fingerTable[i]
			return *iFinger
		}
	}
	return *node.successor[0]
}
