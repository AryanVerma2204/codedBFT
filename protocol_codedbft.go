// protocol_codedbft.go
package main

// CodedBFTProtocol implements the consensus logic for CodedBFT.
type CodedBFTProtocol struct {
	node *Node // Back-reference to the parent node
	// ... internal state like currentView, logs, decoder, etc. ...
	speculativeExecutionEnabled bool
	metrics Metrics
	stopChan chan struct{}
}

func (p *CodedBFTProtocol) Init(node *Node) {
	p.node = node
	p.stopChan = make(chan struct{})
	p.metrics = Metrics{}
}

func (p *CodedBFTProtocol) Propose(block []byte) {
	// ... leader encodes block into packets ...
	// ... broadcasts PROPOSAL messages ...
}

func (p *CodedBFTProtocol) HandleMessage(msg Message) {
	// switch msg.Payload.(type) {
	// case *ProposalPacket:
	//     ... add to decoder ...
	//     if decoded {
	//         if p.speculativeExecutionEnabled {
	//             // vote immediately
	//         } else {
	//             // wait for a GO signal (not implemented in this sketch)
	//         }
	//     }
	// case *VoteMessage:
	//     ... collect votes, form quorum, commit ...
	// }
}

func (p *CodedBFTProtocol) Start() {
	defer p.node.wg.Done()
	// Main event loop
	for {
		select {
		case msg := <-p.node.net.GetChannel(p.node.id):
			p.HandleMessage(msg)
		case <-p.stopChan:
			// clean up resources if any
			return
		}
	}
}

func (p *CodedBFTProtocol) Stop() {
	close(p.stopChan)
}

func (p *CodedBFTProtocol) GetMetrics() Metrics {
	return p.metrics
}
