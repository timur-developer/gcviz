package domain

import "testing"

func TestStorePushAndSnapshotOrder(t *testing.T) {
	store := NewStore(3)

	store.Add(GCEvent{GCNum: 1})
	store.Add(GCEvent{GCNum: 2})
	store.Add(GCEvent{GCNum: 3})

	events := store.Recent()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	assertGCNums(t, events, []int{1, 2, 3})
}

func TestStoreOverflowEvictsOldest(t *testing.T) {
	store := NewStore(3)

	for i := 1; i <= 5; i++ {
		store.Add(GCEvent{GCNum: i})
	}

	events := store.Recent()
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	assertGCNums(t, events, []int{3, 4, 5})
}

func TestStoreBoundedSize(t *testing.T) {
	store := NewStore(10)

	for i := 1; i <= 1000; i++ {
		store.Add(GCEvent{GCNum: i})
	}

	if got := store.Len(); got != 10 {
		t.Fatalf("expected len 10, got %d", got)
	}

	if got := store.Capacity(); got != 10 {
		t.Fatalf("expected capacity 10, got %d", got)
	}

	events := store.Recent()
	if len(events) != 10 {
		t.Fatalf("expected 10 events, got %d", len(events))
	}

	assertGCNums(t, events, []int{991, 992, 993, 994, 995, 996, 997, 998, 999, 1000})
}

func TestStoreInvalidWindowSize(t *testing.T) {
	store := NewStore(0)

	if got := store.Capacity(); got != 1 {
		t.Fatalf("expected fallback capacity 1, got %d", got)
	}

	store.Add(GCEvent{GCNum: 1})
	store.Add(GCEvent{GCNum: 2})

	events := store.Recent()
	assertGCNums(t, events, []int{2})
}

func assertGCNums(t *testing.T, events []GCEvent, want []int) {
	t.Helper()

	if len(events) != len(want) {
		t.Fatalf("expected %d events, got %d", len(want), len(events))
	}

	for i := range want {
		if events[i].GCNum != want[i] {
			t.Fatalf("unexpected GCNum at %d: got %d, want %d", i, events[i].GCNum, want[i])
		}
	}
}
