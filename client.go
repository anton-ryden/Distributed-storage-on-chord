package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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
	listener := initListen()

	// Init go routines that run fixFingers, Stabilize and checkPredecessor
	initRoutines()

	// Go routine for scanning input from user
	go scan()

	//Main for loop
	for {
		listen(listener) // Go routine for listening to traffic
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
			if len(sliceText) == 2 {
				StoreFile(sliceText[1])
			} else {
				fmt.Println("No filepath was given")
			}
			break
		case "Lookup":
			if len(sliceText) == 2 {
				Lookup(sliceText[1])
			} else {
				fmt.Println("No filename was given")
			}
			break
		case "DownloadFile":
			if len(sliceText) == 2 {
				DownloadFile(sliceText[1])
			} else {
				fmt.Println("No filename was given")
			}
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

	if _, err := os.Stat("upload/" + filePath); err == nil {
		foundNode, found, hashed := Lookup(filename)
		fileContent, err := os.ReadFile("upload/" + filePath)
		if err != nil {
			fmt.Println("Method: os.ReadFile Error:", err)
			return
		}

		encryptedfileContent, keyFound := EncryptFileContent(fileContent)
		if !keyFound {
			fmt.Println("No encryption key was found, not sending file...")
			return
		}

		myFile := BasicFile{Filename: filename, Key: hashed, FileContent: encryptedfileContent}

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

// Downloads file with filename as name on the ring
func DownloadFile(filename string) {
	foundNode, found, hashed := Lookup(filename)

	if found {
		myFile := foundNode.rpcGetFile(hashed)
		dir := "download/"

		if _, err := os.Stat(dir); err != nil {
			// Create directory downloads/ if it does not exist and set its permissions
			os.MkdirAll(dir, os.ModePerm)
			err = os.Chmod(dir, 0777)
			if err != nil {
				log.Println("Error in DownloadFile:", err)
			}
		}

		// Create new file to store information in
		newFile, err := os.Create(filepath.Join(dir, filepath.Base(myFile.Filename)))
		if err != nil {
			log.Println("Error in DownloadFile:", err)
			return
		}

		// Setting permissions of the downloaded file
		err = os.Chmod(filepath.Join(dir, myFile.Filename), 0777)
		if err != nil {
			log.Println("Error in DownloadFile:", err)
		}

		defer newFile.Close()

		// Decrypt the contents of the file
		errOcc := false
		myFile.FileContent, errOcc = DecryptFileContent(myFile.FileContent)
		if errOcc {
			fmt.Println("Will not decrypt the file since an error occurred while decrypting")
		}

		// Store file information in the new file
		_, err = newFile.Write(myFile.FileContent)
		if err != nil {
			log.Println("Error in DownloadFile:", err)
			return
		}
		fmt.Println("\nFile was successfully downloaded. Located in " + dir + myFile.Filename)
	} else {
		fmt.Println("Since file was not found in the ring, it will not be downloaded")
	}
}

// Encrypts file content fileContent using GCM and a symmetric encryption key
func EncryptFileContent(fileContent []byte) ([]byte, bool) {
	foundEncryptionKey := false
	filePath := "crypto-key/sym-private.key"
	slicedPath := strings.Split(filePath, "/")
	filename := slicedPath[len(slicedPath)-1]

	if _, err := os.Stat(filepath.Join(slicedPath[0]+"/", filename)); err == nil {
		key, err := os.ReadFile("crypto-key/sym-private.key")
		if err != nil {
			log.Println("Error in EncryptFileContent:", err)
			return fileContent, foundEncryptionKey
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			log.Println("Error in EncryptFileContent:", err)
			return fileContent, foundEncryptionKey
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			log.Println("Error in EncryptFileContent:", err)
			return fileContent, foundEncryptionKey
		}

		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			log.Println("Error in EncryptFileContent:", err)
			return fileContent, foundEncryptionKey
		}

		ciphertext := gcm.Seal(nonce, nonce, fileContent, nil)
		foundEncryptionKey = true
		return ciphertext, foundEncryptionKey
	}
	return fileContent, foundEncryptionKey
}

// Decrypts file content fileContent using GCM and a symmetric encryption key
func DecryptFileContent(fileContent []byte) ([]byte, bool) {
	errOcc := false

	filePath := "crypto-key/sym-private.key"
	slicedPath := strings.Split(filePath, "/")
	filename := slicedPath[1]
	if _, err := os.Stat(filepath.Join(slicedPath[0]+"/", filename)); err == nil {
		key, err := os.ReadFile(filePath)
		if err != nil {
			log.Println("Error in DecryptFileContent:", err)
			errOcc = true
			return fileContent, errOcc
		}

		block, err := aes.NewCipher(key)
		if err != nil {
			log.Println("Error in DecryptFileContent:", err)
			errOcc = true
			return fileContent, errOcc
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			log.Println("Error in DecryptFileContent:", err)
			errOcc = true
			return fileContent, errOcc
		}

		nonce := fileContent[:gcm.NonceSize()]
		fileContentNew := fileContent[gcm.NonceSize():]

		decryptedContent, err := gcm.Open(nil, nonce, fileContentNew, nil)
		if err != nil {
			log.Println("Error in DecryptFileContent:", err)
			errOcc = true
			return fileContent, errOcc
		}

		return decryptedContent, errOcc
	}
	fmt.Println("Did not find an encryption key in crypto-key/sym-private.key, will not decrypt file content.")
	return fileContent, errOcc
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

	if len(myNode.PrimaryBucket) > 0 {
		fmt.Println("\n+-+-+-+-+-+-+ Fingertable info +-+-+-+-+-+--+")
		for _, bucketEntry := range myNode.PrimaryBucket {
			fmt.Println("\tFinger node: ", i, "\tID: ", bucketEntry)
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
