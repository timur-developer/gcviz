package ui

import (
	"context"
	"time"

	"github.com/timur-developer/gcviz/internal/domain"
)

type Model struct {
	ctx    context.Context
	cancel context.CancelFunc

	store *domain.Store

	lastUpdate time.Time
	now        time.Time
	agg        domain.Aggregates

	width  int
	height int

	helpVisible bool
	paused      bool
	layout      layoutMode
	stwTh       STWThresholds
	targetEnv   *TargetEnvInfo

	heapHistory []historyPoint
	stwP50Hist  []historyPoint
	stwP99Hist  []historyPoint
	stwMaxHist  []historyPoint

	cursor        int
	stwLabelsMode stwLabelMode

	pausedWindow   []domain.GCEvent
	pausedAgg      domain.Aggregates
	pausedHeapHist []historyPoint
	pausedSTWP50   []historyPoint
	pausedSTWP99   []historyPoint
	pausedSTWMax   []historyPoint

	snapshotWriter SnapshotWriter
	snapshotDir    string
	lastSnapshot   snapshotStatus

	manualSnapshotInFlight bool
	lastManualSnapshotAt   time.Time

	chartFocus chartFocus
	xSpan      xSpanMode
	heapYZoom  int
	stwYZoom   int
	heapYPan   int
	stwYPan    int
}

type GCEventMsg struct {
	Event domain.GCEvent
	At    time.Time
}

type SnapshotWriter interface {
	WriteSnapshot(events []domain.GCEvent, agg domain.Aggregates) (fileName string, err error)
}

type snapshotStatus struct {
	FileName string
	ErrMsg   string
}

type snapshotResultMsg snapshotStatus

type stwLabelMode int

const (
	stwLabelGCAndSTW stwLabelMode = iota
	stwLabelGCAndHeap
	stwLabelGCOnly
)

type layoutMode int

const (
	layoutSpaced layoutMode = iota
	layoutTight
)

func (l layoutMode) next() layoutMode {
	if l == layoutSpaced {
		return layoutTight
	}
	return layoutSpaced
}

func (m stwLabelMode) next() stwLabelMode {
	switch m {
	case stwLabelGCAndSTW:
		return stwLabelGCAndHeap
	case stwLabelGCAndHeap:
		return stwLabelGCOnly
	default:
		return stwLabelGCAndSTW
	}
}

type chartFocus int

const (
	chartHeap chartFocus = iota
	chartSTW
)

func (c chartFocus) next() chartFocus {
	if c == chartHeap {
		return chartSTW
	}
	return chartHeap
}

type xSpanMode int

const (
	xSpanAll xSpanMode = iota
	xSpan1h
	xSpan15m
	xSpan5m
	xSpan1m
)

func (m xSpanMode) zoomIn() xSpanMode {
	if m >= xSpan1m {
		return xSpan1m
	}
	if m == xSpanAll {
		return xSpan1h
	}
	return m + 1
}

func (m xSpanMode) zoomOut() xSpanMode {
	if m <= xSpanAll {
		return xSpanAll
	}
	return m - 1
}

func (m xSpanMode) duration() time.Duration {
	switch m {
	case xSpan1m:
		return time.Minute
	case xSpan5m:
		return 5 * time.Minute
	case xSpan15m:
		return 15 * time.Minute
	case xSpan1h:
		return time.Hour
	default:
		return 0
	}
}

type TargetEnvInfo struct {
	GOGC       string
	GOMEMLIMIT string
	GODEBUG    string
}

func NewModel(ctx context.Context, cancel context.CancelFunc, windowSize int, snapshotDir string, snapshotWriter SnapshotWriter, stwTh STWThresholds, targetEnv *TargetEnvInfo) Model {
	if stwTh.BadUs <= stwTh.WarnUs {
		stwTh = STWThresholds{WarnUs: 200, BadUs: 1000}
	}
	return Model{
		ctx:            ctx,
		cancel:         cancel,
		store:          domain.NewStore(windowSize),
		now:            time.Now(),
		snapshotDir:    snapshotDir,
		snapshotWriter: snapshotWriter,
		stwLabelsMode:  stwLabelGCAndSTW,
		layout:         layoutSpaced,
		stwTh:          stwTh,
		targetEnv:      targetEnv,
		chartFocus:     chartHeap,
		xSpan:          xSpanAll,
	}
}
