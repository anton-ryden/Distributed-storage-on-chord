package main

import (
	"log"
	"math/big"
	"net"
	"net/rpc"
	"strconv"
)

var listener *net.TCPListener

type Ring int

func (t *Ring) FindSuccessor(id *big.Int, reply *Node) error {
	//_, retNode := myNode.findSuccessor(*id)
	*reply = Node{Address: "Testing"}
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
