// simulation.go
package main

import (
	"crypto/sha256"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Simulation struct {
	config ExperimentConfig
	net    *SimulatedNetwork
	nodes  []*Node
	wg     sync.WaitGroup
	stop   chan struct{}
}

func NewSimulation(config ExperimentConfig) *Simulation {
	return &Simulation{
		config: config,
		stop:   make(chan struct{}),
	}
}

func (s *Simulation) Run() *SimulationResult {
	log.Printf("--- Running: %s | RunID: %d | Protocol: %s | Nodes: %d | Loss: %.2f%% ---", s.config.Name, s.config.RunID, s.config.Protocol, s.config.NumNodes, s.config.PacketLossProb*100)

	s.net = NewSimulatedNetwork(s.config)
	s.nodes = make([]*Node, s.config.NumNodes)
	for i := 0; i < s.config.NumNodes; i++ {
		s.nodes[i] = NewNode(i, s.config, s.net)
		s.wg.Add(1)
		go s.nodes[i].Start(&s.wg)
	}

	// Client goroutine to propose blocks
	go s.clientProposer()

	// Run for the configured duration
	time.Sleep(s.config.SimDuration)
	close(s.stop) // Signal all goroutines to stop

	return s.collectResults()
}

func (s *Simulation) clientProposer() {
	// A simple client that continuously proposes new blocks to the current leader
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// In each round, a different node acts as leader to distribute load
			leaderID := s.nodes[0].GetCurrentView() % s.config.NumNodes
			leaderNode := s.nodes[leaderID]

			if leaderNode != nil && !leaderNode.IsStopped() {
				blockData := make([]byte, s.config.BlockSize)
				rand.Read(blockData)
				block := Block{
					ID:        sha256.Sum256(blockData),
					Proposer:  leaderID,
					Timestamp: time.Now(),
					Data:      blockData,
				}
				leaderNode.Propose(block)
			}
		case <-s.stop:
			return
		}
	}
}

func (s *Simulation) collectResults() *SimulationResult {
	for _, node := range s.nodes {
		node.Stop()
	}
	s.wg.Wait() // Wait for all nodes to finish cleanly

	var totalCommits int
	var totalBytesSent int64
	var totalViewChanges int
	var allLatencies []float64

	for _, node := range s.nodes {
		metrics := node.GetMetrics()
		if node.id == 1 { // Collect detailed stats from one arbitrary node
			totalCommits = metrics.Commits
			allLatencies = append(allLatencies, metrics.LatencyValues...)
		}
		totalBytesSent += metrics.BytesSent
		totalViewChanges += metrics.ViewChanges
	}
	totalViewChanges /= s.config.NumNodes // Average view changes across the cluster

	return &SimulationResult{
		Config:           s.config,
		TotalCommits:     totalCommits,
		TotalBytesSent:   totalBytesSent,
		TotalViewChanges: totalViewChanges,
		LatencyValues:    allLatencies,
	}
}
