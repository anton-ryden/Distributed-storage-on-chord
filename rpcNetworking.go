package main

import (
	"log"
	"net"
	"net/rpc"
	"strconv"
	"time"
)

var listener *net.TCPListener

type Ring int

const timeoutMs = 50000

type RpcReply struct {
	Found bool
	Node  BasicNode
}

func call(address string, method string, request interface{}, response interface{}){
	conn, err := net.DialTimeout("tcp", address, time.Millisecond*timeoutMs)
	if err != nil {
		log.Fatalln(err)
	}
	client := rpc.NewClient(conn)
	defer client.Close()

	//get call
	err = client.Call(method, request, response)
	if err != nil {
		log.Fatalln(err)
	}
}

func (node *Node) rpcCopySuccessorOf(getSucOf *BasicNode) {
	var response *Node
	call(getSucOf.Address, "Ring.CopySuccessorOf", false, &response)

	node.Successor = append([]*BasicNode{getSucOf}, response.Successor...)
	sucLen := len(node.Successor)
	if sucLen > *r {
		node.Successor = node.Successor[:sucLen-1]
	}

}

func (ring *Ring) CopySuccessorOf(inBool bool, reply *Node) error {
	*reply = Node{Successor: myNode.Successor}
	return nil
}

func (node *Node) rpcGetPredecessorOf(getMyPredecessor *BasicNode) BasicNode{
	var response *BasicNode
	call(getMyPredecessor.Address, "Ring.GetPredecessor", false, &response)
	return *response
}

func (ring *Ring) GetPredecessor(inBool bool, reply *BasicNode) error {
	if myNode.Predecessor == nil{
		*reply = BasicNode{}
	}else{
		*reply = *myNode.Predecessor
	}
	return nil
}

func (node *Node) rpcUpdateSuccessorOf(updateMySuccessor *BasicNode) {
	var response bool
	send := BasicNode{Address: node.Address, Id: node.Id}
	call(updateMySuccessor.Address, "Ring.UpdateSuccessorOf", send, &response)
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
	return nil
}

func (node *BasicNode) rpcNotifyOf(notifyOfMe *Node) {
	var response bool
	sendNode := BasicNode{Address: notifyOfMe.Address, Id: notifyOfMe.Id}
	call(node.Address, "Ring.NotifyOf", sendNode, &response)
}

func (ring *Ring) NotifyOf(notifyOf BasicNode, reply *bool) error {
	myNode.notify(notifyOf)
	return nil
}

func (node *BasicNode) rpcIsAlive() bool {
	var response bool
	call(node.Address, "Ring.CheckAlive", false, &response)
	return response
}

func (ring *Ring) CheckAlive(inBool bool, reply *bool) error {
	*reply = true
	return nil
}

func (node *BasicNode) rpcFindSuccessor(id []byte) (bool, BasicNode) {
	var response RpcReply
	call(node.Address, "Ring.FindSuccessor", &id, &response)
	return response.Found, response.Node
}

func (ring *Ring) FindSuccessor(id []byte, reply *RpcReply) error {
	found, retNode := myNode.findSuccessor(id)
	reply.Found = found
	reply.Node = retNode
	return nil
}

func initListen() {
	ring := new(Ring)
	rpc.Register(ring)
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(*p))
	if err != nil{
		log.Fatalln(err)
	}

	listener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil{
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
