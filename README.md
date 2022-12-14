# Distributed-storage-on-chord

## How to use

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

