package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Set with information from arguments
var a, ja, i *string
var p, jp, ts, tff, tcp, r *int

var myNode Node

// Main
func main() {
	// Setup arguments
	setupArguments()
	myNode = newNode(*a, *p, *i, *r)

	if *ja == "" {
		myNode.create()
	} else {
		joinAddress := *ja + ":" + strconv.Itoa(*jp)
		joinNode := BasicNode{Address: joinAddress}
		myNode.join(joinNode)
	}

	// Init for listening
	initListen()

	initRoutines()

	go scan()

	//Main for loop
	for {
		listen() // Go routine for listening to traffic
	}
}

func scan() {
	// Init for reading stdin
	scanner := bufio.NewScanner((os.Stdin))
	for {
		scanner.Scan()
		txt := scanner.Text()
		sliceText := strings.Split(txt, " ")

		switch sliceText[0] {
		case "StoreFile":
			StoreFile(sliceText[1])
			break
		case "Lookup":
			Lookup(sliceText[1])
			break
		case "PrintState":
			PrintState()
			break
		}
	}
}

func StoreFile(filepath string) {
	fmt.Println("Filepath: " + filepath)
	slicedPath := strings.Split(filepath, "/")
	filename := slicedPath[len(slicedPath)-1]

	foundNode := Lookup(filename)
	foundNode.rpcStoreFile(filename)

	fmt.Println("File successfully uploaded to: ")
	fmt.Println("\tID: ", string(foundNode.Id), "\n\tIP/port: ", foundNode.Address)

}

func Lookup(filename string) BasicNode {
	fmt.Println("Filename: " + filename)
	hashed := hash(filename)
	myBasicNode := BasicNode{Address: myNode.Address, Id: myNode.Id}
	foundNode := find(hashed, myBasicNode)

	fmt.Println("\n+-+-+-+-+-+-+ Node with file info +-+-+-+-+-+--+")
	fmt.Println("\tID: ", string(foundNode.Id), "\n\tIP/port: ", foundNode.Address)
	fmt.Println("\t-------------------------------")

	return foundNode
}

func PrintState() {

	fmt.Println("+-+-+-+-+-+ Node info +-+-+-+-+-+-\n")
	fmt.Println("\tID: ", string(myNode.Id), "\n\tIP/port: ", myNode.Address)

	if len(myNode.Successor) > 0 {
		fmt.Println("\n+-+-+-+-+-+-+ Successors info +-+-+-+-+-+--+")
		for i, suc := range myNode.Successor {
			fmt.Println("\n\t-----Successor node", i, "info-----")
			fmt.Println("\tID: ", string(suc.Id), "\n\tIP/port: ", suc.Address)
			fmt.Println("\t-------------------------------")
		}
	} else {
		fmt.Println("\nNo Successors Found")
	}

	if len(myNode.FingerTable) > 0 {
		fmt.Println("\n+-+-+-+-+-+-+ Fingertable info +-+-+-+-+-+--+")
		for i, finger := range myNode.FingerTable {
			if finger != nil {
				//fmt.Println("\n\t-----Finger node", i, "info-----")
				fmt.Println("\tFinger node: ", i, "\tID: ", string(finger.Id), "\tIP/port: ", finger.Address)
				//fmt.Println("\t-------------------------------")
			}
		}
	} else {
		fmt.Println("\nFingertable Empty")
	}

	if myNode.Predecessor != nil {
		fmt.Println("\n+-+-+-+-+-+-+ Predecessor info +-+-+-+-+-+--+")
		fmt.Println("ID: ", string(myNode.Predecessor.Id), "\nIP/port: ", myNode.Predecessor.Address)
		fmt.Println("-------------------------------")
	} else {
		fmt.Println("\nNo Predecessor found")
	}
}
