package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Set with information from arguments
var a, ja, i *string
var p, jp, ts, tff, tcp, r *int

// This clients node
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

	// Init go routines that run fixFingers, Stabilize and checkPredecessor
	initRoutines()

	// Go routine for scanning input from user
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
		default:
			fmt.Print("\nCommand not found")
		}
	}
}

// Stores file from filePath in the ring
func StoreFile(filePath string) {
	slicedPath := strings.Split(filePath, "/")
	filename := slicedPath[len(slicedPath)-1]

	if _, err := os.Stat(filePath); err == nil {
		foundNode, found, hashed := Lookup(filename)
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Println("os.open error:", err)
			return
		}
		myFile := BasicFile{Filename: filename, Key: hashed, FileContent: fileContent}

		if !found && foundNode.Id != nil {
			result := foundNode.rpcStoreFile(myFile)
			if !result {
				return
			}
			fmt.Println("\nFile successfully uploaded to: ")
			fmt.Println("\tID: ", string(foundNode.Id), "\n\tIP/port: ", foundNode.Address)
		} else {
			fmt.Println("\nFile already in the ring on the node above ^^")
		}
	} else {
		fmt.Println("\nError: file does not exist")
	}
}

// Lookups if fileName exists in the ring
func Lookup(fileName string) (BasicNode, bool, []byte) {
	fmt.Println("\nFilename: " + fileName)
	hashed := hash(fileName)
	myBasicNode := BasicNode{Address: myNode.Address, Id: myNode.Id}
	foundNode := find(hashed, myBasicNode)

	isFile := foundNode.rpcFileExist(hashed)

	if isFile == true {

		fmt.Println("\n+-+-+-+-+-+-+ Node with file info +-+-+-+-+-+--+")
		fmt.Println("\tID: ", string(foundNode.Id), "\n\tIP/port: ", foundNode.Address)
		fmt.Println("\t-------------------------------")

	} else {
		fmt.Println("\nFile does not exist in ring")
	}
	return foundNode, isFile, hashed
}

// Print information of this clients node
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
				fmt.Println("\tFinger node: ", i, "\tID: ", string(finger.Id), "\tIP/port: ", finger.Address)
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
