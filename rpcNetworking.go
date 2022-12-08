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
	Node  Node
}

func (node *Node) updateRpc(suc *Node) {
	client, err := rpc.Dial("tcp", string(suc.Address))
	defer client.Close()
	checkError(err)

	var reply *Node
	err = client.Call("Ring.Update", false, &reply)
	if err != nil {
		log.Println("Ring.Update", err)
	}
	sucSuc := reply.Successor
	if !bytes.Equal(sucSuc[0].Id, suc.Id) {
		node.Successor = append([]*Node{suc}, sucSuc...)
		sucLen := len(node.Successor)
		if sucLen > *r {
			node.Successor = node.Successor[:sucLen-1]
		}
	}

}

func (node *Node) getPredecessorRPC (predof *Node) Node{
	client, err := rpc.Dial("tcp", string(predof.Address))
	defer client.Close()
	checkError(err)

	var reply *Node
	err = client.Call("Ring.Update", false, &reply)
	if err != nil {
		log.Println("Ring.Update", err)
	}
	return *reply
}

func (node *Node) updateImmSuccessorRpc(suc *Node) {
	client, err := rpc.Dial("tcp", string(node.Address))
	defer client.Close()
	checkError(err)

	var reply bool

	err = client.Call("Ring.UpdateImmSuccessor", &suc, &reply)
	if err != nil {
		log.Println("Ring.GetNode", err)
	}
}

func (r *Ring) UpdateImmSuccessor(immSuc *Node, reply *bool) error {
	myNode.Successor[0] = immSuc
	return nil
}

func (node *Node) notifyRpc(notifyOfMe *Node) {
	client, err := rpc.Dial("tcp", string(node.Address))
	checkError(err)

	var reply bool
	defer client.Close()
	err = client.Call("Ring.Notify", &notifyOfMe, &reply)
	if err != nil {
		log.Println("Ring.Notify ", err)
	}
}

func (node *Node) checkAliveRpc() bool {
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

func (node *Node) findSuccessorRpc(id []byte) (bool, Node) {
	client, err := rpc.Dial("tcp", string(node.Address))
	defer client.Close()
	checkError(err)

	var reply RpcReply

	err = client.Call("Ring.FindSuccessor", &id, &reply)

	if err != nil {
		fmt.Println("Ring.FindSuccessor", err)
		return false, Node{}
	}
	return reply.Found, reply.Node
}

func (r *Ring) Update(inBool bool, reply *Node) error {
	*reply = myNode
	return nil
}

func (r *Ring) Notify(notifyOf Node, reply *bool) error {
	myNode.notify(notifyOf)
	return nil
}

func (r *Ring) GetPredecessor(inBool bool, reply *Node) error {
	reply = myNode.Predecessor
	return nil
}

func (r *Ring) CheckAlive(inBool bool, reply *bool) error {
	*reply = true
	return nil
}

func (r *Ring) FindSuccessor(id []byte, reply *RpcReply) error {
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
