// node.go
package main

import "sync"

// ConsensusProtocol defines the interface that all protocols (PBFT, CodedBFT) must implement.
type ConsensusProtocol interface {
	Init(node *Node)
	Propose(block []byte)
	HandleMessage(msg Message)
	Start()
	Stop()
	GetMetrics() Metrics
}

// Node is a wrapper that holds a specific consensus protocol implementation.
type Node struct {
	id      int
	config  ExperimentConfig
	net     *SimulatedNetwork
	wg      *sync.WaitGroup
	stopped bool
	mux     sync.Mutex

	protocol ConsensusProtocol
}

func NewNode(id int, config ExperimentConfig, net *SimulatedNetwork) *Node {
	n := &Node{
		id:     id,
		config: config,
		net:    net,
	}

	switch config.Protocol {
	case ProtoCodedBFT:
		n.protocol = &CodedBFTProtocol{speculativeExecutionEnabled: true}
	case ProtoCodedBFTNoSpec:
		n.protocol = &CodedBFTProtocol{speculativeExecutionEnabled: false}
	case ProtoPBFT:
		n.protocol = &PBFTProtocol{}
	}
	n.protocol.Init(n)
	return n
}

func (n *Node) Start(wg *sync.WaitGroup) {
	n.wg = wg
	n.protocol.Start()
}

func (n *Node) Stop() {
	n.mux.Lock()
	n.stopped = true
	n.mux.Unlock()
	n.protocol.Stop()
}

func (n *Node) IsStopped() bool {
	n.mux.Lock()
	defer n.mux.Unlock()
	return n.stopped
}

// Delegate methods to the underlying protocol implementation
func (n *Node) Propose(block []byte)             { n.protocol.Propose(block) }
func (n *Node) HandleMessage(msg Message)        { n.protocol.HandleMessage(msg) }
func (n *Node) GetMetrics() Metrics              { return n.protocol.GetMetrics() }
