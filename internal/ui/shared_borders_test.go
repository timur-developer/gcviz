package ui

import (
	"strings"
	"testing"
)

func TestRenderSharedBorderGrid_ContainsBoxDrawing(t *testing.T) {
	out := renderSharedBorderGrid([]gridRow{{cellWidths: []int{3}, height: 1, cells: []string{"abc"}}}, 10)
	if len(out) == 0 {
		t.Fatalf("expected non-empty output")
	}
	if !strings.Contains(out, "\u250c") || !strings.Contains(out, "\u2510") || !strings.Contains(out, "\u2514") || !strings.Contains(out, "\u2518") {
		t.Fatalf("expected box-drawing corners in output")
	}
}
