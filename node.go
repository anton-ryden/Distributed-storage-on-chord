package main

import (
	"fmt"
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
	id          string

	Bucket map[Key]string
}

func (node Node) print() {
	fmt.Println("\n+-+-+-+-+-+- Node DETAILS +-+-+-+-+-+-+")
	fmt.Println("Adress: " + node.address)
	fmt.Println("ID: " + node.id)
	fmt.Println("Number of Successors: ", len(node.successor))
}

// create a new Chord ring.
func (node *Node) create() {
	node.predecessor = nil
	node.successor = append(node.successor, node)
}

// join a Chord ring containing node n′.
func (node *Node) join(n Node) {
	//node.predecessor = nil
	//joinNode := n.findSuccessor(node.id)
	//node.successor = append(node.successor, &joinNode)
	node.successor = append(node.successor, &n)
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
func (node Node) findSuccessor(id string) Node {

	if id == node.id {
		return node
	}

	if len(node.successor) > 0 || len(node.successor) > node.m {
		for _, suc := range node.successor {
			if id == suc.id {
				return *suc
			}
		}
	} else {
		return node
	}

	return node.closestPrecedingNode(id)
}

// search the local table for the highest predecessor of id
func (node Node) closestPrecedingNode(id string) Node {
	for i := node.m; i > 1; i-- {
		iNode := node.findSuccessor(id)
		if node.fingerTable[i] == &node || node.fingerTable[i] == &iNode {
			iFinger := node.fingerTable[i]
			return *iFinger
		}
	}
	return *node.successor[0]
}
