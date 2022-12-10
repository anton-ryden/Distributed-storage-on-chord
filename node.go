package main

import (
	"bytes"
	"crypto/sha1"
	"log"
	"math/big"
	"strconv"
	"time"
)

const m = sha1.Size * 9

type Key string
type NodeAddress string

var maxSteps = 32

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

// called periodically. verifies n’s immediate
// successor, and tells the successor about n.
func (node *Node) stabilize() {
	node.updateRpc(node.Successor[0])
	x := node.getPredecessorRPC(node.Successor[0])

	if bytes.Equal(x.Id, node.Id) || bytes.Equal(x.Id, node.Successor[0].Id) {
		node.Successor[0] =  &x
	}

	if !bytes.Equal(node.Successor[0].Id, node.Id) {
		node.Successor[0].notifyRpc(node)
	}
}

// create a new Chord ring.
func (node *Node) create() {
	node.Predecessor = nil
	node.Successor = append(node.Successor, &Node{Address: node.Address, R: node.R, Id: node.Id})
}

// join a Chord ring containing node n′.
func (node *Node) join(joinNode Node) {
	log.Println("Joining: " + joinNode.Address + "\t")
	node.Predecessor = nil
	successor := find(node.Id, joinNode)
	// if node did not exist we add joiNode as successor
	node.Successor = append(node.Successor, &successor)
	node.Successor[0].updateImmSuccessorRpc(node)
}

func find(id []byte, start Node) Node {
	found, nextNode := false, start
	i := 0
	for found == false && i < maxSteps {
		found, nextNode = nextNode.findSuccessorRpc(id)
		i++
	}

	if found == true {
		return nextNode
	} else {
		//return find(node.Successor[i])
		return Node{}
	}
}

// n′ thinks it might be our predecessor.
func (node *Node) notify(n Node) {
	if node.Predecessor == nil || between(n.Id, node.Predecessor.Id, node.Id) {
		node.Predecessor = &n
		//n.updateImmSuccessorRpc(node)
	}
}

// called periodically. refreshes finger table entries.
// next stores the index of the next finger to fix.
func (node *Node) fixFingers() {
	node.FingerTable = []*Node{}
	for i := 0; i < m; i++ {
		if i > m {
			i = 0
		}

		jump := node.jump(i)
		_, suc := node.findSuccessor(jump)
		node.FingerTable = append(node.FingerTable, fingerEntry(&suc))
	}
	print("")
}

func fingerEntry(node *Node) *Node {
	retNode := Node{
		Address: node.Address,
		Id:      node.Id,
	}
	return &retNode
}

// called periodically. checks whether predecessor has failed.
func (node *Node) checkPredecessor() {
	if node.Predecessor == nil {
		return
	}
	if bytes.Equal(node.Predecessor.Id, node.Id) {
		return
	}
	if !node.Predecessor.checkAliveRpc() {
		node.Predecessor = nil
	}
}

// ask node n to find the successor of id
// or a better node to continue the search with
func (node *Node) findSuccessor(id []byte) (bool, Node) {
	prev := node.Id
	for _, suc := range node.Successor {
		if between(id, prev, suc.Id) {
			return true, *suc
		} else if between(id, suc.Id, prev) {
			return true, *node
		}
		prev = suc.Id
	}
	return false, node.closestPrecedingNode(id)
}

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

// search the local table for the highest predecessor of id
func (node *Node) closestPrecedingNode(id []byte) Node {
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

func hash(ipPort string) []byte {
	h := sha1.New()
	ha := new(big.Int).SetBytes(h.Sum([]byte(ipPort)))
	hashMod := new(big.Int).Exp(big.NewInt(2), big.NewInt(m), nil)
	ha.Mod(ha, hashMod)
	return []byte(ha.String())
}

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
