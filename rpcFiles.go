// This file is for rpc method for everything file related
package main

import (
	"io/ioutil"
	"log"
	"os"
)

// Cant send plain array in go over rpc
type MultipleFiles struct {
	Files []BasicFile
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

// Receives rpc request to check if Files exist on this node (in bucket)
func (ring *Ring) FileExist(key []byte, reply *bool) error {
	myString := string(key[:])
	_, found := myNode.PrimaryBucket[myString]
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
	myNode.PrimaryBucket[myString] = file.Filename

	// Create new file to store information in
	newFile, err := os.Create("primaryBucket/" + file.Filename)
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
// If the node is missing Files in backup we send it
func (node *BasicNode) rpcUpdateBackupBucketOf(bucketOf *Node) {
	// Send out backup bucket to see what Files the node is missing
	var response MultipleFiles
	sendBucket := bucketOf.PrimaryBucket

	// If bucket is not empty
	if len(sendBucket) > 0 {
		err := call(node.Address, "Ring.UpdateBackupBucketOf", sendBucket, &response)
		if err != nil {
			log.Println("Method: UpdateBackupBucketOf Error: ", err)
		}
	}

	// If the list of Files the node sent to us is not empty we send these Files
	responseFile := response.Files
	if len(responseFile) > 0 {
		for _, file := range responseFile {
			// Read file
			fileContent, err := os.ReadFile("primaryBucket/" + file.Filename)
			if err != nil {
				log.Println("Method: os.ReadFile Error:", err)
				continue
			}

			// Send file to node
			sendFile := BasicFile{Filename: file.Filename, Key: file.Key, FileContent: fileContent}
			var ret *bool
			err = call(node.Address, "Ring.SendBackupFile", sendFile, &ret)
			if err != nil {
				log.Println("Method: SendBackupFile Error: ", err)
			}
		}
	}

}

// Checks what Files we are missing in our backupBucket.
// Replies with the Files needed.
func (ring *Ring) UpdateBackupBucketOf(backupBucket map[string]string, reply *MultipleFiles) error {
	// First check if any Files needs to be moved to primary bucket
	// Why? If predecessor died you should now own their previous keys
	for key, fileName := range myNode.BackupBucket {
		// If key needs to be moved from backup to primary
		if between([]byte(key), myNode.Predecessor.Id, myNode.Id) {
			fileContent, err := os.ReadFile("backupBucket/" + fileName)
			if err != nil {
				continue
			}
			myFile := BasicFile{Filename: fileName, Key: []byte(key), FileContent: fileContent}
			myBasicNode := BasicNode{Address: myNode.Address, Id: myNode.Id}
			myBasicNode.rpcStoreFile(myFile)
			os.Remove("backupBucket/" + fileName)
		}
	}

	// Get all Files in backup dir
	previousFiles, err := ioutil.ReadDir("backupBucket/")
	if err != nil {
		return err
	}
	// Delete dead Files (Files that new node own)
	for _, file := range previousFiles {
		if !contains(backupBucket, file.Name()) {
			err = os.Remove("backupBucket/" + file.Name())
			if err != nil {
				continue
			}
		}
	}

	var neededFiles []BasicFile
	for key, fileName := range backupBucket {
		if _, err := os.Stat(fileName); err != nil { // File does not exist
			neededFiles = append(neededFiles, BasicFile{Filename: fileName, Key: []byte(key)})
		}
	}
	reply.Files = neededFiles
	return nil
}

// Receives rpc request to save backup file
func (ring *Ring) SendBackupFile(file BasicFile, reply *bool) error {
	_, err := os.Stat(file.Filename)
	// If file did exist on this pc return with error
	if err == nil {
		return err
	}

	// Add file to bucket
	myString := string(file.Key)
	myNode.BackupBucket[myString] = file.Filename

	// Create new file to store information in
	newFile, err := os.Create("backupBucket/" + file.Filename)
	if err != nil {
		return err
	}
	defer newFile.Close()

	// Store file information in the new file
	_, err = newFile.Write(file.FileContent)
	if err != nil {
		return err
	}
	*reply = true
	return nil
}

// Checks if map contains string
func contains(s map[string]string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
