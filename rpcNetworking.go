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

func (r *Ring) find(idToFind string, reply *string) {
	//return myNode.find(idToFind)
}

func find(address string, idToFind string) {
	client, err := rpc.Dial("tcp", address)
	checkError(err)

	var reply string
	err = client.Call("Ring.Find", idToFind, &reply)
	myNode.print()
	println(reply)
	checkError(err)
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
