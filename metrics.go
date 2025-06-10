// metrics.go
package main

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type Metrics struct {
	Commits       int
	BytesSent     int64
	ViewChanges   int
	LatencyValues []float64
	mux           sync.RWMutex
}

func (m *Metrics) AddCommit(latency time.Duration) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.Commits++
	m.LatencyValues = append(m.LatencyValues, float64(latency.Milliseconds()))
}

func (m *Metrics) AddBytesSent(bytes int) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.BytesSent += int64(bytes)
}

func (m *Metrics) IncViewChanges() {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.ViewChanges++
}

func (r *SimulationResult) PrintAsCSV() {
	avgLatency := 0.0
	if len(r.LatencyValues) > 0 {
		sum := 0.0
		for _, l := range r.LatencyValues {
			sum += l
		}
		avgLatency = sum / float64(len(r.LatencyValues))
	}

	throughputBps := float64(r.TotalCommits*r.Config.BlockSize*8) / r.Config.SimDuration.Seconds()
	if math.IsNaN(throughputBps) || math.IsInf(throughputBps, 0) {
		throughputBps = 0
	}

	fmt.Printf("%s,%d,%s,%d,%d,%.2f,%.2f,%.2f,%d\n",
		r.Config.Name,
		r.Config.RunID,
		r.Config.Protocol,
		r.Config.NumNodes,
		r.Config.BlockSize,
		r.Config.PacketLossProb,
		throughputBps,
		avgLatency,
		r.TotalViewChanges,
	)
}
