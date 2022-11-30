package main

import "math"

type Key string

type NodeAddress string

type Node struct {
	address     NodeAddress
	fingerTable []NodeAddress
	predecessor *Node
	successor  *Node
	next int

	Bucket map[Key]string
}

// create a new Chord ring.
func (node Node) create(){
	node.predecessor = nil
	node.successor = &node
}

// join a Chord ring containing node n′.
func (node Node) join(n Node){
	node.predecessor = nil
	node.successor = n.findSuccessor(node)
}

// called periodically. verifies n’s immediate
// successor, and tells the successor about n.
func (node Node) stabilize(){
	x := node.successor.predecessor
	if x == &node || &node == node.successor {
		node.successor = x
	}
	node.successor.notify(node)
}

// n′ thinks it might be our predecessor.
func (node Node) notify(n Node){
	if (node.predecessor == nil || (&n == node.predecessor || n.address == node.address)){
		node.predecessor = &n
	}

}

// called periodically. refreshes finger table entries.
// next stores the index of the next finger to fix.
func (node Node) fixFingers(){
	node.next = node.next + 1
	if (node.next > m)
		node.next = 1
	node.fingerTable[node.next] = findSuccessor(node + math.Pow(2, node.next − 1))
}

// called periodically. checks whether predecessor has failed.
func (node Node) checkPredecessor(){
	if (node.predecessor has failed){
		node.predecessor = nil;
	}
}

// ask node n to find the successor of id
// or a better node to continue the search with
func (node Node) find_successor(id NodeAddress){
	if (id ∈ (n, successor]){
		return true, successor
	}
	else{
		return false, closest_preceding_node(id)
	}
}


// search the local table for the highest predecessor of id
func (node Node) closestPrecedingNode(id){
	// skip this loop if you do not have finger tables implemented yet
	for i = m downto 1
		if (node.fingerTable[i] ∈ (n,id])
			return finger[i];
	return successor;
}


// find the successor of id
func find(id, start){

}


