package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/timur-developer/gcviz/internal/domain"
)

func renderCurrentValues(agg domain.Aggregates) string {
	var b strings.Builder

	if !agg.HasData {
		fmt.Fprintf(&b, "GC cycles total: -\n")
		fmt.Fprintf(&b, "last STW (µs):   -\n")
		fmt.Fprintf(&b, "heap live (MB):  -\n")
		fmt.Fprintf(&b, "heap goal (MB):  -\n")
		return boxed("Current Values", b.String())
	}

	fmt.Fprintf(&b, "GC cycles total: %d\n", agg.Current.GCCyclesTotal)
	fmt.Fprintf(&b, "last STW (µs):   %d\n", agg.Current.LastSTWUs)
	fmt.Fprintf(&b, "heap live (MB):  %d\n", agg.Current.HeapLiveMB)
	fmt.Fprintf(&b, "heap goal (MB):  %d\n", agg.Current.HeapGoalMB)

	return boxed("Current Values", b.String())
}

func renderInformation(agg domain.Aggregates, now time.Time, lastUpdate time.Time) string {
	var b strings.Builder

	if lastUpdate.IsZero() {
		fmt.Fprintf(&b, "time since last GC: -\n")
	} else {
		fmt.Fprintf(&b, "time since last GC: %s\n", formatSeconds(now.Sub(lastUpdate)))
	}

	if !agg.HasData {
		fmt.Fprintf(&b, "max STW (µs):     -\n")
		fmt.Fprintf(&b, "uptime:           -\n")
		return boxed("Information", b.String())
	}

	fmt.Fprintf(&b, "max STW (µs):     %d\n", agg.Window.STWMaxUs)
	fmt.Fprintf(&b, "uptime:           %s\n", formatDuration(agg.TargetUptime))
	return boxed("Information", b.String())
}

func renderHelp(width, height int) string {
	body := strings.Join([]string{
		"q       quit",
		"ctrl+c  quit",
		"?       toggle help",
	}, "\n")

	box := boxed("Help", body)

	if width > 0 && height > 0 {
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
	}
	return box
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		return "0s"
	}

	// Keep it compact for MVP.
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		min := int(d / time.Minute)
		sec := int((d % time.Minute) / time.Second)
		return fmt.Sprintf("%dm%02ds", min, sec)
	}

	h := int(d / time.Hour)
	min := int((d % time.Hour) / time.Minute)
	sec := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%dh%02dm%02ds", h, min, sec)
}

func formatSeconds(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
