# Distributed-storage-on-chord
A simple fault-tolerant storage system based on the chord protocol that uses TLS and file encryption.

## How to use
The system is started by compiling with the command ```go build```. This will generate the binary file named chordClient. Run this binary file with necessary arguments

#### Commands
There are four commands that are usable on each Chord client:

- ```Lookup```: Takes as input the name of a file to be searched (e.g., “Hello.txt”). 
The Chord client takes this string, hashes it to a key in the identifier space, 
and performs a search for the node that is the successor to the key 
(i.e., the owner of the key). 
The Chord client then outputs that node’s identifier, IP address and port.
- ```StoreFile```: Takes the location of a file on a local disk, then performs a "Lookup". 
Once the correct place of the file is found, the file gets encrypted and, if the encryption was successful, uploaded to the Chord ring.
- ```PrintState```: requires no input. The Chord client outputs its local state information at the current time, which consists of:
1. The Chord client’s own node information, predecessor, stored files and backup files (if any exists)
2. The node information for all nodes in the successor list
3. The node information for all nodes in the finger table
   where “node information” corresponds to the identifier, IP address, and port for a given node.
- ```DownloadFile```: Takes the name of a file that exists on the chord ring and downloads it from the node holding
the file. It then decrypts the file if the Chord clients own node has the correct key.

##### Structure
There are six folders that exists on each Chord client

- ```upload```: Should contain files that are to be uploaded to the Chord ring.
- ```primaryBucket```: Contains all files that's been uploaded to the local Chord client from other clients
- ```download```: Contains all files downloaded from the Chord ring
- ```crypto-key```: Contains the local Chord clients own private key for encrypting and decrypting
- ```certs```: Contains all the necessary certificates for the local Chord client to communicate with other clients through TLS
- ```backupBucket```: Contains backups of files that are already in the Chord ring in order for it to be fault-tolerant

##### Scripts
There are three scripts in total, two of which runs automatically when starting a Chord client

- ```generate-ca-cert.sh```: Located in the ```certs``` folder. It generates a CA certificate and key that is to be used to sign the certificates of any Chord client that wants to join the system.
This script only needs to be executed once by the user that starts the system. All clients must have their certificates signed by the same CA for it to work.
- ```generate-cert.sh```: Located in the ```certs``` folder. It generates both server and client certificates and keys for the local Chord client. CA certificate
and key needs to be present in the certs folder before running since each certificate will be signed by the CA. This script runs automatically when starting a new local Chord client.
- ```generate-key.sh```: Located in the ```crypto-key``` folder. It generates a symmetric 128-bit AES key to be used by the local Chord client when uploading to or downloading
files from the Chord ring. This script runs automatically when starting a new local Chord client.

## Arguments
Current argument list:
```
-h  --help  Print help information

-a  --a     The IP address that the Chord client will bind to, as well as
            advertise to other nodes. Represented as an ASCII string (e.g.,
            128.8.126.63)
            
-p  --p     The port that the Chord client will bind to and listen on.
            Represented as a base-10 integer. Must be specified.
            
--jp        The port that an existing Chord node is bound to and listening on.
            The Chord client will join this node’s ring. Represented as
            a base-10 integer. Must be specified if --ja is specified.
            
--ja        The IP address of the machine running a Chord node. The Chord
            client will join this node’s ring. Represented as an ASCII
            string (e.g., 128.8.126.63). Must be specified if --jp is
            specified.
            
--ts        The time in milliseconds between invocations of ‘stabilize’.
            Represented as a base-10 integer. Must be specified, with a value
            in the range of [1,60000].
            
--tff       The time in milliseconds between invocations of ‘fix
            fingers’. Represented as a base-10 integer. Must be specified,
            with a value in the range of [1,60000].
            
--tcp       The time in milliseconds between invocations of ‘check
            predecessor’. Represented as a base-10 integer. Must be
            specified, with a value in the range of [1,60000].
            
-r  --r     The number of successors maintained by the Chord client.
            Represented as a base-10 integer. Must be specified, with a value
            in the range of [1,32].
            
-i  --i     The identifier (ID) assigned to the Chord client which will
            override the ID computed by the SHA1 sum of the client’s IP
            address and port number. Represented as a string of 40 characters
            matching [0-9a-fA-F]. Optional parameter.
```

