// protocol_pbft.go
package main

// PBFTProtocol implements the consensus logic for standard PBFT.
type PBFTProtocol struct {
	node *Node
	// ... internal state like pre-prepare, prepare, commit logs ...
	metrics Metrics
	stopChan chan struct{}
}

func (p *PBFTProtocol) Init(node *Node) {
	p.node = node
	p.stopChan = make(chan struct{})
	p.metrics = Metrics{}
}

func (p *PBFTProtocol) Propose(block []byte) {
	// ... leader sends a single PRE-PREPARE message with the whole block ...
}

func (p *PBFTProtocol) HandleMessage(msg Message) {
	// switch msg.Payload.(type) {
	// case *PrePrepareMessage:
	//     ... validate, then broadcast PREPARE ...
	// case *PrepareMessage:
	//     ... collect 2f prepares, then broadcast COMMIT ...
	// case *CommitMessage:
	//     ... collect 2f+1 commits, then execute and commit locally ...
	// }
}

func (p *PBFTProtocol) Start() {
	defer p.node.wg.Done()
	// Main event loop
	for {
		select {
		case msg := <-p.node.net.GetChannel(p.node.id):
			p.HandleMessage(msg)
		case <-p.stopChan:
			return
		}
	}
}

func (p *PBFTProtocol) Stop() {
	close(p.stopChan)
}

func (p.PBFTProtocol) GetMetrics() Metrics {
	return p.metrics
}
