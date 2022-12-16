package main

import (
	"bytes"
	"crypto/sha1"
	"log"
	"math/big"
	"strconv"
	"time"
)

// m bits
const m = sha1.Size * 24

// Max amount of request to find successor
var maxSteps = 32

// Struct for this clients node
type Node struct {
	Address      string
	FingerTable  []*BasicNode
	Predecessor  *BasicNode
	Successor    []*BasicNode
	Id           []byte
	Bucket       map[string]string
	BackupBucket map[string]string
}

// BasicNode Struct: For nodes inside Node struct. Require less information and no recursion
type BasicNode struct {
	Address string
	Id      []byte
}

// Struct for sending and saving files
type BasicFile struct {
	Filename    string
	Key         []byte
	FileContent []byte
}

// Creates a new client node
func newNode(ip string, port int, iArg string, r int) Node {
	// Setups the certificate to communicate with other nodes with tls
	setConfig()
	// Error handling in arguments file, so we only need to check if ja is set
	addr := ip + ":" + strconv.Itoa(port)

	// If i argument is used we set the id to that
	var id []byte
	if iArg == "" {
		iArg = ip + strconv.Itoa(port)
		id = hash(iArg)
	} else {
		id = []byte(iArg)
	}

	myBucket := make(map[string]string)
	myBackupBucket := make(map[string]string)
	return Node{Address: addr, Id: id, Bucket: myBucket, BackupBucket: myBackupBucket}
}

// Called periodically. verifies node's immediate
// successor, and tells the successor about n
func (node *Node) stabilize() {
	// If it fails for some reason
	node.rpcCopySuccessor()
	x := node.rpcGetPredecessorOf(node.Successor[0])

	// Update successor if x is closer than successor[0]
	if x.Id != nil && between(x.Id, node.Id, node.Successor[0].Id) {
		node.Successor[0] = &x
	}

	// Notify immediate successor of node
	node.Successor[0].rpcNotifyOf(node)
	//node.Successor[0].rpcUpdateBackupBucketOf(node)
}

// n′ thinks it might be our predecessor.
func (node *Node) notify(n BasicNode) {
	if node.Predecessor == nil || between(n.Id, node.Predecessor.Id, node.Id) {
		node.Predecessor = &n
		log.Println("New Predecessor:\nNode:\n\tAddress:", node.Predecessor.Address, "\n\tId:\t", string(node.Predecessor.Id))
	}
}

// Create a new Chord ring.
func (node *Node) create() {
	// Set predecessor to itself to ensure predecessor never is nil. Not strictly necessary could be nil
	node.Predecessor = &BasicNode{Address: node.Address, Id: node.Id}
	// Add itself ass immediate successor
	node.Successor = append(node.Successor, &BasicNode{Address: node.Address, Id: node.Id})
}

// Join a Chord ring containing joiNode
func (node *Node) join(joinNode BasicNode) {
	log.Println("Sending request to join to: " + joinNode.Address + "\t")
	successor := find(node.Id, joinNode)
	// If node did not exist we add joiNode as successor
	node.Successor = append(node.Successor, &successor)
	predOfSuc := node.rpcGetPredecessorOf(node.Successor[0])
	// Tell our predecessor that we are the new node
	node.rpcUpdateSuccessorOf(&predOfSuc)
	// Some information prints
	log.Println("Immediate successor is:\n",
		"Node:\n",
		"\tAddress:", node.Successor[0].Address, "\n",
		"\tId:\t", string(node.Successor[0].Id))
}

// Find immediate successor to join
func find(id []byte, start BasicNode) BasicNode {
	found, nextNode := false, start
	i := 0
	for found == false && i < maxSteps {
		found, nextNode = nextNode.rpcFindSuccessor(id)
		i++
	}

	if found == true {
		return nextNode
	} else {
		return BasicNode{}
	}
}

// Called periodically. Refreshes finger table entries
func (node *Node) fixFingers() {
	node.FingerTable = []*BasicNode{}
	for i := 0; i < m; i++ {
		if i > m {
			i = 0
		}

		jump := node.jump(i)
		_, suc := node.findSuccessor(jump)
		node.FingerTable = append(node.FingerTable, &BasicNode{Address: suc.Address, Id: suc.Id})
	}
}

// Called periodically. Checks whether predecessor has failed
func (node *Node) checkPredecessor() {
	if node.Predecessor == nil {
		return
	}

	if bytes.Equal(node.Predecessor.Id, node.Id) {
		return
	}

	if !node.Predecessor.rpcIsAlive() {
		node.Predecessor = nil
	}
}

// Ask node to find the successor of id
// or a better node to continue the search with
func (node *Node) findSuccessor(id []byte) (bool, BasicNode) {
	prev := node.Id
	for _, suc := range node.Successor {
		if between(id, prev, suc.Id) {
			return true, *suc
		}
		prev = suc.Id
	}
	return false, node.closestPrecedingNode(id)
}

// Check if elt id is between start and end
func between(elt, start, end []byte) bool {
	switch bytes.Compare(start, end) {
	case 1:
		return bytes.Compare(start, elt) == -1 || bytes.Compare(end, elt) >= 0
	case -1:
		return bytes.Compare(start, elt) == -1 && bytes.Compare(end, elt) >= 0
	case 0:
		return bytes.Compare(start, elt) != 0
	}
	return false

}

// Search the local table for the highest predecessor of id
func (node *Node) closestPrecedingNode(id []byte) BasicNode {
	for i := m - 1; i > 1; i-- {
		if len(node.FingerTable) <= i {
			continue
		} else if node.FingerTable[i] == nil {
			continue
		}
		if bytes.Equal(node.FingerTable[i].Id, id) {
			iFinger := node.FingerTable[i]
			return *iFinger
		}
	}
	return *node.Successor[0]
}

// Initialize the go routines that run fixFingers, Stabilize and checkPredecessor
func initRoutines() {
	// Periodically fixFingers the node.
	go func() {
		ticker := time.NewTicker(time.Millisecond * time.Duration(*tff))
		done := make(chan bool)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				myNode.fixFingers()
			}
		}
	}()

	// Periodically stabilize the node.
	go func() {
		ticker := time.NewTicker(time.Millisecond * time.Duration(*ts))
		done := make(chan bool)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				myNode.stabilize()
			}
		}
	}()
	// Periodically checkPredecessor the node.
	go func() {
		ticker := time.NewTicker(time.Millisecond * time.Duration(*tcp))
		done := make(chan bool)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				myNode.checkPredecessor()
			}
		}
	}()

}

// Hash string with sha1. Returns []byte
func hash(stringInput string) []byte {
	h := sha1.New()
	hashedVal := new(big.Int).SetBytes(h.Sum([]byte(stringInput)))
	hashMod := new(big.Int).Exp(big.NewInt(2), big.NewInt(m), nil)
	hashedVal.Mod(hashedVal, hashMod)
	return []byte(hashedVal.String())
}

// Arithmetic used in fingerTable
func (node *Node) jump(fingerentry int) []byte {
	two := big.NewInt(2)
	hashMod := new(big.Int).Exp(big.NewInt(2), big.NewInt(m), nil)
	jump := new(big.Int).Exp(two, big.NewInt(int64(fingerentry)), nil)
	n := new(big.Int).SetBytes(node.Id)
	sum := new(big.Int).Add(n, jump)
	result := new(big.Int).Mod(sum, hashMod)
	res := result.Bytes()
	return res
}
