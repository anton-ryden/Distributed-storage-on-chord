package main

import (
	"crypto/sha1"
	"fmt"
	"math/big"
	"strconv"
)

const keySize = sha1.Size * 9

var two = big.NewInt(2)
var hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(keySize), nil)

type Key string

type NodeAddress string

type Node struct {
	Address     NodeAddress
	FingerTable []*Node
	Predecessor *Node
	Successor   []*Node
	Next        int
	R           int
	Id          *big.Int

	Bucket map[Key]string
}

func newNode(ip string, port int, iArg string, r int) Node {
	// Error handling in arguments file, so we only need to check if ja is set
	addr := NodeAddress(ip + ":" + strconv.Itoa(port))

	if iArg == "" {
		iArg = ip + strconv.Itoa(port)
	}

	id := hash(iArg)

	id = new(big.Int).Mod(id, hashMod)
	fmt.Println("ID: ", id)

	return Node{Address: addr, R: r, Id: id}
}

func hash(ipPort string) *big.Int {
	h := sha1.New()
	return new(big.Int).SetBytes(h.Sum([]byte(ipPort)))
}

func (node Node) print() {
	fmt.Println("\n+-+-+-+-+-+- Node DETAILS +-+-+-+-+-+-+")
	fmt.Println("Adress: " + node.Address)
	fmt.Printf("ID: %d\n", node.Id)
	fmt.Println("Number of Successors: ", len(node.Successor))
}

// create a new Chord ring.
func (node *Node) create() {
	node.Predecessor = nil
	node.Successor = append(node.Successor, &Node{Address: node.Address, R: node.R, Id: node.Id})
}

// join a Chord ring containing node n′.
func (node *Node) join(joinNode Node) {
	node.Predecessor = nil
	//successors := joinNode.findSuccessor(joinNode.Id)
	//node.Successor = append(node.Successor, &successors)
	node.Successor = append(node.Successor, &joinNode)
}

func (node Node) find(id big.Int, start Node) Node {
	found, nextNode := false, start
	for found == false {
		found, nextNode = nextNode.findSuccessor(id)
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
func (node Node) findSuccessor(id big.Int) (bool, Node) {
	if id.Cmp(node.Id) == 0 {
		return true, node
	}

	for _, suc := range node.Successor {
		if id.Cmp(suc.Id) == 0 {
			return true, *suc
		}
	}
	return false, node.closestPrecedingNode(id)
}

// search the local table for the highest predecessor of id
func (node Node) closestPrecedingNode(id big.Int) Node {
	for i := node.R; i > 1; i-- {
		_, iNode := node.findSuccessor(id)
		if node.FingerTable[i] == &node || node.FingerTable[i] == &iNode {
			iFinger := node.FingerTable[i]
			return *iFinger
		}
	}
	return *node.Successor[0]
}

func (node Node) jump(fingerentry int) *big.Int {
	n := hash(string(node.Address))
	fingerentryminus1 := big.NewInt(int64(fingerentry) - 1)
	jump := new(big.Int).Exp(two, fingerentryminus1, nil)
	sum := new(big.Int).Add(n, jump)

	return new(big.Int).Mod(sum, hashMod)
}
