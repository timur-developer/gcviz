package ui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/timur-developer/gcscope/internal/domain"
)

func (m Model) View() string {
	if m.helpVisible {
		return renderHelp(m.width, m.height)
	}

	if m.layout == layoutTight {
		return m.viewTight()
	}
	return m.viewSpaced()
}

func (m Model) viewSpaced() string {
	const (
		paddingX = 0
		paddingY = 0
		gapX     = 1
		gapY     = 1
	)

	w := m.width
	h := m.height
	if w <= 0 {
		w = 120
	}
	if h <= 0 {
		h = 40
	}

	screen := Rect{W: w, H: h}
	content := Rect{
		X: paddingX,
		Y: paddingY,
		W: w - paddingX*2,
		H: h - paddingY*2,
	}
	if content.W < 0 {
		content.W = 0
	}
	if content.H < 0 {
		content.H = 0
	}

	// Reserve one line for the footer so panel borders don't get truncated.
	footerH := 1
	contentPanels := content
	if contentPanels.H > footerH {
		contentPanels.H -= footerH
	} else {
		contentPanels.H = 0
	}

	// Fallback for narrow terminals: stack panels vertically.
	if content.W < 90 {
		rows := stackPanels(contentPanels, gapY, 4, []int{7, 7, 10, 8, 10, 12})
		if len(rows) == 0 {
			return lipgloss.NewStyle().Padding(paddingY, paddingX).Render("(terminal too small)")
		}

		window, agg, heapHist, p50Hist, p99Hist, maxHist := m.displayData()

		current := renderCurrentValues(agg, frameBoxed, m.stwTh, rows[0].W, rows[0].H)
		parts := []string{current}

		if len(rows) > 1 {
			info := renderInformation(window, agg, m.now, m.lastUpdate, m.snapshotDir, m.lastSnapshot, frameBoxed, m.stwTh, m.targetEnv, rows[1].W, rows[1].H)
			parts = append(parts, info)
		}

		var visWindow []domain.GCEvent
		visCursor := 0
		if len(rows) > 2 {
			visWindow, visCursor = m.barViewport(window, frameBoxed, rows[2].W, rows[2].H)
			bar := renderSTWBarChart(visWindow, visCursor, frameBoxed, m.stwLabelsMode, m.stwTh, 0, rows[2].H, rows[2].W)
			parts = append(parts, bar)
		}
		if len(rows) > 3 {
			details := renderCycleDetails(visWindow, visCursor, frameBoxed, m.stwTh, rows[3].W, rows[3].H)
			parts = append(parts, details)
		}
		if len(rows) > 4 {
			heap := renderHeapLiveHistory(heapHist, frameBoxed, chartView{XSpan: m.xSpan.duration(), YZoomSteps: m.heapYZoom, YPanSteps: m.heapYPan, Focused: m.chartFocus == chartHeap}, rows[4].W, rows[4].H)
			parts = append(parts, heap)
		}
		if len(rows) > 5 {
			stw := renderSTWPercentilesHistory(p50Hist, p99Hist, maxHist, frameBoxed, chartView{XSpan: m.xSpan.duration(), YZoomSteps: m.stwYZoom, YPanSteps: m.stwYPan, Focused: m.chartFocus == chartSTW}, rows[5].W, rows[5].H)
			parts = append(parts, stw)
		}

		app := strings.Join(parts, strings.Repeat("\n", gapY))
		app = m.withFooter(app, content.W)
		app = fitViewport(app, content.W, content.H)
		_ = screen
		return lipgloss.NewStyle().Padding(paddingY, paddingX).Render(app)
	}

	// Height-based layout: scale rows to fit available height.
	// Priorities: row1 (current+info) > row2 (stw+heap) > row3 (stw p50/p99).
	rows := stackPanels(contentPanels, gapY, 6, []int{8, 12, 10})
	if len(rows) == 0 {
		return lipgloss.NewStyle().Padding(paddingY, paddingX).Render("(terminal too small)")
	}

	row1AvailW := rows[0].W - gapX
	if row1AvailW < 0 {
		row1AvailW = 0
	}
	row1Cols := Cols(Rect{W: row1AvailW, H: rows[0].H}, 0.50, 0.50)

	window, agg, heapHist, p50Hist, p99Hist, maxHist := m.displayData()

	current := renderCurrentValues(agg, frameBoxed, m.stwTh, row1Cols[0].W, row1Cols[0].H)
	info := renderInformation(window, agg, m.now, m.lastUpdate, m.snapshotDir, m.lastSnapshot, frameBoxed, m.stwTh, m.targetEnv, row1Cols[1].W, row1Cols[1].H)

	parts := []string{
		lipgloss.JoinHorizontal(lipgloss.Top, current, strings.Repeat(" ", gapX), info),
	}

	if len(rows) >= 2 {
		row2AvailW := rows[1].W - gapX*2
		if row2AvailW < 0 {
			row2AvailW = 0
		}
		// Give the bar chart and details more room; heap chart is still readable at ~40%.
		row2Cols := Cols(Rect{W: row2AvailW, H: rows[1].H}, 0.36, 0.24, 0.40)

		visWindow, visCursor := m.barViewport(window, frameBoxed, row2Cols[0].W, row2Cols[0].H)
		bar := renderSTWBarChart(visWindow, visCursor, frameBoxed, m.stwLabelsMode, m.stwTh, 0, row2Cols[0].H, row2Cols[0].W)
		details := renderCycleDetails(visWindow, visCursor, frameBoxed, m.stwTh, row2Cols[1].W, row2Cols[1].H)
		heap := renderHeapLiveHistory(heapHist, frameBoxed, chartView{XSpan: m.xSpan.duration(), YZoomSteps: m.heapYZoom, YPanSteps: m.heapYPan, Focused: m.chartFocus == chartHeap}, row2Cols[2].W, row2Cols[2].H)

		parts = append(parts,
			lipgloss.JoinHorizontal(lipgloss.Top, bar, strings.Repeat(" ", gapX), details, strings.Repeat(" ", gapX), heap),
		)
	}

	if len(rows) >= 3 {
		stw := renderSTWPercentilesHistory(p50Hist, p99Hist, maxHist, frameBoxed, chartView{XSpan: m.xSpan.duration(), YZoomSteps: m.stwYZoom, YPanSteps: m.stwYPan, Focused: m.chartFocus == chartSTW}, rows[2].W, rows[2].H)
		parts = append(parts, stw)
	}

	app := strings.Join(parts, strings.Repeat("\n", gapY))
	app = m.withFooter(app, content.W)
	app = fitViewport(app, content.W, content.H)

	_ = screen
	return lipgloss.NewStyle().Padding(paddingY, paddingX).Render(app)
}

func (m Model) viewTight() string {
	const (
		paddingX = 0
		paddingY = 0
		gridSepX = 1
		gridSepY = 1
	)

	w := m.width
	h := m.height
	if w <= 0 {
		w = 120
	}
	if h <= 0 {
		h = 40
	}

	screen := Rect{W: w, H: h}
	content := Rect{
		X: paddingX,
		Y: paddingY,
		W: w - paddingX*2,
		H: h - paddingY*2,
	}
	if content.W < 0 {
		content.W = 0
	}
	if content.H < 0 {
		content.H = 0
	}

	// Reserve one line for the footer so panel borders don't get truncated.
	footerH := 1
	contentPanels := content
	if contentPanels.H > footerH {
		contentPanels.H -= footerH
	} else {
		contentPanels.H = 0
	}

	// Fallback for narrow terminals: stack panels vertically.
	if content.W < 90 {
		gridW := contentPanels.W
		gridH := contentPanels.H
		cellW := gridW - 2
		if cellW < 1 || gridH < 3 {
			return lipgloss.NewStyle().Padding(paddingY, paddingX).Render("(terminal too small)")
		}

		rows := stackPanels(Rect{W: cellW, H: gridH - 2}, gridSepY, 3, []int{5, 5, 8, 6, 8, 10})
		if len(rows) == 0 {
			return lipgloss.NewStyle().Padding(paddingY, paddingX).Render("(terminal too small)")
		}

		window, agg, heapHist, p50Hist, p99Hist, maxHist := m.displayData()

		current := renderCurrentValues(agg, framePanel, m.stwTh, cellW, rows[0].H)
		gridRows := []gridRow{
			{cellWidths: []int{cellW}, height: rows[0].H, cells: []string{current}},
		}

		if len(rows) > 1 {
			info := renderInformation(window, agg, m.now, m.lastUpdate, m.snapshotDir, m.lastSnapshot, framePanel, m.stwTh, m.targetEnv, cellW, rows[1].H)
			gridRows = append(gridRows, gridRow{cellWidths: []int{cellW}, height: rows[1].H, cells: []string{info}})
		}

		var visWindow []domain.GCEvent
		visCursor := 0
		if len(rows) > 2 {
			visWindow, visCursor = m.barViewport(window, framePanel, cellW, rows[2].H)
			bar := renderSTWBarChart(visWindow, visCursor, framePanel, m.stwLabelsMode, m.stwTh, 0, rows[2].H, cellW)
			gridRows = append(gridRows, gridRow{cellWidths: []int{cellW}, height: rows[2].H, cells: []string{bar}})
		}
		if len(rows) > 3 {
			details := renderCycleDetails(visWindow, visCursor, framePanel, m.stwTh, cellW, rows[3].H)
			gridRows = append(gridRows, gridRow{cellWidths: []int{cellW}, height: rows[3].H, cells: []string{details}})
		}
		if len(rows) > 4 {
			heap := renderHeapLiveHistory(heapHist, framePanel, chartView{XSpan: m.xSpan.duration(), YZoomSteps: m.heapYZoom, YPanSteps: m.heapYPan, Focused: m.chartFocus == chartHeap}, cellW, rows[4].H)
			gridRows = append(gridRows, gridRow{cellWidths: []int{cellW}, height: rows[4].H, cells: []string{heap}})
		}
		if len(rows) > 5 {
			stw := renderSTWPercentilesHistory(p50Hist, p99Hist, maxHist, framePanel, chartView{XSpan: m.xSpan.duration(), YZoomSteps: m.stwYZoom, YPanSteps: m.stwYPan, Focused: m.chartFocus == chartSTW}, cellW, rows[5].H)
			gridRows = append(gridRows, gridRow{cellWidths: []int{cellW}, height: rows[5].H, cells: []string{stw}})
		}

		app := renderSharedBorderGrid(gridRows, gridW)
		app = m.withFooter(app, content.W)
		app = fitViewport(app, content.W, content.H)
		_ = screen
		return lipgloss.NewStyle().Padding(paddingY, paddingX).Render(app)
	}

	// Height-based layout: scale rows to fit available height.
	// Priorities: row1 (current+info) > row2 (stw+heap) > row3 (stw p50/p99).
	gridW := contentPanels.W
	gridH := contentPanels.H
	if gridW < 10 || gridH < 6 {
		return lipgloss.NewStyle().Padding(paddingY, paddingX).Render("(terminal too small)")
	}
	rows := stackPanels(Rect{W: gridW, H: gridH - 2}, gridSepY, 4, []int{6, 10, 8})
	if len(rows) == 0 {
		return lipgloss.NewStyle().Padding(paddingY, paddingX).Render("(terminal too small)")
	}

	row1AvailW := rows[0].W - 2 - gridSepX
	if row1AvailW < 0 {
		row1AvailW = 0
	}
	row1Cols := Cols(Rect{W: row1AvailW, H: rows[0].H}, 0.50, 0.50)

	window, agg, heapHist, p50Hist, p99Hist, maxHist := m.displayData()

	current := renderCurrentValues(agg, framePanel, m.stwTh, row1Cols[0].W, row1Cols[0].H)
	info := renderInformation(window, agg, m.now, m.lastUpdate, m.snapshotDir, m.lastSnapshot, framePanel, m.stwTh, m.targetEnv, row1Cols[1].W, row1Cols[1].H)

	gridRows := []gridRow{
		{cellWidths: []int{row1Cols[0].W, row1Cols[1].W}, height: rows[0].H, cells: []string{current, info}},
	}

	if len(rows) >= 2 {
		row2AvailW := rows[1].W - 2 - gridSepX*2
		if row2AvailW < 0 {
			row2AvailW = 0
		}
		// Give the bar chart and details more room; heap chart is still readable at ~40%.
		row2Cols := Cols(Rect{W: row2AvailW, H: rows[1].H}, 0.36, 0.24, 0.40)

		visWindow, visCursor := m.barViewport(window, framePanel, row2Cols[0].W, row2Cols[0].H)
		bar := renderSTWBarChart(visWindow, visCursor, framePanel, m.stwLabelsMode, m.stwTh, 0, row2Cols[0].H, row2Cols[0].W)
		details := renderCycleDetails(visWindow, visCursor, framePanel, m.stwTh, row2Cols[1].W, row2Cols[1].H)
		heap := renderHeapLiveHistory(heapHist, framePanel, chartView{XSpan: m.xSpan.duration(), YZoomSteps: m.heapYZoom, YPanSteps: m.heapYPan, Focused: m.chartFocus == chartHeap}, row2Cols[2].W, row2Cols[2].H)

		gridRows = append(gridRows, gridRow{
			cellWidths: []int{row2Cols[0].W, row2Cols[1].W, row2Cols[2].W},
			height:     rows[1].H,
			cells:      []string{bar, details, heap},
		})
	}

	if len(rows) >= 3 {
		cellW := rows[2].W - 2
		if cellW < 1 {
			cellW = 1
		}
		stw := renderSTWPercentilesHistory(p50Hist, p99Hist, maxHist, framePanel, chartView{XSpan: m.xSpan.duration(), YZoomSteps: m.stwYZoom, YPanSteps: m.stwYPan, Focused: m.chartFocus == chartSTW}, cellW, rows[2].H)
		gridRows = append(gridRows, gridRow{cellWidths: []int{cellW}, height: rows[2].H, cells: []string{stw}})
	}

	app := renderSharedBorderGrid(gridRows, gridW)
	app = m.withFooter(app, content.W)
	app = fitViewport(app, content.W, content.H)

	_ = screen
	return lipgloss.NewStyle().Padding(paddingY, paddingX).Render(app)
}

func (m *Model) displayData() ([]domain.GCEvent, domain.Aggregates, []historyPoint, []historyPoint, []historyPoint, []historyPoint) {
	if m.paused {
		return m.pausedWindow, m.pausedAgg, m.pausedHeapHist, m.pausedSTWP50, m.pausedSTWP99, m.pausedSTWMax
	}
	return m.store.Recent(), m.agg, m.heapHistory, m.stwP50Hist, m.stwP99Hist, m.stwMaxHist
}

func (m *Model) barViewport(window []domain.GCEvent, frame frameMode, w, h int) ([]domain.GCEvent, int) {
	inner := InnerRect(frameStyle(frame), Rect{W: w, H: h})
	_, _, maxBars := stwBarsCapacity(inner.W)
	if maxBars < 1 {
		maxBars = 1
	}
	if len(window) == 0 {
		return nil, 0
	}

	// In LIVE, cursor tracks last element; in PAUSED it is in absolute window coords (m.cursor).
	cursorAbs := len(window) - 1
	if m.paused {
		cursorAbs = m.cursor
		if cursorAbs < 0 {
			cursorAbs = 0
		}
		if cursorAbs >= len(window) {
			cursorAbs = len(window) - 1
		}
	}

	if len(window) <= maxBars {
		return window, cursorAbs
	}

	// In LIVE we always show the latest bars. In PAUSED we allow paging by selecting a window
	// around the cursor position.
	if !m.paused {
		start := len(window) - maxBars
		return window[start:], maxBars - 1
	}

	start := cursorAbs - maxBars/2
	if start < 0 {
		start = 0
	}
	if start > len(window)-maxBars {
		start = len(window) - maxBars
	}

	vis := window[start : start+maxBars]
	return vis, cursorAbs - start
}

func (m *Model) withFooter(app string, w int) string {
	state := "LIVE"
	if m.paused {
		state = "PAUSED"
	}
	footer := lipgloss.NewStyle().Foreground(lipgloss.Color("#5f5f5f")).Render(state + "  q quit | s snapshot | space pause | left/right scrub | ? help")
	if w > 0 {
		footer = ansi.Truncate(footer, w, "")
	}
	return app + "\n" + footer
}

// stackPanels builds vertical rows that always fit into content.H (including gaps).
func stackPanels(content Rect, gapY int, minH int, desired []int) []Rect {
	if len(desired) == 0 || content.H <= 0 || content.W <= 0 {
		return nil
	}
	if minH <= 0 {
		minH = 1
	}

	count := len(desired)
	available := content.H - gapY*(count-1)
	for count > 1 && available < minH*count {
		count--
		available = content.H - gapY*(count-1)
	}
	if available <= 0 {
		return []Rect{{X: content.X, Y: content.Y, W: content.W, H: content.H}}
	}

	desired = desired[:count]
	wanted := 0
	for _, v := range desired {
		if v > 0 {
			wanted += v
		}
	}
	if wanted == 0 {
		wanted = 1
	}

	heights := make([]int, 0, count)
	scale := float64(available) / float64(wanted)
	for _, v := range desired {
		h := int(math.Round(float64(v) * scale))
		if h < minH {
			h = minH
		}
		heights = append(heights, h)
	}

	sum := 0
	for _, h := range heights {
		sum += h
	}
	diff := available - sum
	for diff != 0 {
		adjusted := false
		for i := len(heights) - 1; i >= 0 && diff != 0; i-- {
			if diff > 0 {
				heights[i]++
				diff--
				adjusted = true
				continue
			}
			if heights[i] > minH {
				heights[i]--
				diff++
				adjusted = true
			}
		}
		if !adjusted {
			break
		}
	}

	rows := make([]Rect, 0, len(heights))
	y := content.Y
	for _, h := range heights {
		rows = append(rows, Rect{X: content.X, Y: y, W: content.W, H: h})
		y += h + gapY
	}
	return rows
}

// fitViewport trims output to the available terminal size to avoid scroll/wrap.
func fitViewport(s string, w, h int) string {
	if w <= 0 && h <= 0 {
		return s
	}

	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if h > 0 && len(lines) > h {
		lines = lines[:h]
	}
	if w > 0 {
		for i, ln := range lines {
			lines[i] = ansi.Truncate(ln, w, "")
		}
	}

	return strings.Join(lines, "\n")
}
