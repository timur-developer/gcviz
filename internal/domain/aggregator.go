package domain

import (
	"math"
	"sort"
	"time"
)

type Aggregates struct {
	HasData      bool
	TargetUptime time.Duration
	Current      CurrentValues
	Window       WindowStats
}

type CurrentValues struct {
	GCCyclesTotal int
	LastSTWUs     int64
	HeapLiveMB    int
	HeapGoalMB    int
}

type WindowStats struct {
	STWP50Us int64
	STWP99Us int64
	STWMaxUs int64
}

// ComputeAggregates computes derived metrics over a sliding window of recent GC cycles.
//
// Expected usage: called on each TUI refresh. Window size is expected to be in the hundreds.
// For significantly larger windows or higher refresh rates, consider caching/incremental updates.
//
// Contract: window must be ordered by time ascending; window[len(window)-1] is the most recent GCEvent.
func ComputeAggregates(window []GCEvent) Aggregates {
	if len(window) == 0 {
		return Aggregates{}
	}

	last := window[len(window)-1]

	agg := Aggregates{
		HasData:      true,
		TargetUptime: durationFromSeconds(last.TimeSinceStartS),
		Current: CurrentValues{
			GCCyclesTotal: last.GCNum,
			LastSTWUs:     stwUs(last),
			HeapLiveMB:    last.HeapLiveMB,
			HeapGoalMB:    last.HeapGoalMB,
		},
	}

	stws := make([]int64, 0, len(window))
	for _, ev := range window {
		stws = append(stws, stwUs(ev))
	}
	sort.Slice(stws, func(i, j int) bool { return stws[i] < stws[j] })

	agg.Window.STWMaxUs = stws[len(stws)-1]
	agg.Window.STWP50Us = nearestRankPercentileUs(stws, 0.50)
	agg.Window.STWP99Us = nearestRankPercentileUs(stws, 0.99)

	return agg
}

func stwUs(ev GCEvent) int64 {
	// Source of truth is runtime output (ms); convert to us for UI-facing aggregates.
	ms := ev.STWSweepTermMs + ev.STWMarkTermMs
	return int64(math.Round(ms * 1000))
}

func durationFromSeconds(seconds float64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(math.Round(seconds * 1e9))
}

func nearestRankPercentileUs(sortedUs []int64, p float64) int64 {
	n := len(sortedUs)
	if n == 0 {
		return 0
	}

	// nearest-rank percentile on a sorted array of length n:
	// idx = ceil(p*n) - 1
	idx := int(math.Ceil(p*float64(n))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= n {
		idx = n - 1
	}
	return sortedUs[idx]
}
