package ui

import (
	"math"
	"time"
)

func appendLimited(s []historyPoint, v historyPoint, limit int) []historyPoint {
	s = append(s, v)
	if limit <= 0 {
		return s
	}
	if len(s) <= limit {
		return s
	}
	return s[len(s)-limit:]
}

func (m *Model) clampZoomState() {
	var heapHist []historyPoint
	var p50Hist []historyPoint
	var p99Hist []historyPoint
	var maxHist []historyPoint
	if m.paused {
		heapHist = m.pausedHeapHist
		p50Hist = m.pausedSTWP50
		p99Hist = m.pausedSTWP99
		maxHist = m.pausedSTWMax
	} else {
		heapHist = m.heapHistory
		p50Hist = m.stwP50Hist
		p99Hist = m.stwP99Hist
		maxHist = m.stwMaxHist
	}

	xSpan := m.xSpan.duration()
	m.heapYPan = clampPanSteps(heapHist, xSpan, m.heapYZoom, m.heapYPan, 1.0)

	all := make([]historyPoint, 0, len(p50Hist)+len(p99Hist)+len(maxHist))
	all = append(all, p50Hist...)
	all = append(all, p99Hist...)
	all = append(all, maxHist...)
	m.stwYPan = clampPanSteps(all, xSpan, m.stwYZoom, m.stwYPan, 10.0)
}

func clampPanSteps(points []historyPoint, xSpan time.Duration, zoomSteps int, panSteps int, minRange float64) int {
	if len(points) == 0 {
		return 0
	}

	vis := points
	if xSpan > 0 {
		end := points[len(points)-1].At
		start := end.Add(-xSpan)
		idx := 0
		for i, p := range points {
			if !p.At.Before(start) {
				idx = i
				break
			}
		}
		if idx < len(points) {
			vis = points[idx:]
		}
	}
	if len(vis) == 0 {
		return 0
	}

	minY := vis[0].Value
	maxY := vis[0].Value
	for i := 1; i < len(vis); i++ {
		v := vis[i].Value
		if v < minY {
			minY = v
		}
		if v > maxY {
			maxY = v
		}
	}
	if minY < 0 {
		minY = 0
	}
	if maxY < minY {
		maxY = minY
	}

	rng := maxY - minY
	if rng < minRange {
		rng = minRange
	}

	pad := rng * 0.05
	if pad < minRange*0.05 {
		pad = minRange * 0.05
	}
	boundMin := minY - pad
	if boundMin < 0 {
		boundMin = 0
	}
	boundMax := maxY + pad
	if boundMax-boundMin < minRange {
		boundMax = boundMin + minRange
	}

	zoom := zoomSteps
	if zoom < 0 {
		zoom = 0
	}
	if zoom > 0 {
		scale := math.Pow(0.75, float64(zoom))
		if scale < 0.05 {
			scale = 0.05
		}
		rng = rng * scale
		if rng < minRange {
			rng = minRange
		}
	}

	center := (minY + maxY) / 2
	if len(vis) > 0 {
		center = vis[len(vis)-1].Value
	}
	minV0 := center - rng/2
	maxV0 := center + rng/2

	step := rng * 0.2
	if step < minRange*0.2 {
		step = minRange * 0.2
	}
	if step <= 0 {
		return 0
	}

	// We want [minV0 + s, maxV0 + s] to be inside [boundMin, boundMax].
	minShift := boundMin - minV0
	maxShift := boundMax - maxV0

	minSteps := int(math.Ceil(minShift / step))
	maxSteps := int(math.Floor(maxShift / step))
	if minSteps > maxSteps {
		return 0
	}
	if panSteps < minSteps {
		return minSteps
	}
	if panSteps > maxSteps {
		return maxSteps
	}
	return panSteps
}

// stackPanels builds vertical rows that always fit into content.H (including gaps).
