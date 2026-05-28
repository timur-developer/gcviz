package reporter

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net"
	"net/http"
	"runtime/metrics"
	"sync"
	"time"
)

const DefaultPath = "/gcscope/metrics"

type Reporter struct {
	path string

	mu     sync.Mutex
	server *http.Server
	ln     net.Listener

	startedAt time.Time
}

type Option func(*Reporter)

func WithPath(path string) Option {
	return func(r *Reporter) {
		if path != "" {
			r.path = path
		}
	}
}

func New(opts ...Option) *Reporter {
	r := &Reporter{
		path:      DefaultPath,
		startedAt: time.Now(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Reporter) Path() string {
	return r.path
}

func (r *Reporter) Handler() http.Handler {
	return http.HandlerFunc(r.handle)
}

// Start starts an HTTP server on addr (for example ":8080") and serves the metrics endpoint.
// If you already have an HTTP server, prefer Handler() and register it into your mux.
func (r *Reporter) Start(addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.server != nil {
		return errors.New("reporter already started")
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Handler:           r.serverHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	r.ln = ln
	r.server = srv

	go func() {
		_ = srv.Serve(ln)
	}()

	return nil
}

func (r *Reporter) Stop(ctx context.Context) error {
	r.mu.Lock()
	srv := r.server
	ln := r.ln
	r.server = nil
	r.ln = nil
	r.mu.Unlock()

	if ln != nil {
		_ = ln.Close()
	}
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

func (r *Reporter) handle(w http.ResponseWriter, _ *http.Request) {
	payload := r.buildPayload()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(payload); err != nil {
		http.Error(w, `{"error":"failed to encode metrics"}`, http.StatusInternalServerError)
		return
	}
}

func (r *Reporter) serverHandler() http.Handler {
	m := http.NewServeMux()
	m.Handle(r.path, r.Handler())
	return m
}

type payload struct {
	UptimeSeconds float64  `json:"uptime_seconds"`
	Samples       []sample `json:"samples"`
}

type sample struct {
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Uint64    uint64    `json:"uint64,omitempty"`
	Float64   float64   `json:"float64,omitempty"`
	Histogram histogram `json:"histogram,omitempty"`
}

type histogram struct {
	Counts  []uint64 `json:"counts"`
	Buckets []string `json:"buckets"`
}

func (r *Reporter) buildPayload() payload {
	descs := metrics.All()
	samples := make([]metrics.Sample, len(descs))
	for i := range descs {
		samples[i].Name = descs[i].Name
	}
	metrics.Read(samples)

	out := make([]sample, 0, len(samples))
	for _, s := range samples {
		switch s.Value.Kind() {
		case metrics.KindUint64:
			out = append(out, sample{
				Name:   s.Name,
				Kind:   "uint64",
				Uint64: s.Value.Uint64(),
			})
		case metrics.KindFloat64:
			out = append(out, sample{
				Name:    s.Name,
				Kind:    "float64",
				Float64: s.Value.Float64(),
			})
		case metrics.KindFloat64Histogram:
			h := s.Value.Float64Histogram()
			buckets := make([]string, 0, len(h.Buckets))
			for _, b := range h.Buckets {
				buckets = append(buckets, formatBucket(b))
			}
			out = append(out, sample{
				Name: s.Name,
				Kind: "histogram",
				Histogram: histogram{
					Counts:  append([]uint64(nil), h.Counts...),
					Buckets: buckets,
				},
			})
		default:
		}
	}

	return payload{
		UptimeSeconds: time.Since(r.startedAt).Seconds(),
		Samples:       out,
	}
}

func formatBucket(v float64) string {
	switch {
	case math.IsInf(v, 1):
		return "inf"
	case math.IsInf(v, -1):
		return "-inf"
	default:
		// Keep as a string to avoid NaN/Inf encoding issues; normal numbers are safe.
		return jsonFloat(v)
	}
}

func jsonFloat(v float64) string {
	// Use encoding/json float formatting via Marshal.
	b, err := json.Marshal(v)
	if err != nil {
		return "0"
	}
	if len(b) > 0 && b[0] == '"' {
		// Shouldn't happen for float64, but keep it safe.
		return "0"
	}
	return string(b)
}
