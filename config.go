// config.go
package main

import "time"

const (
	KB = 1024
	MB = 1024 * KB
)

// ProtocolType defines the consensus algorithm to be used.
type ProtocolType string

const (
	ProtoCodedBFT      ProtocolType = "CodedBFT"
	ProtoPBFT          ProtocolType = "PBFT"
	ProtoCodedBFTNoSpec ProtocolType = "CodedBFT-NoSpec"
)

// ExperimentConfig holds all parameters for a single simulation run.
type ExperimentConfig struct {
	Name             string
	RunID            int
	Protocol         ProtocolType
	NumNodes         int
	NumFaulty        int
	BlockSize        int
	PacketSize       int
	NetworkLatency   time.Duration
	PacketLossProb   float64
	SimDuration      time.Duration
	ConsensusTimeout time.Duration
}

// SimulationResult holds the aggregated metrics collected from a run.
type SimulationResult struct {
	Config           ExperimentConfig
	TotalCommits     int
	TotalBytesSent   int64
	TotalViewChanges int
	LatencyValues    []float64
}
