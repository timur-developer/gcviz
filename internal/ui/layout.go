package ui

import "github.com/charmbracelet/lipgloss"

type Rect struct {
	X int
	Y int
	W int
	H int
}

func (r Rect) ClampMinSize(minW, minH int) Rect {
	if r.W < minW {
		r.W = minW
	}
	if r.H < minH {
		r.H = minH
	}
	return r
}

func Rows(parent Rect, fracs ...float64) []Rect {
	out := make([]Rect, 0, len(fracs))
	y := parent.Y
	used := 0
	for i, f := range fracs {
		h := int(float64(parent.H) * f)
		if i == len(fracs)-1 {
			h = parent.H - used
		}
		if h < 0 {
			h = 0
		}
		out = append(out, Rect{X: parent.X, Y: y, W: parent.W, H: h})
		y += h
		used += h
	}
	return out
}

func Cols(parent Rect, fracs ...float64) []Rect {
	out := make([]Rect, 0, len(fracs))
	x := parent.X
	used := 0
	for i, f := range fracs {
		w := int(float64(parent.W) * f)
		if i == len(fracs)-1 {
			w = parent.W - used
		}
		if w < 0 {
			w = 0
		}
		out = append(out, Rect{X: x, Y: parent.Y, W: w, H: parent.H})
		x += w
		used += w
	}
	return out
}

func InnerRect(style lipgloss.Style, r Rect) Rect {
	fx, fy := style.GetFrameSize()
	w := r.W - fx
	h := r.H - fy
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return Rect{X: 0, Y: 0, W: w, H: h}
}
