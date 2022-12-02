package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/rpc"
	"os"
	"strconv"
)

// Set with information from arguments
var a, ja, i *string
var p, jp, ts, tff, tcp, r *int

type Ring int

var myNode Node

// Main
func main() {
	// Setup arguments
	setupArguments()

	// Error handling in arguments file, so we only need to check if ja is set
	addr := NodeAddress(*a + ":" + strconv.Itoa(*p))
	id := hash(*a + strconv.Itoa(*p))
	myNode = Node{Address: addr, M: *r, Id: id}

	if *ja == "" {
		myNode.create()
		myNode.print()
	} else {
		client, err := rpc.Dial("tcp", *ja+":"+strconv.Itoa(*jp))
		checkError(err)

		var reply int
		err = client.Call("Ring.Join", myNode, &reply)
		myNode.print()
		println(reply)
		checkError(err)
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

func hash(ipPort string) string {
	h := sha1.New()
	return hex.EncodeToString(h.Sum([]byte(ipPort)))
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
	fmt.Println("ID: ", myNode.Id, "\nIP addr:"+myNode.Address)

	fmt.Println("\n+-+-+-+-+-+ Successors info +-+-+-+-+-+-\n")
	for i, suc := range myNode.Successor {
		fmt.Println("\nSuccessor node ", i, "info\n")
		fmt.Println("ID: ", suc.Id, "\nIP addr:"+suc.Address)
	}

	fmt.Println("\n+-+-+-+-+-+ Fingertable info +-+-+-+-+-+-\n")
	for i, finger := range myNode.FingerTable {
		fmt.Println("\nFinger node", i, "info\n")
		fmt.Println("ID: ", finger.Id, "\nIP addr:"+finger.Address)
	}
}
