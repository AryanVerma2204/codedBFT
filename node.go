// node.go
package main

import "sync"

type ConsensusProtocol interface {
	Init(node *Node)
	Propose(block Block)
	HandleMessage(msg Message)
	Start(wg *sync.WaitGroup)
	Stop()
	GetMetrics() *Metrics
	GetCurrentView() int
}

type Node struct {
	id       int
	config   ExperimentConfig
	net      *SimulatedNetwork
	stopped  bool
	mux      sync.Mutex
	protocol ConsensusProtocol
}

func NewNode(id int, config ExperimentConfig, net *SimulatedNetwork) *Node {
	n := &Node{id: id, config: config, net: net}
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

func (n *Node) Start(wg *sync.WaitGroup)      { n.protocol.Start(wg) }
func (n *Node) Stop()                         { n.protocol.Stop() }
func (n *Node) IsStopped() bool               { n.mux.Lock(); defer n.mux.Unlock(); return n.stopped }
func (n *Node) Propose(block Block)           { n.protocol.Propose(block) }
func (n *Node) GetMetrics() *Metrics          { return n.protocol.GetMetrics() }
func (n *Node) GetCurrentView() int           { return n.protocol.GetCurrentView() }
