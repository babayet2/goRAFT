# goRAFT
An implementation of the RAFT leader election sub-protocol in go, spawning multiple processes and communicating over RPC. A description of the RAFT Consensus Algorithm can be found here: https://raft.github.io/

A bash script has been provided for testing on linux systems. Requires `num_nodes` open ports starting at 1200, defaulting at 10.

`Usage: raft.sh [num_nodes]`
