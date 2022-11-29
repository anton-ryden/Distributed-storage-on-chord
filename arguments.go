package main

import (
	"github.com/akamensky/argparse"
	"log"
	"os"
	"strconv"
)

// Setups all the arguments
func setupArguments() {
	// Create new parser object
	parser := argparse.NewParser("Chord client", "Chord client will be a command-line utility")

	// Create map that contains upperbound for argument
	argWithRanges := make(map[string]int)

	// All arguments with option (ex. if they are required and help argument)
	a = parser.String("a", "a", &argparse.Options{Required: true, Help: "The IP address that the Chord client will bind to, as well as advertise to other nodes. Represented as an ASCII string (e.g., 128.8.126.63)"})
	p = parser.Int("p", "p", &argparse.Options{Required: true, Help: "The identifier (ID) assigned to the Chord client which will override the ID computed by the SHA1 sum of the client’s IP address and port number. Represented as a string of 40 characters matching [0-9a-fA-F]. Optional parameter."})
	jp = parser.Int("", "jp", &argparse.Options{Help: "The port that an existing Chord node is bound to and listening on. The Chord client will join this node’s ring. Represented as a base-10 integer. Must be specified if --ja is specified."})
	ja = parser.String("", "ja", &argparse.Options{Help: "The IP address of the machine running a Chord node. The Chord client will join this node’s ring. Represented as an ASCII string (e.g., 128.8.126.63). Must be specified if --jp is specified."})
	ts = parser.Int("", "ts", &argparse.Options{Required: true, Help: "The time in milliseconds between invocations of ‘stabilize’. Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000]."})
	argWithRanges["ts"] = 60000
	tff = parser.Int("", "tff", &argparse.Options{Required: true, Help: "The time in milliseconds between invocations of ‘fix fingers’. Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000]."})
	argWithRanges["tff"] = 60000
	tcp = parser.Int("", "tcp", &argparse.Options{Required: true, Help: "The time in milliseconds between invocations of ‘check predecessor’. Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000]."})
	argWithRanges["tcp"] = 60000
	r = parser.Int("r", "r", &argparse.Options{Required: true, Help: "The number of successors maintained by the Chord client. Represented as a base-10 integer. Must be specified, with a value in the range of [1,32]."})
	argWithRanges["r"] = 32
	i = parser.String("i", "i", &argparse.Options{Help: "The IP address of the machine running a Chord node. The Chord client will join this node’s ring. Represented as an ASCII string (e.g., 128.8.126.63). Must be specified if --jp is specified."})

	// Parse arguments
	err := parser.Parse(os.Args)
	if err != nil {
		log.Println(err)
	}

	// Check if range of arguments is correct
	args := parser.GetArgs()
	checkRange(args, argWithRanges)

	// Check so if ja is set jp is set and vice versa
	if *ja != "" && *jp == 0 || *ja != "" && *jp == 0 {
		log.Println("If ja is set jp need to be set and vice versa")
		os.Exit(1)
	}
}

// Check if the upper and lower bound of argument is within range
func checkRange(args []argparse.Arg, upperBound map[string]int) {
	for _, arg := range args { // For all argument
		if upper, ok := upperBound[arg.GetLname()]; ok { // If the argument has a specified upper bound
			resultInt := *(arg.GetResult().(*int))  // From argument result (interface) to int
			if resultInt < 1 || resultInt > upper { // Check if inside range
				log.Println(arg.GetLname() + " has value " + strconv.Itoa(resultInt) + " which is outside of range: [1, " + strconv.Itoa(upper) + "]")
			}
		}
	}

}
