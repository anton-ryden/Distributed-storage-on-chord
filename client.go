package main

import (
	"bufio"
	"crypto/sha1"
	"fmt"
	"math/big"
	"os"
	"strconv"
)

// Set with information from arguments
var a, ja, i *string
var p, jp, ts, tff, tcp, r *int
var m int = 40

type Ring int

var myNode Node

// Main
func main() {
	// Setup arguments
	setupArguments()

	// Error handling in arguments file, so we only need to check if ja is set
	addr := NodeAddress(*a + ":" + strconv.Itoa(*p))
	if *i == "" {
		*i = *a + strconv.Itoa(*p)
	}

	id := hash(*i)
	myNode = Node{Address: addr, R: *r, Id: id}

	if *ja == "" {
		myNode.create()

	} else {
		node1 := Node{
			Address:     "node1",
			FingerTable: nil,
			Predecessor: nil,
			Successor:   nil,
			Next:        0,
			R:           0,
			Id:          big.NewInt(1),
			Bucket:      nil,
		}
		node2 := Node{
			Address:     "node2",
			FingerTable: nil,
			Predecessor: nil,
			Successor:   nil,
			Next:        0,
			R:           0,
			Id:          big.NewInt(2),
			Bucket:      nil,
		}
		node3 := Node{
			Address:     "node3",
			FingerTable: nil,
			Predecessor: nil,
			Successor:   nil,
			Next:        0,
			R:           0,
			Id:          big.NewInt(3),
			Bucket:      nil,
		}
		node4 := Node{
			Address:     "node4",
			FingerTable: nil,
			Predecessor: nil,
			Successor:   nil,
			Next:        0,
			R:           0,
			Id:          big.NewInt(4),
			Bucket:      nil,
		}

		node1.Successor = append(node1.Successor, &node2)
		node1.Successor = append(node1.Successor, &node3)
		node2.Successor = append(node2.Successor, &node3)
		node2.Successor = append(node2.Successor, &node4)
		node3.Successor = append(node3.Successor, &node4)
		node3.Successor = append(node3.Successor, &node1)
		node4.Successor = append(node4.Successor, &node1)
		node4.Successor = append(node4.Successor, &node2)

		nodeTemp := Node{
			Address:     "temp",
			FingerTable: nil,
			Predecessor: nil,
			Successor:   nil,
			Next:        0,
			R:           0,
			Id:          big.NewInt(666),
			Bucket:      nil,
		}
		nodeTemp.find(*big.NewInt(69), node1)
	}

	// Init for listening
	initListen()

	// Init for reading stdin
	scanner := bufio.NewScanner((os.Stdin))

	functions := map[string]interface{}{
		"StoreFile":  StoreFile,
		"Lookup":     Lookup,
		"PrintState": PrintState,
	}

	//Main for loop
	for {
		go listen() // Go routine for listening to traffic

		scanner.Scan()
		txt := scanner.Text()
		for key, element := range functions {
			if key == txt {
				element.(func())()
			}
		}
	}
}

func hash(ipPort string) *big.Int {
	h := sha1.New()
	return new(big.Int).SetBytes(h.Sum([]byte(ipPort))[:m])
}

func checkError(err error) {
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		os.Exit(1)
	}
}

func StoreFile() {

}

func Lookup() {

}

func PrintState() {
	fmt.Println("+-+-+-+-+-+ Node info +-+-+-+-+-+-\n")
	fmt.Println("   ID:", myNode.Id, "\n   IP/port: "+myNode.Address)

	if len(myNode.Successor) > 0 {
		fmt.Println("\n+-+-+-+-+-+ Successors info +-+-+-+-+-+-")
		for i, suc := range myNode.Successor {
			fmt.Println("\n   Successor node", i, "info -------------")
			fmt.Println("    ID:", suc.Id, "\n    IP/port: "+suc.Address)
			fmt.Println("   ------------------------------------")
		}
	} else {
		fmt.Println("\nNo Successors Found")
	}

	if len(myNode.FingerTable) > 0 {
		fmt.Println("\n+-+-+-+-+-+ Fingertable info +-+-+-+-+-+-")
		for i, finger := range myNode.FingerTable {
			fmt.Println("\n   Finger node", i, "info -------------")
			fmt.Println("    ID:", finger.Id, "\n    IP/port: "+finger.Address)
			fmt.Println("   ------------------------------------")
		}
	} else {
		fmt.Println("\nFingertable Empty")
	}
}
