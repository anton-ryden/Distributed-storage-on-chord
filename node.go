package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"math/big"
	"strconv"
)

const m = sha1.Size * 9

var two = big.NewInt(2)
var hashMod = new(big.Int).Exp(big.NewInt(2), big.NewInt(m), nil)

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

func newNode(ip string, port int, iArg string, r int) Node {
	// Error handling in arguments file, so we only need to check if ja is set
	addr := NodeAddress(ip + ":" + strconv.Itoa(port))

	// If i argument is used we set the id to that
	var id []byte
	if iArg == "" {
		iArg = ip + strconv.Itoa(port)
		id = hash(iArg)
	} else {
		id = []byte(iArg)
	}

	return Node{Address: addr, R: r, Id: id}
}

func hash(ipPort string) []byte {
	h := sha1.New()
	ha := new(big.Int).SetBytes(h.Sum([]byte(ipPort)))
	return []byte(ha.String())
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
	log.Println("Joining: " + joinNode.Address + "\tLooking for id: ")
	node.Predecessor = nil
	found, successor, maxSteps := false, &Node{}, 5

	for !found && maxSteps > 0 {
		found, successor = joinNode.findSuccessorRpc(node.Id)
		maxSteps--
	}

	node.Successor = append(node.Successor, successor)
}

func (node Node) find(id []byte, start Node) Node {
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
	if bytes.Equal(x.Id, node.Id) || bytes.Equal(node.Id, suc.Id) {
		node.Successor[0] = x
	}
	node.Successor[0].notifyRpc(node)
}

// n′ thinks it might be our predecessor.
func (node Node) notify(n Node) {
	if node.Predecessor == nil || (&n == node.Predecessor || n.Address == node.Address) {
		node.Predecessor = &n
	}
}

// called periodically. refreshes finger table entries.
// next stores the index of the next finger to fix.
func (node Node) fixFingers() {
	for i := 1; i < m; i++ {
		if i > m {
			i = 1
		}
		found, suc := node.findSuccessor(node.jump(i))
		if found {
			node.FingerTable[i] = &suc
		}
	}
}

// called periodically. checks whether predecessor has failed.
func (node Node) checkPredecessor() {
	if !node.Predecessor.checkAliveRpc() {
		node.Predecessor = nil
	}
}

// ask node n to find the successor of id
// or a better node to continue the search with
func (node Node) findSuccessor(id []byte) (bool, Node) {
	for _, suc := range node.Successor {
		if bytes.Equal(id, suc.Id) {
			return true, *suc
		}
	}

	return false, node.closestPrecedingNode(id)
}

// search the local table for the highest predecessor of id
func (node Node) closestPrecedingNode(id []byte) Node {
	for i := m; i > 1; i-- {
		if bytes.Equal(node.FingerTable[i].Id, id) {
			iFinger := node.FingerTable[i]
			return *iFinger
		}
	}
	return *node.Successor[0]
}

func (node Node) jump(fingerentry int) []byte {
	fingerentryminus1 := big.NewInt(int64(fingerentry) - 1)
	jump := new(big.Int).Exp(two, fingerentryminus1, nil)
	n := new(big.Int).SetBytes(node.Id)
	sum := new(big.Int).Add(n, jump)
	result := new(big.Int).Mod(sum, hashMod)
	return []byte(result.String())
}
