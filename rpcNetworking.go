package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
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
		log.Fatal(caCert)
	}

	// Create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal(err)
	}

	// Read client cert
	clientCert, err := tls.LoadX509KeyPair("certs/client.crt", "certs/client.key")
	if err != nil {
		log.Fatal(err)
	}

	clientConfig = &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}
}

// Function that performs the sends rpc request and receive reply.
func call(address string, method string, request interface{}, response interface{}) error {
	// Create connection with a timeout value
	conn, err := tls.Dial("tcp", address, clientConfig)

	if err != nil {
		log.Fatalf("client: dial: %s", err)
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
			log.Println("Successor is dead. Trying next successor in list:\nNode:\n\tAddress:", node.Successor[0].Address, "\n\tId:\t", string(node.Successor[0].Id))
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
		log.Println("All successors in list is dead new succesor is:\nNode:\n\tAddress:", node.Successor[0].Address, "\n\tId:\t", string(node.Successor[0].Id))
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
		log.Println("Method: Ring.GetPredecessor Error: ", err)
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
		log.Fatalln("Method: Ring.UpdateSuccessorOf Error: ", err)
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
	log.Println("New immediate successor:\n",
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
		log.Println("Method: Ring.NotifyOf Error: ", err)
	}
}

// Receivec rpc request to notify this node of a new predecessor
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
		log.Println("Address: ", node.Address, " Id: ", string(node.Id), " is no longer alive"+
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
		log.Println("Method: Ring.FindSuccessor Error: ", err)
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

// Send rpc request to check if file exist on node
func (node *BasicNode) rpcFileExist(key []byte) bool {
	var response *bool
	err := call(node.Address, "Ring.FileExist", key, &response)
	if err != nil {
		log.Println("Method: Ring.FileExist Error: ", err)
	}
	return *response
}

// Receives rpc request to check if files exist on this node (in bucket)
func (ring *Ring) FileExist(key []byte, reply *bool) error {
	myString := string(key[:])
	_, found := myNode.Bucket[myString]
	if found {
		*reply = true
	} else {
		*reply = false
	}
	return nil
}

// Send rpc request to store file on node
func (node *BasicNode) rpcStoreFile(file BasicFile) bool {
	var response *bool
	err := call(node.Address, "Ring.StoreFile", file, &response)
	if err != nil {
		log.Println("Method: Ring.StoreFile Error: ", err)
	}
	return *response
}

// Receives rpc request to store file on this node
func (ring *Ring) StoreFile(file BasicFile, reply *bool) error {
	_, err := os.Stat(file.Filename)
	// If file did exist on this pc return with error
	if err == nil {
		*reply = false
		return err
	}

	// Add file to bucket
	myString := string(file.Key)
	myNode.Bucket[myString] = file.Filename

	// Create new file to store information in
	newFile, err := os.Create(file.Filename)
	if err != nil {
		*reply = false
		return err
	}
	defer newFile.Close()

	// Store file information in the new file
	_, err = newFile.Write(file.FileContent)
	if err != nil {
		*reply = false
		return err
	}
	*reply = true
	return nil
}

// Sends backup bucket to node, the response is the file that node does not have in backup that need to be sent
// If the node is missing files in backup we send it
func (node *BasicNode) rpcUpdateBackupBucketOf(backupBucketNode *Node) {
	// Send out backup bucket to see what files the node is missing
	var response []BasicFile
	err := call(node.Address, "Ring.UpdateBackupBucketOf", backupBucketNode.BackupBucket, &response)
	if err != nil {
		log.Println("Method: UpdateBackupBucketOf Error: ", err)
	}

	// If the list of files node sent us is not empty we send these files
	if len(response) > 0 {
		for _, file := range response {
			node.rpcStoreFile(file)
		}
	}

}

func (ring *Ring) UpdateBackupBucketOf(backupBucket map[string]string, reply []BasicFile) error {
	for _, temp := range backupBucket {
		temp = temp
	}
	reply = nil
	return nil
}

func (node *Node) rpcSendBackupTo(toNode BasicNode) {
	var fileArray []BasicFile
	for key, fileName := range node.Bucket {
		if _, err := os.Stat(fileName); err == nil {
			fileContent, err := os.ReadFile(fileName)
			if err != nil {
				log.Println("Method: os.ReadFile Error:", err)
				return
			}
			fileArray = append(fileArray, BasicFile{Filename: fileName, Key: []byte(key), FileContent: fileContent})
		}
	}
	var response *bool
	err := call(toNode.Address, "Ring.SendBackupTo", &fileArray, &response)
	if err != nil {
		log.Println("Method: Ring.StoreFile Error: ", err)
	}
}

func (ring *Ring) SendBackupTo(files []BasicFile, reply *bool) error {
	for _, file := range files {

		if _, err := os.Stat(file.Filename); err == nil {
			// Add file to bucket
			myString := string(file.Key)
			myNode.BackupBucket[myString] = file.Filename

			// Create new file to store information in
			newFile, err := os.Create(file.Filename)
			if err != nil {
				*reply = false
				return err
			}
			defer newFile.Close()

			// Store file information in the new file
			_, err = newFile.Write(file.FileContent)
			if err != nil {
				*reply = false
				return err
			}
		}
	}
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

	/*tcpAddr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(*p))
	if err != nil {
		log.Fatalln(err)
	}*/

	listener, err := tls.Listen("tcp", myNode.Address, config)
	if err != nil {
		log.Fatalln(err)
	}
	return listener
}

// Checks for incoming rpc calls
func listen(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		log.Println("Listen accept error: " + err.Error())
	}
	defer conn.Close()
	rpc.ServeConn(conn)
}
