// main.go
package main

import (
	"fmt"
	"time"
)

// main orchestrates the entire experimental evaluation.
func main() {
	// Print CSV header
	fmt.Println("experiment_name,protocol,num_nodes,block_size,packet_loss,stragglers,throughput_bps,avg_latency_ms,view_changes")

	runScalabilityStudy()
	runPacketLossStudy()
	runAblationStudy()
}

// runScalabilityStudy tests how protocols perform as the number of nodes increases.
func runScalabilityStudy() {
	for _, n := range []int{4, 8, 16, 32} {
		// CodedBFT
		cfgCoded := BaseConfig("scalability", ProtoCodedBFT, n, 1*MB)
		simCoded := NewSimulation(cfgCoded)
		resCoded := simCoded.Run()
		resCoded.PrintAsCSV()

		// PBFT
		cfgPbft := BaseConfig("scalability", ProtoPBFT, n, 1*MB)
		simPbft := NewSimulation(cfgPbft)
		resPbft := simPbft.Run()
		resPbft.PrintAsCSV()
	}
}

// runPacketLossStudy tests protocol resilience to network unreliability.
func runPacketLossStudy() {
	for _, loss := range []float64{0.0, 0.01, 0.02, 0.05, 0.10} {
		// CodedBFT
		cfgCoded := BaseConfig("packet_loss", ProtoCodedBFT, 16, 1*MB)
		cfgCoded.PacketLossProb = loss
		simCoded := NewSimulation(cfgCoded)
		resCoded := simCoded.Run()
		resCoded.PrintAsCSV()

		// PBFT
		cfgPbft := BaseConfig("packet_loss", ProtoPBFT, 16, 1*MB)
		cfgPbft.PacketLossProb = loss
		simPbft := NewSimulation(cfgPbft)
		resPbft := simPbft.Run()
		resPbft.PrintAsCSV()
	}
}

// runAblationStudy tests the specific contributions of CodedBFT's design.
func runAblationStudy() {
	// Test 1: CodedBFT without speculative execution
	cfgNoSpec := BaseConfig("ablation", ProtoCodedBFTNoSpec, 16, 1*MB)
	cfgNoSpec.PacketLossProb = 0.02 // Test under moderate loss
	simNoSpec := NewSimulation(cfgNoSpec)
	resNoSpec := simNoSpec.Run()
	resNoSpec.PrintAsCSV()

	// For comparison, the full CodedBFT under the same conditions
	cfgFullCoded := BaseConfig("ablation", ProtoCodedBFT, 16, 1*MB)
	cfgFullCoded.PacketLossProb = 0.02
	simFullCoded := NewSimulation(cfgFullCoded)
	resFullCoded := simFullCoded.Run()
	resFullCoded.PrintAsCSV()
}

// BaseConfig provides a default configuration for experiments.
func BaseConfig(name string, proto ProtocolType, numNodes int, blockSize int) ExperimentConfig {
	return ExperimentConfig{
		Name:            name,
		Protocol:        proto,
		NumNodes:        numNodes,
		NumFaulty:       (numNodes - 1) / 3,
		BlockSize:       blockSize,
		PacketSize:      1400,
		NetworkLatency:  20 * time.Millisecond,
		PacketLossProb:  0,
		SimDuration:     10 * time.Second,
		StragglerNodes:  make(map[int]StragglerConfig),
		ConsensusTimeout: 2 * time.Second,
	}
}
