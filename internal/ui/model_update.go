package ui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/timur-developer/gcviz/internal/domain"
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(waitContextDone(m.ctx), tick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyRunes && !msg.Paste && !msg.Alt {
			var cmds []tea.Cmd
			for _, r := range msg.Runes {
				switch r {
				case 'q':
					m.cancel()
					return m, tea.Quit
				case ' ':
					m.togglePause()
					m.clampZoomState()
				case '?', 'h':
					m.helpVisible = !m.helpVisible
				case 's':
					m.manualSnapshotInFlight = true
					cmds = append(cmds, takeSnapshotCmd(m.store.Recent(), m.agg, m.snapshotWriter))
				case 'l':
					m.stwLabelsMode = m.stwLabelsMode.next()
				case 'g':
					m.layout = m.layout.next()
				case 'z':
					m.chartFocus = m.chartFocus.next()
				case '+', '=':
					const maxZoomSteps = 8
					switch m.chartFocus {
					case chartHeap:
						if m.heapYZoom < maxZoomSteps {
							m.heapYZoom++
						}
					case chartSTW:
						if m.stwYZoom < maxZoomSteps {
							m.stwYZoom++
						}
					}
					m.clampZoomState()
				case '-':
					switch m.chartFocus {
					case chartHeap:
						if m.heapYZoom > 0 {
							m.heapYZoom--
						}
					case chartSTW:
						if m.stwYZoom > 0 {
							m.stwYZoom--
						}
					}
					m.clampZoomState()
				case '0':
					switch m.chartFocus {
					case chartHeap:
						m.heapYZoom = 0
						m.heapYPan = 0
					case chartSTW:
						m.stwYZoom = 0
						m.stwYPan = 0
					}
					m.clampZoomState()
				case '[':
					m.xSpan = m.xSpan.zoomIn()
					m.clampZoomState()
				case ']':
					m.xSpan = m.xSpan.zoomOut()
					m.clampZoomState()
				case 'r':
					m.chartFocus = chartHeap
					m.xSpan = xSpanAll
					m.heapYZoom = 0
					m.stwYZoom = 0
					m.heapYPan = 0
					m.stwYPan = 0
				}
			}
			return m, tea.Batch(cmds...)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.cancel()
			return m, tea.Quit
		case "?", "h", "f1":
			m.helpVisible = !m.helpVisible
			return m, nil
		case "s":
			m.manualSnapshotInFlight = true
			return m, takeSnapshotCmd(m.store.Recent(), m.agg, m.snapshotWriter)
		case "l":
			m.stwLabelsMode = m.stwLabelsMode.next()
			return m, nil
		case "g":
			m.layout = m.layout.next()
			return m, nil
		case "z":
			m.chartFocus = m.chartFocus.next()
			return m, nil
		case "+", "=":
			const maxZoomSteps = 8
			switch m.chartFocus {
			case chartHeap:
				if m.heapYZoom < maxZoomSteps {
					m.heapYZoom++
				}
			case chartSTW:
				if m.stwYZoom < maxZoomSteps {
					m.stwYZoom++
				}
			}
			m.clampZoomState()
			return m, nil
		case "-":
			switch m.chartFocus {
			case chartHeap:
				if m.heapYZoom > 0 {
					m.heapYZoom--
				}
			case chartSTW:
				if m.stwYZoom > 0 {
					m.stwYZoom--
				}
			}
			m.clampZoomState()
			return m, nil
		case "0":
			switch m.chartFocus {
			case chartHeap:
				m.heapYZoom = 0
				m.heapYPan = 0
			case chartSTW:
				m.stwYZoom = 0
				m.stwYPan = 0
			}
			m.clampZoomState()
			return m, nil
		case "shift+up", "ctrl+up", "ctrl+shift+up":
			switch m.chartFocus {
			case chartHeap:
				m.heapYPan++
			case chartSTW:
				m.stwYPan++
			}
			m.clampZoomState()
			return m, nil
		case "shift+down", "ctrl+down", "ctrl+shift+down":
			switch m.chartFocus {
			case chartHeap:
				m.heapYPan--
			case chartSTW:
				m.stwYPan--
			}
			m.clampZoomState()
			return m, nil
		case "r":
			m.chartFocus = chartHeap
			m.xSpan = xSpanAll
			m.heapYZoom = 0
			m.stwYZoom = 0
			m.heapYPan = 0
			m.stwYPan = 0
			return m, nil
		case "[":
			m.xSpan = m.xSpan.zoomIn()
			m.clampZoomState()
			return m, nil
		case "]":
			m.xSpan = m.xSpan.zoomOut()
			m.clampZoomState()
			return m, nil
		case " ":
			m.togglePause()
			m.clampZoomState()
			return m, nil
		case "left":
			m.moveCursor(-1)
			return m, nil
		case "right":
			m.moveCursor(1)
			return m, nil
		case "home":
			m.setCursor(0)
			return m, nil
		case "end":
			m.setCursor(m.currentWindowLen() - 1)
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case GCEventMsg:
		m.lastUpdate = msg.At
		m.now = msg.At
		m.store.Add(msg.Event)
		m.agg = domain.ComputeAggregates(m.store.Recent())
		m.pushHistory(msg.At)
		m.clampZoomState()
		if !m.paused {
			m.cursor = m.currentWindowLen() - 1
		}
		return m, nil
	case tickMsg:
		m.now = msg.At
		return m, tick()
	case contextDoneMsg:
		return m, tea.Quit
	case snapshotResultMsg:
		m.lastSnapshot = snapshotStatus(msg)
		if m.manualSnapshotInFlight {
			m.manualSnapshotInFlight = false
			if m.lastSnapshot.FileName != "" && m.lastSnapshot.ErrMsg == "" {
				m.lastManualSnapshotAt = m.now
			}
		}
		return m, nil
	}

	return m, nil
}

func (m Model) HasRecentManualSnapshot(now time.Time, within time.Duration) bool {
	if m.lastManualSnapshotAt.IsZero() {
		return false
	}
	if within <= 0 {
		return true
	}
	return now.Sub(m.lastManualSnapshotAt) <= within
}

func (m Model) SnapshotState() ([]domain.GCEvent, domain.Aggregates) {
	return m.store.Recent(), m.agg
}

type contextDoneMsg struct{}

func waitContextDone(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		<-ctx.Done()
		return contextDoneMsg{}
	}
}

type tickMsg struct{ At time.Time }

func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{At: t}
	})
}

func takeSnapshotCmd(events []domain.GCEvent, agg domain.Aggregates, w SnapshotWriter) tea.Cmd {
	if w == nil {
		return nil
	}
	if len(events) == 0 {
		return nil
	}

	return func() tea.Msg {
		name, err := w.WriteSnapshot(events, agg)
		if err != nil {
			return snapshotResultMsg(snapshotStatus{ErrMsg: err.Error()})
		}
		return snapshotResultMsg(snapshotStatus{FileName: name})
	}
}

func (m *Model) pushHistory(at time.Time) {
	if !m.agg.HasData {
		return
	}

	const limit = 180

	m.heapHistory = appendLimited(m.heapHistory, historyPoint{At: at, Value: float64(m.agg.Current.HeapLiveMB)}, limit)
	m.stwP50Hist = appendLimited(m.stwP50Hist, historyPoint{At: at, Value: float64(m.agg.Window.STWP50Us)}, limit)
	m.stwP99Hist = appendLimited(m.stwP99Hist, historyPoint{At: at, Value: float64(m.agg.Window.STWP99Us)}, limit)
	m.stwMaxHist = appendLimited(m.stwMaxHist, historyPoint{At: at, Value: float64(m.agg.Window.STWMaxUs)}, limit)
}

func (m *Model) togglePause() {
	if m.paused {
		m.paused = false
		m.cursor = m.currentWindowLen() - 1
		m.pausedWindow = nil
		m.pausedAgg = domain.Aggregates{}
		m.pausedHeapHist = nil
		m.pausedSTWP50 = nil
		m.pausedSTWP99 = nil
		m.pausedSTWMax = nil
		return
	}

	m.paused = true
	m.pausedWindow = m.store.Recent()
	m.pausedAgg = domain.ComputeAggregates(m.pausedWindow)
	m.pausedHeapHist = append([]historyPoint(nil), m.heapHistory...)
	m.pausedSTWP50 = append([]historyPoint(nil), m.stwP50Hist...)
	m.pausedSTWP99 = append([]historyPoint(nil), m.stwP99Hist...)
	m.pausedSTWMax = append([]historyPoint(nil), m.stwMaxHist...)
	m.cursor = len(m.pausedWindow) - 1
}

func (m *Model) currentWindowLen() int {
	if m.paused {
		return len(m.pausedWindow)
	}
	return m.store.Len()
}

func (m *Model) moveCursor(delta int) {
	if !m.paused {
		return
	}
	m.setCursor(m.cursor + delta)
}

func (m *Model) setCursor(v int) {
	if !m.paused {
		return
	}
	max := len(m.pausedWindow) - 1
	if max < 0 {
		m.cursor = 0
		return
	}
	if v < 0 {
		v = 0
	}
	if v > max {
		v = max
	}
	m.cursor = v
}
