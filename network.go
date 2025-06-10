// network.go
package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

type SimulatedNetwork struct {
	config   ExperimentConfig
	channels []chan Message
	wg       sync.WaitGroup
	stop     chan struct{}
}

func NewSimulatedNetwork(config ExperimentConfig) *SimulatedNetwork {
	channels := make([]chan Message, config.NumNodes)
	for i := range channels {
		channels[i] = make(chan Message, 1000) // Buffered channel
	}
	return &SimulatedNetwork{
		config:   config,
		channels: channels,
		stop:     make(chan struct{}),
	}
}

func (s *SimulatedNetwork) GetChannel(nodeID int) chan Message {
	return s.channels[nodeID]
}

func (s *SimulatedNetwork) Send(msg Message) {
	// Simulate network conditions for a specific destination
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		latency := s.config.NetworkLatency
		// Check if the destination is a straggler
		if stragglerCfg, ok := s.config.StragglerNodes[msg.To]; ok {
			latency += stragglerCfg.ExtraLatency
		}

		select {
		case <-time.After(latency):
			if rand.Float64() > s.config.PacketLossProb {
				s.channels[msg.To] <- msg
			}
		case <-s.stop:
			return
		}
	}()
}

func (s *SimulatedNetwork) Broadcast(fromNodeID int, payload interface{}) {
	for i := 0; i < s.config.NumNodes; i++ {
		if i == fromNodeID {
			continue
		}
		msg := Message{From: fromNodeID, To: i, Payload: payload}
		s.Send(msg)
	}
}

func (s *SimulatedNetwork) Stop() {
	close(s.stop)
	s.wg.Wait()
}
