package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"strconv"
)

var listener *net.TCPListener

type Ring int

type RpcReply struct {
	Found bool
	Node  Node
}

func (node Node) findSuccessorRpc(id []byte) (bool, *Node) {
	fmt.Println("Trying to join: " + string(node.Address))
	client, err := rpc.Dial("tcp", string(node.Address))
	checkError(err)

	var reply RpcReply

	err = client.Call("Ring.FindSuccessor", &id, &reply)
	return reply.Found, &reply.Node
}

func (t *Ring) FindSuccessor(id []byte, reply *RpcReply) error {
	found, retNode := myNode.findSuccessor(id)
	*reply = RpcReply{Found: found, Node: retNode}
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
