package main

import (
	"bufio"
	"fmt"
	"os"
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
		joinNode := newNode(*ja, *jp, "", *r)
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

	functions := map[string]interface{}{
		"StoreFile":  StoreFile,
		"Lookup":     Lookup,
		"PrintState": PrintState,
	}

	for {
		scanner.Scan()
		txt := scanner.Text()
		for key, element := range functions {
			if key == txt {
				element.(func())()
			}
		}
	}
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
}
