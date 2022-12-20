// This file is for all rpc methods for everything node related
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"
	"time"
)

type Ring int

// Amount of ms before timeout of connection
const timeoutMs = 1000

// Used in find rpc call
type RpcReply struct {
	Found bool
	Node  BasicNode
}

var clientConfig *tls.Config

// Sets the tls config, so it can communicate with other nodes with tls
func setConfig() {
	// Read ca's cert
	caCert, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		fmt.Println(caCert)
		os.Exit(0)
	}

	// Create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		fmt.Println(err)
		os.Exit(0)
	}

	// Read client cert
	clientCert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}

	clientConfig = &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}
}

// Function that performs the sends rpc request and receive reply.
func call(address string, method string, request interface{}, response interface{}) error {
	// Create connection with a timeout value
	ctx, cancel := context.WithTimeout(context.Background(), timeoutMs*time.Millisecond)
	d := tls.Dialer{
		Config: clientConfig,
	}
	conn, err := d.DialContext(ctx, "tcp", address)
	cancel() // Ensure cancel is always called

	if err != nil {
		fmt.Println("Method: tls.Dial Error: ", err)
		return err
	}

	defer conn.Close()
	client := rpc.NewClient(conn)
	defer client.Close()
	// Make request
	err = client.Call(method, request, response)
	return err
}

// Sends rpc request that copies(updates) successor information from immediate successor(or next successor if fail)
func (node *Node) rpcCopySuccessor() {
	var response *Node
	var err error
	var suc *BasicNode
	var i int
	// If call fails try next successor in list until list is empty
	// When list becomes empty set successor to myself
	for i, suc = range node.Successor {
		if i != 0 {
			fmt.Println("Successor is dead. Trying next successor in list:\nNode:\n\tAddress:", node.Successor[0].Address, "\n\tId:\t", string(node.Successor[0].Id))
		}
		err := call(suc.Address, "Ring.CopySuccessor", false, &response)
		if err == nil { // If call did not generate error we found the immediate successor
			break
		}
	}

	// If no alive successor was found set successor to itself
	if err != nil {
		myself := BasicNode{Address: node.Address, Id: node.Id}
		node.Successor[0] = &myself
		fmt.Println("All successors in list is dead new succesor is:\nNode:\n\tAddress:", node.Successor[0].Address, "\n\tId:\t", string(node.Successor[0].Id))
		return
	}

	// Append new successor information to list and ensure max r successor in list
	node.Successor = append([]*BasicNode{suc}, response.Successor...)
	sucLen := len(node.Successor)
	if sucLen > *r {
		node.Successor = node.Successor[:sucLen-1]
	}

}

// Receives request from rpcCopySuccessor. Returns this clients successor list
func (ring *Ring) CopySuccessor(inBool bool, reply *Node) error {
	*reply = Node{Successor: myNode.Successor}
	return nil
}

// Sends rpc request to get predecessor of the node getMyPredecessor
func (node *Node) rpcGetPredecessorOf(getMyPredecessor *BasicNode) BasicNode {
	var response *BasicNode
	err := call(getMyPredecessor.Address, "Ring.GetPredecessor", false, &response)
	if err != nil {
		fmt.Println("Method: Ring.GetPredecessor Error: ", err)
	}
	return *response
}

// Receives rpc request to get predecessor of this node.
func (ring *Ring) GetPredecessor(inBool bool, reply *BasicNode) error {
	// Reply with empty node basicNode if predecessor is nil
	if myNode.Predecessor == nil {
		*reply = BasicNode{}
	} else {
		*reply = *myNode.Predecessor
	}
	return nil
}

// Sends rpc request to update the immediate successor of updateMySuccessor to node
func (node *Node) rpcUpdateSuccessorOf(updateMySuccessor *BasicNode) {
	var response bool
	send := BasicNode{Address: node.Address, Id: node.Id}
	err := call(updateMySuccessor.Address, "Ring.UpdateSuccessorOf", send, &response)
	if err != nil {
		fmt.Println("Method: Ring.UpdateSuccessorOf Error: ", err)
	}
}

// Receives rpc request to update successor of this node
func (ring *Ring) UpdateSuccessorOf(newSuccessor *BasicNode, reply *bool) error {
	// Append new successor in the begging of successor array
	oldSuccessors := myNode.Successor
	myNode.Successor = append([]*BasicNode{newSuccessor}, oldSuccessors...)

	// Check if array length need to be changed
	sucLen := len(myNode.Successor)
	if sucLen > *r {
		myNode.Successor = myNode.Successor[:sucLen-1]
	}

	// Tell user we got a new immediate successor
	fmt.Println("New immediate successor:\n",
		"Node:\n",
		"\tAddress:", myNode.Successor[0].Address, "\n",
		"\tId:\t", string(myNode.Successor[0].Id))

	return nil
}

// Sends rpc request to notify node of notifyOfMe
func (node *BasicNode) rpcNotifyOf(notifyOfMe *Node) {
	var response bool
	sendNode := BasicNode{Address: notifyOfMe.Address, Id: notifyOfMe.Id}
	err := call(node.Address, "Ring.NotifyOf", sendNode, &response)
	if err != nil {
		fmt.Println("Method: Ring.NotifyOf Error: ", err)
	}
}

// Receives rpc request to notify this node of a new predecessor
func (ring *Ring) NotifyOf(notifyOf BasicNode, reply *bool) error {
	myNode.notify(notifyOf)
	return nil
}

// Send rpc request to tell check if node is alive
func (node *BasicNode) rpcIsAlive() bool {
	var response bool
	err := call(node.Address, "Ring.CheckAlive", false, &response)
	// If no error the node is alive
	if err != nil {
		fmt.Println("Address: ", node.Address, " Id: ", string(node.Id), " is no longer alive"+
			", Predecessor is now nil")
		return false
	}
	return response
}

// Receives rpc request to check if node is alive
func (ring *Ring) CheckAlive(inBool bool, reply *bool) error {
	*reply = true
	return nil
}

// Send rpc request to check for id of successor
func (node *BasicNode) rpcFindSuccessor(id []byte) (bool, BasicNode) {
	var response RpcReply
	err := call(node.Address, "Ring.FindSuccessor", &id, &response)
	if err != nil {
		fmt.Println("Method: Ring.FindSuccessor Error: ", err)
	}
	return response.Found, response.Node
}

// Receives rpc request to search for id in successor
func (ring *Ring) FindSuccessor(id []byte, reply *RpcReply) error {
	found, retNode := myNode.findSuccessor(id)
	reply.Found = found
	reply.Node = retNode
	return nil
}

// Initializes a port to listen for rpc calls
func initListen() net.Listener {

	rpc.Register(new(Ring))

	// read ca's cert, verify to client's certificate
	caPem, err := ioutil.ReadFile("certs/ca.crt")
	if err != nil {
		log.Fatal(err)
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPem) {
		log.Fatal(err)
	}

	// read server cert & key
	serverCert, err := tls.LoadX509KeyPair("certs/server.crt", "certs/server.key")
	if err != nil {
		log.Fatal(err)
	}

	// configuration of the certificate what we want to
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}

	ipPort := strings.Split(myNode.Address, ":")
	listener, err := tls.Listen("tcp", ":"+ipPort[1], config)
	if err != nil {
		log.Fatalln(err)
	}
	return listener
}

// Checks for incoming rpc calls
func listen(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Listen accept error: " + err.Error())
	}
	defer conn.Close()
	rpc.ServeConn(conn)
}
