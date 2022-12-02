package main

import (
	"log"
	"net"
	"net/rpc"
)

var listener *net.TCPListener

func (r *Ring) Join(node Node, reply *int) error {
	node.print()
	//myNode.join(args.node)
	//myNode.print()
	return nil
}

func initListen() {
	ring := new(Ring)
	rpc.Register(ring)

	tcpAddr, err := net.ResolveTCPAddr("tcp", string(myNode.Address))
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
