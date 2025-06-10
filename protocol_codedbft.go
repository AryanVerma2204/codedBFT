// protocol_codedbft.go
package main

import (
	"crypto/sha256"
	"log"
	"sync"
	"time"
)

type CodedBFTProtocol struct {
	node                        *Node
	speculativeExecutionEnabled bool
	metrics                     Metrics
	stopChan                    chan struct{}

	mux         sync.RWMutex
	currentView int
	decoder     *Decoder
	votes       map[[32]byte]map[int]struct{}
	committed   map[[32]byte]bool
}

func (p *CodedBFTProtocol) Init(node *Node) {
	p.node = node
	p.stopChan = make(chan struct{})
	p.committed = make(map[[32]byte]bool)
	p.votes = make(map[[32]byte]map[int]struct{})
}

func (p *CodedBFTProtocol) Propose(blockData []byte) {
	p.mux.Lock()
	view := p.currentView
	p.mux.Unlock()

	block := Block{
		ID:        sha256.Sum256(blockData), // Using data hash as ID for simplicity
		Proposer:  p.node.id,
		View:      view,
		Timestamp: time.Now(),
		Data:      blockData,
	}
	hash := sha256.Sum256(block.Data)

	encoder, err := NewEncoder(block.Data, p.node.config.PacketSize)
	if err != nil {
		log.Printf("[Node %d] Encoder error: %v", p.node.id, err)
		return
	}

	// In a real system, this would continue until a commit is observed
	for i := 0; i < len(block.Data)/p.node.config.PacketSize+20; i++ {
		packet := encoder.GetEncodedPacket()
		payload := &ProposalPacket{View: view, BlockID: block.ID, Hash: hash, Packet: packet}
		p.node.net.Broadcast(p.node.id, payload)
		p.metrics.AddBytesSent(len(packet))
	}
}

func (p *CodedBFTProtocol) HandleMessage(msg Message) {
	p.mux.Lock()
	defer p.mux.Unlock()

	switch payload := msg.Payload.(type) {
	case *ProposalPacket:
		if payload.View < p.currentView {
			return
		}
		if p.decoder == nil {
			p.decoder = NewDecoder()
		}
		if p.decoder.AddPacket(payload.Packet) {
			decodedData, err := p.decoder.GetDecodedData()
			if err != nil || sha256.Sum256(decodedData) != payload.Hash {
				return // Invalid decode
			}
			p.decoder = nil // Reset for next block

			// This is the speculative execution step
			if p.speculativeExecutionEnabled {
				vote := &VoteMsg{View: payload.View, BlockID: payload.BlockID, Hash: payload.Hash}
				p.node.net.Broadcast(p.node.id, vote)
				p.metrics.AddBytesSent(68) // Approx size of VoteMsg
			}
		}

	case *VoteMsg:
		if payload.View < p.currentView {
			return
		}
		if _, ok := p.votes[payload.Hash]; !ok {
			p.votes[payload.Hash] = make(map[int]struct{})
		}
		p.votes[payload.Hash][msg.From] = struct{}{}

		quorum := 2*p.node.config.NumFaulty + 1
		if len(p.votes[payload.Hash]) >= quorum && !p.committed[payload.Hash] {
			p.committed[payload.Hash] = true
			latency := time.Since(time.Unix(0, 0)).Milliseconds() // Simplified latency
			p.metrics.AddCommit(float64(latency))
			log.Printf("[Node %d] CodedBFT: Committed block %x in view %d", p.node.id, payload.Hash[:6], payload.View)
		}
	}
}

func (p *CodedBFTProtocol) Start() {
	defer p.node.wg.Done()
	for {
		select {
		case msg := <-p.node.net.GetChannel(p.node.id):
			p.HandleMessage(msg)
		case <-p.stopChan:
			return
		}
	}
}

func (p *CodedBFTProtocol) Stop()       { close(p.stopChan) }
func (p *CodedBFTProtocol) GetMetrics() Metrics { return p.metrics }
