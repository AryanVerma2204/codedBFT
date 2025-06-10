// protocol_codedbft.go
package main

import (
	"crypto/sha256"
	"fmt"
	"log"
	"sync"
	"time"
)

type CodedBFTProtocol struct {
	node                        *Node
	speculativeExecutionEnabled bool
	metrics                     Metrics
	stopChan                    chan struct{}
	mux                         sync.RWMutex
	currentView                 int
	decoders                    map[[32]byte]*Decoder
	votes                       map[[32]byte]map[int]struct{}
	committed                   map[[32]byte]bool
}

func (p *CodedBFTProtocol) Init(node *Node) {
	p.node = node
	p.stopChan = make(chan struct{})
	p.decoders = make(map[[32]byte]*Decoder)
	p.votes = make(map[[32]byte]map[int]struct{})
	p.committed = make(map[[32]byte]bool)
}

func (p *CodedBFTProtocol) Propose(block Block) {
	view := p.GetCurrentView()
	if p.node.id != view%p.node.config.NumNodes { return } // Only leaders propose

	hash := sha256.Sum256(block.Data)
	encoder, err := NewEncoder(block.Data, p.node.config.PacketSize)
	if err != nil { return }

	numPackets := len(block.Data)/p.node.config.PacketSize + 20 // Send with overhead
	for i := 0; i < numPackets; i++ {
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
		if payload.View < p.currentView || p.committed[payload.Hash] { return }
		if _, ok := p.decoders[payload.BlockID]; !ok {
			p.decoders[payload.BlockID] = NewDecoder()
		}
		if p.decoders[payload.BlockID].AddPacket(payload.Packet) {
			decodedData, err := p.decoders[payload.BlockID].GetDecodedData()
			delete(p.decoders, payload.BlockID)
			if err != nil || sha256.Sum256(decodedData) != payload.Hash { return }
			
			if p.speculativeExecutionEnabled {
				vote := &VoteMsg{View: payload.View, BlockID: payload.BlockID, Hash: payload.Hash}
				p.node.net.Broadcast(p.node.id, vote)
				p.metrics.AddBytesSent(68)
			}
		}
	case *VoteMsg:
		if payload.View < p.currentView || p.committed[payload.Hash] { return }
		if _, ok := p.votes[payload.Hash]; !ok {
			p.votes[payload.Hash] = make(map[int]struct{})
		}
		p.votes[payload.Hash][msg.From] = struct{}{}
		
		quorum := 2*p.node.config.NumFaulty + 1
		if len(p.votes[payload.Hash]) >= quorum {
			p.committed[payload.Hash] = true
			log.Printf("[Node %d] CodedBFT: Committed block %x", p.node.id, payload.Hash[:4])
			p.metrics.AddCommit(time.Since(time.Now())) // Simplified
		}
	}
}

func (p *CodedBFTProtocol) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	timeout := time.NewTimer(p.node.config.ConsensusTimeout)
	for {
		select {
		case msg := <-p.node.net.GetChannel(p.node.id):
			timeout.Reset(p.node.config.ConsensusTimeout)
			p.HandleMessage(msg)
		case <-timeout.C:
			p.mux.Lock()
			p.currentView++
			p.metrics.IncViewChanges()
			log.Printf("[Node %d] CodedBFT: Timeout, moving to view %d", p.node.id, p.currentView)
			p.mux.Unlock()
			timeout.Reset(p.node.config.ConsensusTimeout)
		case <-p.stopChan:
			return
		}
	}
}
func (p *CodedBFTProtocol) Stop() { close(p.stopChan) }
func (p *CodedBFTProtocol) GetMetrics() *Metrics { return &p.metrics }
func (p *CodedBFTProtocol) GetCurrentView() int { p.mux.RLock(); defer p.mux.RUnlock(); return p.currentView }
