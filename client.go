package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

// Set with information from arguments
var a, ja, i *string
var p, jp, ts, tff, tcp, r *int

type Ring int

var myNode Node

type Args struct {
	Address string
	Id      int
}

// Main
func main() {
	// Setup arguments
	setupArguments()

	// Error handling in arguments file, so we only need to check if ja is set
	addr := NodeAddress(*a + ":" + strconv.Itoa(*p))
	myNode = Node{address: addr, m: *r}
	if *ja == "" {
		//myNode.create()
		startListen()
	} else {
		joinAddr := NodeAddress(*ja + ":" + strconv.Itoa(*jp))
		joinNode := Node{address: joinAddr, m: *r, id: 5}
		//myNode.join(joinNode)
		client, err := rpc.Dial("tcp", string(joinNode.address))
		checkError(err)

		args := Args{string(joinAddr), 5}

		var reply int
		err = client.Call("Ring.Join", args, &reply)
		println(reply)
		checkError(err)
	}

}

func (r *Ring) Join(args *Args, reply *int) error {
	log.Println("Adress: " + args.Address)
	log.Println("Id: ", args.Id)
	return nil
}

func startListen() {
	ring := new(Ring)
	rpc.Register(ring)

	tcpAddr, err := net.ResolveTCPAddr("tcp", string(myNode.address))
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		rpc.ServeConn(conn)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}
