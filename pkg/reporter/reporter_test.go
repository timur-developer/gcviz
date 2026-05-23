package reporter

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReporterHandler_EmitsRuntimeMetrics(t *testing.T) {
	rep := New()

	mux := http.NewServeMux()
	mux.Handle(rep.Path(), rep.Handler())
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + rep.Path())
	if err != nil {
		t.Fatalf("http get: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		t.Fatalf("expected Content-Type, got empty")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("expected non-empty body, got empty")
	}

	var got struct {
		UptimeSeconds float64 `json:"uptime_seconds"`
		Samples       []struct {
			Name string `json:"name"`
			Kind string `json:"kind"`
		} `json:"samples"`
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode: %v", err)
	}

	need := map[string]bool{
		"/gc/cycles/total:gc-cycles":     false,
		"/gc/heap/live:bytes":            false,
		"/gc/heap/goal:bytes":            false,
		"/sched/pauses/total/gc:seconds": false,
	}

	for _, s := range got.Samples {
		if _, ok := need[s.Name]; ok {
			need[s.Name] = true
		}
	}

	for name, ok := range need {
		if !ok {
			t.Fatalf("missing sample %s", name)
		}
	}

	if got.UptimeSeconds <= 0 {
		t.Fatalf("expected uptime_seconds > 0, got %v", got.UptimeSeconds)
	}
}
