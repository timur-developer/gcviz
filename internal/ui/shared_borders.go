package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	bdTL    = '\u250c' // ┌
	bdTR    = '\u2510' // ┐
	bdBL    = '\u2514' // └
	bdBR    = '\u2518' // ┘
	bdH     = '\u2500' // ─
	bdV     = '\u2502' // │
	bdCross = '\u253c' // ┼
	bdT     = '\u252c' // ┬
	bdB     = '\u2534' // ┴
	bdL     = '\u251c' // ├
	bdR     = '\u2524' // ┤
)

type gridRow struct {
	cellWidths []int
	height     int
	cells      []string
}

func renderSharedBorderGrid(rows []gridRow, w int) string {
	if w <= 0 {
		return ""
	}
	if len(rows) == 0 {
		return borderStyle().Render(emptyBorderLine(w, bdTL, bdTR, bdH))
	}

	sepSets := make([]map[int]struct{}, 0, len(rows))
	for _, r := range rows {
		sepSets = append(sepSets, sepSet(r.cellWidths))
	}

	var out []string

	// Top border.
	out = append(out, borderStyle().Render(borderLine(w, sepSets[0], bdTL, bdTR, bdT, bdH)))

	// Content rows.
	for i, r := range rows {
		out = append(out, renderGridRowContent(r, w)...)
		if i < len(rows)-1 {
			above := sepSets[i]
			below := sepSets[i+1]
			out = append(out, borderStyle().Render(rowSeparatorLine(w, above, below)))
		}
	}

	// Bottom border.
	out = append(out, borderStyle().Render(borderLine(w, sepSets[len(rows)-1], bdBL, bdBR, bdB, bdH)))

	return strings.Join(out, "\n")
}

func renderGridRowContent(r gridRow, totalW int) []string {
	if r.height < 1 {
		r.height = 1
	}
	if len(r.cellWidths) == 0 || len(r.cells) == 0 {
		lines := make([]string, 0, r.height)
		for i := 0; i < r.height; i++ {
			lines = append(lines, borderStyle().Render("\u2502")+strings.Repeat(" ", max(0, totalW-2))+borderStyle().Render("\u2502"))
		}
		return lines
	}

	cellLines := make([][]string, 0, len(r.cells))
	for i := range r.cells {
		cl := normalizeLines(r.cells[i], r.height)
		cellLines = append(cellLines, cl)
	}

	sep := borderStyle().Render("\u2502")

	lines := make([]string, 0, r.height)
	for row := 0; row < r.height; row++ {
		var b strings.Builder
		b.WriteString(sep)
		for i := range r.cells {
			if i < len(cellLines) && row < len(cellLines[i]) {
				b.WriteString(cellLines[i][row])
			} else {
				b.WriteString(strings.Repeat(" ", r.cellWidths[i]))
			}
			if i < len(r.cells)-1 {
				b.WriteString(sep)
			}
		}
		b.WriteString(sep)
		lines = append(lines, b.String())
	}
	return lines
}

func sepSet(cellWidths []int) map[int]struct{} {
	set := map[int]struct{}{}
	if len(cellWidths) <= 1 {
		return set
	}
	x := 1
	for i, cw := range cellWidths {
		if cw < 0 {
			cw = 0
		}
		x += cw
		if i < len(cellWidths)-1 {
			set[x] = struct{}{}
			x++
		}
	}
	return set
}

func borderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(borderColor)
}

func borderLine(w int, seps map[int]struct{}, left, right, cross, fill rune) string {
	if w <= 0 {
		return ""
	}
	if w == 1 {
		return string(left)
	}

	var b strings.Builder
	for x := 0; x < w; x++ {
		switch x {
		case 0:
			b.WriteRune(left)
		case w - 1:
			b.WriteRune(right)
		default:
			if _, ok := seps[x]; ok {
				b.WriteRune(cross)
			} else {
				b.WriteRune(fill)
			}
		}
	}
	return b.String()
}

func emptyBorderLine(w int, left, right, fill rune) string {
	if w <= 0 {
		return ""
	}
	if w == 1 {
		return string(left)
	}
	return string(left) + strings.Repeat(string(fill), w-2) + string(right)
}

func rowSeparatorLine(w int, aboveSeps, belowSeps map[int]struct{}) string {
	if w <= 0 {
		return ""
	}
	if w == 1 {
		return "\u251c"
	}

	var b strings.Builder
	for x := 0; x < w; x++ {
		switch x {
		case 0:
			b.WriteRune(bdL)
		case w - 1:
			b.WriteRune(bdR)
		default:
			_, up := aboveSeps[x]
			_, down := belowSeps[x]
			switch {
			case up && down:
				b.WriteRune(bdCross)
			case up && !down:
				b.WriteRune(bdB)
			case !up && down:
				b.WriteRune(bdT)
			default:
				b.WriteRune(bdH)
			}
		}
	}
	return b.String()
}

func normalizeLines(s string, h int) []string {
	if h < 1 {
		h = 1
	}

	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	if len(lines) > h {
		lines = lines[:h]
	}
	for len(lines) < h {
		lines = append(lines, "")
	}
	return lines
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
