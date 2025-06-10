// protocol_pbft.go
package main

import (
	"crypto/sha256"
	"log"
	"sync"
	"time"
)

type PBFTProtocol struct {
	node        *Node
	metrics     Metrics
	stopChan    chan struct{}
	mux         sync.RWMutex
	currentView int
	prepares    map[[32]byte]map[int]struct{}
	commits     map[[32]byte]map[int]struct{}
	committed   map[[32]byte]bool
}

func (p *PBFTProtocol) Init(node *Node) {
	p.node = node
	p.stopChan = make(chan struct{})
	p.prepares = make(map[[32]byte]map[int]struct{})
	p.commits = make(map[[32]byte]map[int]struct{})
	p.committed = make(map[[32]byte]bool)
}

func (p *PBFTProtocol) Propose(block Block) {
	view := p.GetCurrentView()
	if p.node.id != view%p.node.config.NumNodes { return }
	
	payload := &PrePrepareMsg{View: view, Block: block}
	p.node.net.Broadcast(p.node.id, payload)
	p.metrics.AddBytesSent(len(block.Data))
}

func (p *PBFTProtocol) HandleMessage(msg Message) {
	p.mux.Lock()
	defer p.mux.Unlock()

	switch payload := msg.Payload.(type) {
	case *PrePrepareMsg:
		if payload.View < p.currentView { return }
		hash := sha256.Sum256(payload.Block.Data)
		prepare := &PrepareMsg{View: payload.View, BlockID: payload.Block.ID, Hash: hash}
		p.node.net.Broadcast(p.node.id, prepare)
		p.metrics.AddBytesSent(68)
	case *PrepareMsg:
		if payload.View < p.currentView || p.committed[payload.Hash] { return }
		if _, ok := p.prepares[payload.Hash]; !ok { p.prepares[payload.Hash] = make(map[int]struct{}) }
		p.prepares[payload.Hash][msg.From] = struct{}{}
		
		quorum := 2 * p.node.config.NumFaulty
		if len(p.prepares[payload.Hash]) >= quorum {
			commit := &CommitMsg{View: payload.View, BlockID: payload.BlockID, Hash: payload.Hash}
			p.node.net.Broadcast(p.node.id, commit)
			p.metrics.AddBytesSent(68)
		}
	case *CommitMsg:
		if payload.View < p.currentView || p.committed[payload.Hash] { return }
		if _, ok := p.commits[payload.Hash]; !ok { p.commits[payload.Hash] = make(map[int]struct{}) }
		p.commits[payload.Hash][msg.From] = struct{}{}
		
		quorum := 2*p.node.config.NumFaulty + 1
		if len(p.commits[payload.Hash]) >= quorum {
			p.committed[payload.Hash] = true
			log.Printf("[Node %d] PBFT: Committed block %x", p.node.id, payload.Hash[:4])
			p.metrics.AddCommit(time.Since(time.Now())) // Simplified
		}
	}
}
func (p *PBFTProtocol) Start(wg *sync.WaitGroup) {
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
			log.Printf("[Node %d] PBFT: Timeout, moving to view %d", p.node.id, p.currentView)
			p.mux.Unlock()
			timeout.Reset(p.node.config.ConsensusTimeout)
		case <-p.stopChan:
			return
		}
	}
}
func (p *PBFTProtocol) Stop() { close(p.stopChan) }
func (p *PBFTProtocol) GetMetrics() *Metrics { return &p.metrics }
func (p *PBFTProtocol) GetCurrentView() int { p.mux.RLock(); defer p.mux.RUnlock(); return p.currentView }
