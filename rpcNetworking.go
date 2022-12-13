package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"time"
)

var listener *net.TCPListener

type Ring int

const timeoutMs = 1000

type RpcReply struct {
	Found bool
	Node  BasicNode
}

func call(address string, method string, request interface{}, response interface{}) error {
	conn, err := net.DialTimeout("tcp", address, time.Millisecond*timeoutMs)
	if err != nil {
		return err
	}

	client := rpc.NewClient(conn)
	defer client.Close()
	//get call
	err = client.Call(method, request, response)

	return err
}

func (node *Node) rpcCopySuccessor() {
	var response *Node
	var err error
	var suc *BasicNode
	var i int
	// if call fails try next successor in list until list is empty.
	// When list becomes empty set successor to nil
	for i, suc = range node.Successor {
		if i != 0 {
			log.Println("Successor is dead. Trying next successor in list:\nNode:\n\tAddress:", node.Successor[0].Address, "\n\tId:\t", string(node.Successor[0].Id))
		}
		err := call(suc.Address, "Ring.CopySuccessor", false, &response)
		if err == nil { // if call did not generate error
			break
		}
	}

	if err != nil {
		myself := BasicNode{Address: node.Address, Id: node.Id}
		node.Successor[0] = &myself
		log.Println("All successors in list is dead new succesor is:\nNode:\n\tAddress:", node.Successor[0].Address, "\n\tId:\t", string(node.Successor[0].Id))
		//log.Fatalln("Method: Ring.rpcCopySuccessor Error: ", err)
		return
	}

	node.Successor = append([]*BasicNode{suc}, response.Successor...)
	sucLen := len(node.Successor)
	if sucLen > *r {
		node.Successor = node.Successor[:sucLen-1]
	}

}

func (ring *Ring) CopySuccessor(inBool bool, reply *Node) error {
	*reply = Node{Successor: myNode.Successor}
	return nil
}

func (node *Node) rpcGetPredecessorOf(getMyPredecessor *BasicNode) BasicNode {
	var response *BasicNode
	err := call(getMyPredecessor.Address, "Ring.GetPredecessor", false, &response)
	if err != nil {
		log.Println("Method: Ring.GetPredecessor Error: ", err)
	}
	return *response
}

func (ring *Ring) GetPredecessor(inBool bool, reply *BasicNode) error {
	if myNode.Predecessor == nil {
		*reply = BasicNode{}
	} else {
		*reply = *myNode.Predecessor
	}
	return nil
}

func (node *Node) rpcUpdateSuccessorOf(updateMySuccessor *BasicNode) {
	var response bool
	send := BasicNode{Address: node.Address, Id: node.Id}
	err := call(updateMySuccessor.Address, "Ring.UpdateSuccessorOf", send, &response)
	if err != nil {
		log.Fatalln("Method: Ring.UpdateSuccessorOf Error: ", err)
	}
}

func (ring *Ring) UpdateSuccessorOf(newSuccessor *BasicNode, reply *bool) error {
	// Append new successor in the begging of successor array
	oldSuccessors := myNode.Successor
	myNode.Successor = append([]*BasicNode{newSuccessor}, oldSuccessors...)

	// Check if array length need to be changed
	sucLen := len(myNode.Successor)
	if sucLen > *r {
		myNode.Successor = myNode.Successor[:sucLen-1]
	}

	log.Println("New immediate successor:\n",
		"Node:\n",
		"\tAddress:", myNode.Successor[0].Address, "\n",
		"\tId:\t", string(myNode.Successor[0].Id))

	return nil
}

func (node *BasicNode) rpcNotifyOf(notifyOfMe *Node) {
	var response bool
	sendNode := BasicNode{Address: notifyOfMe.Address, Id: notifyOfMe.Id}
	err := call(node.Address, "Ring.NotifyOf", sendNode, &response)
	if err != nil {
		log.Println("Method: Ring.NotifyOf Error: ", err)
	}
}

func (ring *Ring) NotifyOf(notifyOf BasicNode, reply *bool) error {
	myNode.notify(notifyOf)
	return nil
}

func (node *BasicNode) rpcIsAlive() bool {
	var response bool
	err := call(node.Address, "Ring.CheckAlive", false, &response)
	if err != nil {
		log.Println("Address: ", node.Address, " Id: ", string(node.Id), " is no longer alive"+
			", Predecessor is now nil")
		return false
	}
	return response
}

func (ring *Ring) CheckAlive(inBool bool, reply *bool) error {
	*reply = true
	return nil
}

func (node *BasicNode) rpcFindSuccessor(id []byte) (bool, BasicNode) {
	var response RpcReply
	err := call(node.Address, "Ring.FindSuccessor", &id, &response)
	if err != nil {
		log.Println("Method: Ring.FindSuccessor Error: ", err)
	}
	return response.Found, response.Node
}

func (ring *Ring) FindSuccessor(id []byte, reply *RpcReply) error {
	found, retNode := myNode.findSuccessor(id)
	reply.Found = found
	reply.Node = retNode
	return nil
}

func (node *BasicNode) rpcFileExist(key []byte) bool {
	var response *bool
	err := call(node.Address, "Ring.FileExist", key, &response)
	if err != nil {
		log.Println("Method: Ring.FileExist Error: ", err)
	}
	return *response
}

func (ring *Ring) FileExist(key []byte, reply *bool) error {
	myString := string(key[:])
	_, found := myNode.Bucket[myString]
	if found {
		*reply = true
	} else {
		*reply = false
	}
	return nil
}

func (node *BasicNode) rpcStoreFile(filename BasicFile) {
	var response *bool
	err := call(node.Address, "Ring.StoreFile", filename, &response)
	if err != nil {
		log.Println("Method: Ring.StoreFile Error: ", err)
	}
}

func (ring *Ring) StoreFile(file BasicFile, reply *bool) error {
	if _, err := os.Stat(file.Filename); err != nil {
		myString := string(file.Key[:])
		myNode.Bucket[myString] = file.Filename
	} else {
		fmt.Println("file already on system")
	}
	return nil
}

func initListen() {
	ring := new(Ring)
	rpc.Register(ring)
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(*p))
	if err != nil {
		log.Fatalln(err)
	}

	listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Fatalln(err)
	}
}

func listen() {
	conn, err := listener.Accept()
	if err != nil {
		log.Println("Listen accept error: " + err.Error())
	}
	rpc.ServeConn(conn)
}
