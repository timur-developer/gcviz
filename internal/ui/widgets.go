package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/timur-developer/gcviz/internal/domain"
)

const (
	barChartWidth  = 30
	barChartHeight = 8
)

var (
	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(0, 1)
	titleStyle = lipgloss.NewStyle().Bold(true)
)

func boxed(title, body string) string {
	body = strings.TrimRight(body, "\n")

	content := titleStyle.Render(title) + "\n" + body
	return boxStyle.Render(content)
}

func joinPanels(width int, left, right string) string {
	if width <= 0 {
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}

	if width < 80 {
		return lipgloss.JoinVertical(lipgloss.Left, left, right)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func renderSTWBarChart(window []domain.GCEvent, maxBars int, height int) string {
	values := lastN(window, maxBars)
	if len(values) == 0 {
		return boxed("STW per cycle", "(max: -µs)\n\n(no data)")
	}

	stwUs := make([]int64, 0, len(values))
	var max int64
	for _, ev := range values {
		v := stwPerCycleUs(ev)
		stwUs = append(stwUs, v)
		if v > max {
			max = v
		}
	}

	lines := renderBars(stwUs, max, height)
	axis := strings.Repeat("─", len(values))
	body := fmt.Sprintf("(max: %dµs)\n%s\n%s", max, strings.Join(lines, "\n"), axis)
	return boxed("STW per cycle", body)
}

func lastN(window []domain.GCEvent, n int) []domain.GCEvent {
	if n <= 0 || len(window) == 0 {
		return nil
	}
	if len(window) <= n {
		return window
	}
	return window[len(window)-n:]
}

func stwPerCycleUs(ev domain.GCEvent) int64 {
	ms := ev.STWSweepTermMs + ev.STWMarkTermMs
	return int64(math.Round(ms * 1000))
}

func renderBars(values []int64, max int64, height int) []string {
	if height <= 0 {
		height = 1
	}
	if max <= 0 {
		max = 1
	}

	heights := make([]int, 0, len(values))
	for _, v := range values {
		h := int(math.Round(float64(v) / float64(max) * float64(height)))
		if h < 0 {
			h = 0
		}
		if h > height {
			h = height
		}
		heights = append(heights, h)
	}

	out := make([]string, 0, height)
	for row := height; row >= 1; row-- {
		var b strings.Builder
		for _, h := range heights {
			if h >= row {
				b.WriteRune('█')
			} else {
				b.WriteRune(' ')
			}
		}
		out = append(out, b.String())
	}
	return out
}
