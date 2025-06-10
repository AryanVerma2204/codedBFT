// main.go
package main

import (
	"fmt"
	"time"
)

// NumRuns specifies how many times each experimental configuration is run
// to ensure statistical significance.
const NumRuns = 5

func main() {
	// Print CSV header with the new run_id column for statistical analysis
	fmt.Println("experiment_name,run_id,protocol,num_nodes,block_size,packet_loss,throughput_bps,avg_latency_ms,view_changes")

	fmt.Println("# Running Scalability Study...")
	runScalabilityStudy()
	fmt.Println("# Running Packet Loss Resilience Study...")
	runPacketLossStudy()
	fmt.Println("# Running Ablation Study...")
	runAblationStudy()
	fmt.Println("# All experiments complete.")
}

// runExperiment is a helper to execute a given configuration NumRuns times.
func runExperiment(cfg ExperimentConfig) {
	for i := 0; i < NumRuns; i++ {
		cfg.RunID = i
		sim := NewSimulation(cfg)
		res := sim.Run()
		res.PrintAsCSV()
	}
}

// runScalabilityStudy tests protocol performance as the number of nodes increases.
func runScalabilityStudy() {
	for _, n := range []int{4, 8, 16, 32} {
		runExperiment(BaseConfig("scalability", ProtoCodedBFT, n, 1*MB))
		runExperiment(BaseConfig("scalability", ProtoPBFT, n, 1*MB))
	}
}

// runPacketLossStudy tests protocol resilience to network unreliability.
func runPacketLossStudy() {
	for _, loss := range []float64{0.0, 0.01, 0.02, 0.05} {
		cfgCodedBFT := BaseConfig("packet_loss", ProtoCodedBFT, 16, 1*MB)
		cfgCodedBFT.PacketLossProb = loss
		runExperiment(cfgCodedBFT)

		cfgPBFT := BaseConfig("packet_loss", ProtoPBFT, 16, 1*MB)
		cfgPBFT.PacketLossProb = loss
		runExperiment(cfgPBFT)
	}
}

// runAblationStudy tests the specific contributions of CodedBFT's design choices.
func runAblationStudy() {
	// Compare CodedBFT with and without its key feature: speculative execution
	cfgNoSpec := BaseConfig("ablation", ProtoCodedBFTNoSpec, 16, 1*MB)
	cfgNoSpec.PacketLossProb = 0.02 // Test under moderate loss
	runExperiment(cfgNoSpec)

	cfgFullCoded := BaseConfig("ablation", ProtoCodedBFT, 16, 1*MB)
	cfgFullCoded.PacketLossProb = 0.02
	runExperiment(cfgFullCoded)
}

// BaseConfig provides a default configuration for experiments.
func BaseConfig(name string, proto ProtocolType, numNodes int, blockSize int) ExperimentConfig {
	return ExperimentConfig{
		Name:             name,
		Protocol:         proto,
		NumNodes:         numNodes,
		NumFaulty:        (numNodes - 1) / 3,
		BlockSize:        blockSize,
		PacketSize:       1400,
		NetworkLatency:   20 * time.Millisecond,
		PacketLossProb:   0,
		SimDuration:      5 * time.Second, // Shorter for quick runs, increase for more data
		ConsensusTimeout: 1 * time.Second,
	}
}
