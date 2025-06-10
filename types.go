// types.go
package main

import "time"

// Message is the generic wrapper for all network communication.
type Message struct {
	From    int
	To      int
	Payload interface{}
	// In a real system, you would add a cryptographic signature
}

// Block represents a set of transactions.
type Block struct {
	ID        [32]byte
	Proposer  int
	View      int
	Timestamp time.Time
	Data      []byte
}

// ProposalPacket is a single encoded packet from a block.
type ProposalPacket struct {
	View    int
	BlockID [32]byte
	Hash    [32]byte
	Packet  []byte
}

// VoteMsg is a vote for a specific block hash.
type VoteMsg struct {
	View    int
	BlockID [32]byte
	Hash    [32]byte
}

// NewViewMsg is a vote to move to a new view.
type NewViewMsg struct {
	RequestedView int
}

// PrePrepareMsg is the first phase of PBFT.
type PrePrepareMsg struct {
	View  int
	Block Block
}

// PrepareMsg is the second phase of PBFT.
type PrepareMsg struct {
	View    int
	BlockID [32]byte
	Hash    [32]byte
}

// CommitMsg is the third phase of PBFT.
type CommitMsg struct {
	View    int
	BlockID [32]byte
	Hash    [32]byte
}
