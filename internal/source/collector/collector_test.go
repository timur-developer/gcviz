package collector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCollector_EmitsEvent(t *testing.T) {
	payload := reporterPayload{
		UptimeSeconds: 12.5,
		Samples: []reporterSample{
			{Name: "/gc/cycles/total:gc-cycles", Kind: "uint64", Uint64: 7},
			{Name: "/gc/heap/live:bytes", Kind: "uint64", Uint64: 64 * 1024 * 1024},
			{Name: "/gc/heap/goal:bytes", Kind: "uint64", Uint64: 96 * 1024 * 1024},
			{
				Name: "/sched/pauses/total/gc:seconds",
				Kind: "histogram",
				Histogram: float64Histogram{
					Buckets: []string{"0", "0.001", "0.01"},
					Counts:  []uint64{0, 0},
				},
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := NewCollector(srv.URL, 10*time.Millisecond, srv.Client())
	if err := c.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}

	select {
	case ev := <-c.Events():
		if ev.GCNum != 7 {
			t.Fatalf("expected GCNum 7, got %d", ev.GCNum)
		}
		if ev.TimeSinceStartS != 12.5 {
			t.Fatalf("expected uptime 12.5, got %v", ev.TimeSinceStartS)
		}
		if ev.HeapLiveMB != 64 {
			t.Fatalf("expected heap live 64MB, got %d", ev.HeapLiveMB)
		}
		if ev.HeapGoalMB != 96 {
			t.Fatalf("expected heap goal 96MB, got %d", ev.HeapGoalMB)
		}
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for event")
	}

	cancel()
	_ = c.Wait()
}

func TestCollector_TemporaryHTTPErrorDoesNotStop(t *testing.T) {
	var calls int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		payload := reporterPayload{
			UptimeSeconds: 1,
			Samples: []reporterSample{
				{Name: "/gc/cycles/total:gc-cycles", Kind: "uint64", Uint64: 1},
				{Name: "/gc/heap/live:bytes", Kind: "uint64", Uint64: 1},
				{Name: "/gc/heap/goal:bytes", Kind: "uint64", Uint64: 1},
				{
					Name: "/sched/pauses/total/gc:seconds",
					Kind: "histogram",
					Histogram: float64Histogram{
						Buckets: []string{"0", "0.001"},
						Counts:  []uint64{0},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := NewCollector(srv.URL, 10*time.Millisecond, srv.Client())
	if err := c.Start(ctx); err != nil {
		t.Fatalf("start: %v", err)
	}

	select {
	case <-c.Errors():
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for error")
	}

	select {
	case <-c.Events():
	case <-time.After(time.Second):
		t.Fatalf("timeout waiting for event after error")
	}

	cancel()
	_ = c.Wait()
}
