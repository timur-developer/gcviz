package main

import (
	"path/filepath"
	"time"

	"github.com/timur-developer/gcviz/internal/domain"
	"github.com/timur-developer/gcviz/internal/snapshot"
	"github.com/timur-developer/gcviz/internal/ui"
)

type snapshotWriter struct {
	dir string
}

func (w snapshotWriter) WriteSnapshot(events []domain.GCEvent, agg domain.Aggregates) (string, error) {
	path, err := snapshot.Write(w.dir, events, agg)
	if err != nil {
		return "", err
	}
	return filepath.Base(path), nil
}

func writeSnapshotOnExit(dir string, m ui.Model) error {
	events, agg := m.SnapshotState()
	if len(events) == 0 {
		return nil
	}
	if m.HasRecentManualSnapshot(time.Now(), 5*time.Second) {
		return nil
	}
	_, err := snapshot.Write(dir, events, agg)
	return err
}
