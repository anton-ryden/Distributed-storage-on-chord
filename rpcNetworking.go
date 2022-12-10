package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strconv"
	"time"
)

var listener *net.TCPListener

type Ring int

type RpcReply struct {
	Found bool
	Node  BasicNode
}

func (node *Node) updateRpc(suc *BasicNode) {
	client, err := rpc.Dial("tcp", string(suc.Address))
	defer client.Close()
	checkError(err)

	var reply *Node
	err = client.Call("Ring.Update", false, &reply)
	if err != nil {
		log.Println("Ring.Update", err)
	}
	replySuccessor := reply.Successor
	if !bytes.Equal(replySuccessor[0].Id, suc.Id) {
		node.Successor = append([]*BasicNode{suc}, replySuccessor...)
		sucLen := len(node.Successor)
		if sucLen > *r {
			node.Successor = node.Successor[:sucLen-1]
		}
	}

}

func (node *Node) getPredecessorRPC (predof *BasicNode) BasicNode{
	client, err := rpc.Dial("tcp", string(predof.Address))
	defer client.Close()
	checkError(err)

	var reply *BasicNode
	err = client.Call("Ring.GetPredecessor", false, &reply)
	if err != nil {
		log.Println("Ring.GetPredecessor", err)
	}
	return *reply
}

func (node *BasicNode) updateImmSuccessorRpc(suc *Node) {
	client, err := rpc.Dial("tcp", string(node.Address))
	defer client.Close()
	checkError(err)

	var reply bool
	send := BasicNode{Address: suc.Address, Id: suc.Id}

	err = client.Call("Ring.UpdateImmSuccessor", &send, &reply)
	if err != nil {
		log.Println("Ring.GetNode", err)
	}
}

func (ring *Ring) UpdateImmSuccessor(newSuccessor *BasicNode, reply *bool) error {
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

func (node *BasicNode) notifyRpc(notifyOfMe *Node) {
	client, err := rpc.Dial("tcp", string(node.Address))
	checkError(err)

	var reply bool
	defer client.Close()
	send := BasicNode{Address: notifyOfMe.Address, Id: notifyOfMe.Id}
	err = client.Call("Ring.Notify", &send, &reply)
	if err != nil {
		log.Println("Ring.Notify ", err)
	}
}

func (node *BasicNode) checkAliveRpc() bool {
	// Timeout connection if exceed 400ms. If timout occur we consider node dead
	conn, err := net.DialTimeout("tcp", string(node.Address), time.Millisecond*400)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	client := rpc.NewClient(conn)
	defer client.Close()
	var reply bool
	err = client.Call("Ring.CheckAlive", false, &reply)
	if err != nil {
		fmt.Println("ID: ", string(node.Id), " is dead")
		fmt.Println(err)
		reply = false
	}

	return reply
}

func (node *BasicNode) findSuccessorRpc(id []byte) (bool, BasicNode) {
	client, err := rpc.Dial("tcp", string(node.Address))
	defer client.Close()
	checkError(err)

	var reply RpcReply

	err = client.Call("Ring.FindSuccessor", &id, &reply)

	if err != nil {
		fmt.Println("Ring.FindSuccessor", err)
		return false, BasicNode{}
	}
	return reply.Found, reply.Node
}

func (ring *Ring) Update(inBool bool, reply *Node) error {
	*reply = myNode
	return nil
}

func (ring *Ring) Notify(notifyOf BasicNode, reply *bool) error {
	myNode.notify(notifyOf)
	return nil
}

func (ring *Ring) GetPredecessor(inBool bool, reply *BasicNode) error {
	*reply = *myNode.Predecessor
	return nil
}

func (ring *Ring) CheckAlive(inBool bool, reply *bool) error {
	*reply = true
	return nil
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
	checkError(err)

	listener, err = net.ListenTCP("tcp", tcpAddr)
	checkError(err)
}

func listen() {
	conn, err := listener.Accept()
	if err != nil {
		log.Println("Listen accept error: " + err.Error())
	}
	rpc.ServeConn(conn)
}
