package main

import (
	"flag"
	"os"
)

var a, ja, i string
var p, jp, ts, tff, tcp, r int

func main() {
	a = *flag.String("a", "", "The IP address that the Chord client will bind to, as well as advertise to other nodes. Represented as an ASCII string (e.g., 128.8.126.63)")
	ja = *flag.String("ja", "", "The IP address of the machine running a Chord node. The Chord client will join this node’s ring. Represented as an ASCII string (e.g., 128.8.126.63). Must be specified if --jp is specified.")
	i = *flag.String("i", "", "The IP address of the machine running a Chord node. The Chord client will join this node’s ring. Represented as an ASCII string (e.g., 128.8.126.63). Must be specified if --jp is specified.")

	p = *flag.Int("p", 0, "The identifier (ID) assigned to the Chord client which will override the ID computed by the SHA1 sum of the client’s IP address and port number. Represented as a string of 40 characters matching [0-9a-fA-F]. Optional parameter.")
	jp = *flag.Int("jp", 0, "The port that an existing Chord node is bound to and listening on. The Chord client will join this node’s ring. Represented as a base-10 integer. Must be specified if --ja is specified.")
	ts = *flag.Int("ts", 0, "The time in milliseconds between invocations of ‘stabilize’. Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000].")
	tff = *flag.Int("tff", 0, "The time in milliseconds between invocations of ‘fix fingers’. Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000].")
	tcp = *flag.Int("tcp", 0, "The time in milliseconds between invocations of ‘check predecessor’. Represented as a base-10 integer. Must be specified, with a value in the range of [1,60000].")
	r = *flag.Int("r", 0, "The number of successors maintained by the Chord client. Represented as a base-10 integer. Must be specified, with a value in the range of [1,32].")

	flag.Parse()

	checkEmptyFlags()
}

func checkEmptyFlags() {
	exit := false
	allMandatory := []string{"a", "p", "ts", "tff", "tcp", "r"}

	if isFlagPassed("jp") {
		if !isFlagPassed("ja") {
			print("Since -jp was set you need to specify ja.")
			exit = true
		}
	}

	if isFlagPassed("ja") {
		if !isFlagPassed("jp") {
			print("Since -ja was set you need to specify ja.")
			exit = true
		}
	}

	for _, flag := range allMandatory {
		if !isFlagPassed(flag) {
			print("No value was set for: -" + flag + ".\n")
			exit = true
		}
	}
	if exit {
		os.Exit(1)
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
