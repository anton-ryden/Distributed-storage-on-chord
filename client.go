package main

import (
	"log"
	"net"
	"strconv"
)

// Set with information from arguments
var a, ja, i *string
var p, jp, ts, tff, tcp, r *int

// Main
func main() {
	// Setup arguments
	setupArguments()

	// Error handling in arguments file, so we only need to check if ja is set
	addr := NodeAddress(*a + ":" + strconv.Itoa(*p))
	node := Node{address: addr}
	if *ja == "" {
		node.create()
	} else {
		addr := NodeAddress(*ja + ":" + strconv.Itoa(*jp))
		joinNode := Node{address: addr}
		node.join(joinNode)
	}

	// Check if valid address
	checkIPAddress(*a)

	// Start tcp server with port from argument
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(*p))
	defer ln.Close()
	if err != nil {
		log.Fatalln(err)
	}
}

// Checks if ip address is valid
func checkIPAddress(ip string) {
	if net.ParseIP(ip) == nil {
		log.Fatalf("{\tIP Address: %s - Invalid\n", ip)
	} else {
		log.Printf("\tIP Address: %s - Valid\n", ip)
	}
}
