// simulation.go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
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
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	return &Simulation{
		config: config,
		stop:   make(chan struct{}),
	}
}

func (s *Simulation) Run() *SimulationResult {
	log.Printf("--- Starting Experiment: %s | Protocol: %s | Nodes: %d ---", s.config.Name, s.config.Protocol, s.config.NumNodes)

	// Clean up old DB files for a fresh start
	for i := 0; i < s.config.NumNodes; i++ {
		os.Remove(fmt.Sprintf("node_%s_%d.db", s.config.Protocol, i))
	}

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

	// Collect results
	return s.collectResults()
}

func (s *Simulation) clientProposer() {
	currentLeader := 0
	ticker := time.NewTicker(50 * time.Millisecond) // Propose frequently
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			block := make([]byte, s.config.BlockSize)
			rand.Read(block)

			// Find the current leader to propose to
			// In a real system, a client would have more robust leader discovery
			leaderNode := s.nodes[currentLeader]
			if leaderNode != nil && !leaderNode.IsStopped() {
				leaderNode.Propose(block)
			}
			// Simple round-robin leader assumption for client
			currentLeader = (currentLeader + 1) % s.config.NumNodes
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
		// We only count commits from one node to avoid duplication
		if node.id == 1 {
			totalCommits = metrics.Commits
			allLatencies = append(allLatencies, metrics.LatencyValues...)
		}
		totalBytesSent += metrics.BytesSent
		totalViewChanges += metrics.ViewChanges
	}
	// Average view changes across the cluster
	totalViewChanges /= s.config.NumNodes

	return &SimulationResult{
		Config:          s.config,
		TotalCommits:    totalCommits,
		TotalBytesSent:  totalBytesSent,
		TotalViewChanges: totalViewChanges,
		LatencyValues:   allLatencies,
	}
}
