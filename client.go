package main

import (
	"crypto/sha1"
	"encoding/hex"
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
	Id      string
	node    Node
}

// Main
func main() {
	// Setup arguments
	setupArguments()

	h := sha1.New()
	id := hex.EncodeToString(h.Sum([]byte(*ja + string(*jp))))

	// Error handling in arguments file, so we only need to check if ja is set
	addr := NodeAddress(*a + ":" + strconv.Itoa(*p))
	myNode = Node{address: addr, m: *r, id: id}
	if *ja == "" {
		myNode.create()
		myNode.print()
		startListen()

	} else {
		joinAddr := NodeAddress(*ja + ":" + strconv.Itoa(*jp))

		joinNode := Node{address: joinAddr, m: *r}
		client, err := rpc.Dial("tcp", string(joinNode.address))
		checkError(err)

		args := Args{string(addr), id, myNode}

		var reply int
		err = client.Call("Ring.Join", args, &reply)
		myNode.print()
		println(reply)
		checkError(err)
	}

}

func (r *Ring) Join(args Args, reply *int) error {
	log.Println("A new node has joined the network!")
	log.Println("Adress: " + args.Address)
	log.Println("ID: " + args.Id)

	args.node.print()

	myNode.join(args.node)
	myNode.print()

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

func PrintState(node Node) {
	fmt.Println("+-+-+-+-+-+ Node info +-+-+-+-+-+-\n")
	fmt.Println("ID: ", node.id, "\nIP addr:"+node.address)

	fmt.Println("\n+-+-+-+-+-+ Successors info +-+-+-+-+-+-\n")
	for i, suc := range node.successor {
		fmt.Println("\nSuccessor node ", i, "info\n")
		fmt.Println("ID: ", suc.id, "\nIP addr:"+suc.address)
	}

	fmt.Println("\n+-+-+-+-+-+ Fingertable info +-+-+-+-+-+-\n")
	for i, finger := range node.fingerTable {
		fmt.Println("\nFinger node", i, "info\n")
		fmt.Println("ID: ", finger.id, "\nIP addr:"+finger.address)
	}
}
