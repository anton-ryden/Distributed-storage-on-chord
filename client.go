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
	myNode = Node{Address: addr, M: *r, Id: id}
	if *ja == "" {
		myNode.create()
		myNode.print()
		startListen()

	} else {
		joinAddr := NodeAddress(*ja + ":" + strconv.Itoa(*jp))

		joinNode := Node{Address: joinAddr, M: *r}
		joinJoinNode := Node{
			Address:     "recursive",
			FingerTable: nil,
			Predecessor: nil,
			Successor:   nil,
			Next:        0,
			M:           0,
			Id:          "Aids",
			Bucket:      nil,
		}
		joinNode.Successor = append(joinNode.Successor, &joinJoinNode)

		client, err := rpc.Dial("tcp", string(joinNode.Address))
		checkError(err)

		var reply int
		err = client.Call("Ring.Join", joinNode, &reply)
		myNode.print()
		println(reply)
		checkError(err)
	}

}

func (r *Ring) Join(node Node, reply *int) error {
	log.Println("A new node has joined the network!")
	log.Println("Adress: " + node.Address)
	log.Println("ID: " + node.Id)

	//node.node.print()

	//myNode.join(args.node)
	myNode.print()

	return nil
}

func startListen() {
	ring := new(Ring)
	rpc.Register(ring)

	tcpAddr, err := net.ResolveTCPAddr("tcp", string(myNode.Address))
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
	fmt.Println("ID: ", node.Id, "\nIP addr:"+node.Address)

	fmt.Println("\n+-+-+-+-+-+ Successors info +-+-+-+-+-+-\n")
	for i, suc := range node.Successor {
		fmt.Println("\nSuccessor node ", i, "info\n")
		fmt.Println("ID: ", suc.Id, "\nIP addr:"+suc.Address)
	}

	fmt.Println("\n+-+-+-+-+-+ Fingertable info +-+-+-+-+-+-\n")
	for i, finger := range node.FingerTable {
		fmt.Println("\nFinger node", i, "info\n")
		fmt.Println("ID: ", finger.Id, "\nIP addr:"+finger.Address)
	}
}
