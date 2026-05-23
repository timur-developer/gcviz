package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/timur-developer/gcviz/internal/domain"
)

var (
	ErrAlreadyStarted = errors.New("collector already started")
	ErrNotStarted     = errors.New("collector not started")
)

type Collector struct {
	url      string
	interval time.Duration
	client   *http.Client

	mu      sync.Mutex
	started bool
	waitErr error

	doneCh chan struct{}

	eventCh chan domain.GCEvent
	errCh   chan error

	lastCyclesTotal uint64
	lastPauseHist   *float64Histogram
}

func NewCollector(url string, interval time.Duration, client *http.Client) *Collector {
	if interval <= 0 {
		interval = time.Second
	}
	if client == nil {
		client = http.DefaultClient
	}

	return &Collector{
		url:      url,
		interval: interval,
		client:   client,
		eventCh:  make(chan domain.GCEvent),
		errCh:    make(chan error),
		doneCh:   make(chan struct{}),
	}
}

func (c *Collector) Start(ctx context.Context) error {
	c.mu.Lock()
	if c.started {
		c.mu.Unlock()
		return ErrAlreadyStarted
	}
	if c.url == "" {
		c.mu.Unlock()
		return errors.New("url is empty")
	}
	c.started = true
	c.mu.Unlock()

	go c.run(ctx)
	return nil
}

func (c *Collector) Events() <-chan domain.GCEvent {
	return c.eventCh
}

func (c *Collector) Errors() <-chan error {
	return c.errCh
}

func (c *Collector) Wait() error {
	c.mu.Lock()
	started := c.started
	c.mu.Unlock()
	if !started {
		return ErrNotStarted
	}

	<-c.doneCh

	c.mu.Lock()
	defer c.mu.Unlock()
	return c.waitErr
}

func (c *Collector) Close() error {
	return c.Wait()
}

func (c *Collector) run(ctx context.Context) {
	defer close(c.eventCh)
	defer close(c.errCh)
	defer close(c.doneCh)

	if err := c.collectOnce(ctx); err != nil {
		select {
		case c.errCh <- err:
		default:
		}
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.setWaitErr(nil)
			return
		case <-ticker.C:
			if err := c.collectOnce(ctx); err != nil {
				select {
				case c.errCh <- err:
				default:
				}
			}
		}
	}
}

func (c *Collector) collectOnce(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var payload reporterPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	cyclesTotal, ok := payload.uint64Metric("/gc/cycles/total:gc-cycles")
	if !ok {
		return errors.New("missing /gc/cycles/total:gc-cycles")
	}

	uptimeSeconds := payload.UptimeSeconds

	heapLiveBytes, ok := payload.uint64Metric("/gc/heap/live:bytes")
	if !ok {
		return errors.New("missing /gc/heap/live:bytes")
	}

	heapGoalBytes, ok := payload.uint64Metric("/gc/heap/goal:bytes")
	if !ok {
		return errors.New("missing /gc/heap/goal:bytes")
	}

	pauseHist, ok := payload.histogramMetric("/sched/pauses/total/gc:seconds")
	if !ok {
		return errors.New("missing /sched/pauses/total/gc:seconds")
	}

	var stwMs float64
	if c.lastPauseHist != nil {
		stwMs = averagePauseMsDelta(*c.lastPauseHist, pauseHist)
	}

	shouldEmit := c.lastCyclesTotal == 0 || cyclesTotal != c.lastCyclesTotal

	c.lastCyclesTotal = cyclesTotal
	c.lastPauseHist = &pauseHist

	if !shouldEmit {
		return nil
	}

	ev := domain.GCEvent{
		GCNum:           int(cyclesTotal),
		TimeSinceStartS: uptimeSeconds,
		HeapLiveMB:      bytesToMB(heapLiveBytes),
		HeapGoalMB:      bytesToMB(heapGoalBytes),
		STWMarkTermMs:   stwMs,
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case c.eventCh <- ev:
		return nil
	}
}

func (c *Collector) setWaitErr(err error) {
	c.mu.Lock()
	if c.waitErr == nil {
		c.waitErr = err
	}
	c.mu.Unlock()
}

func bytesToMB(v uint64) int {
	return int(v / (1024 * 1024))
}

func averagePauseMsDelta(prev, curr float64Histogram) float64 {
	if len(prev.Buckets) == 0 || len(curr.Buckets) == 0 {
		return 0
	}
	if len(prev.Buckets) != len(curr.Buckets) || len(prev.Counts) != len(curr.Counts) {
		return 0
	}

	var sumSeconds float64
	var total uint64
	for i := 0; i < len(curr.Counts) && i+1 < len(curr.Buckets); i++ {
		if curr.Counts[i] < prev.Counts[i] {
			return 0
		}

		delta := curr.Counts[i] - prev.Counts[i]
		if delta == 0 {
			continue
		}

		lo := parseBucket(curr.Buckets[i])
		hi := parseBucket(curr.Buckets[i+1])
		mid := bucketMidpoint(lo, hi)
		sumSeconds += float64(delta) * mid
		total += delta
	}
	if total == 0 {
		return 0
	}
	return (sumSeconds / float64(total)) * 1000
}

func bucketMidpoint(lo, hi float64) float64 {
	if math.IsInf(lo, 0) || math.IsInf(hi, 0) || math.IsNaN(lo) || math.IsNaN(hi) {
		return 0
	}
	return (lo + hi) / 2
}

func parseBucket(s string) float64 {
	switch s {
	case "inf":
		return math.Inf(1)
	case "-inf":
		return math.Inf(-1)
	default:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0
		}
		return v
	}
}

type reporterPayload struct {
	UptimeSeconds float64          `json:"uptime_seconds"`
	Samples       []reporterSample `json:"samples"`
}

func (p reporterPayload) uint64Metric(name string) (uint64, bool) {
	for _, s := range p.Samples {
		if s.Name == name && s.Kind == "uint64" {
			return s.Uint64, true
		}
	}
	return 0, false
}

func (p reporterPayload) histogramMetric(name string) (float64Histogram, bool) {
	for _, s := range p.Samples {
		if s.Name == name && s.Kind == "histogram" {
			return s.Histogram, true
		}
	}
	return float64Histogram{}, false
}

type reporterSample struct {
	Name      string           `json:"name"`
	Kind      string           `json:"kind"`
	Uint64    uint64           `json:"uint64,omitempty"`
	Float64   float64          `json:"float64,omitempty"`
	Histogram float64Histogram `json:"histogram,omitempty"`
}

type float64Histogram struct {
	Counts  []uint64 `json:"counts"`
	Buckets []string `json:"buckets"`
}
