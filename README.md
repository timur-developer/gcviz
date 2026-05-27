# gcviz - Go Garbage Collector Visualizer

![gcviz demo](docs/assets/demo_hero.readme.gif)

[![CI](https://github.com/timur-developer/gcviz/actions/workflows/ci.yml/badge.svg)](https://github.com/timur-developer/gcviz/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/go-1.22%2B-00ADD8?logo=go&logoColor=white)
![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)

Read this in other languages: [Russian](docs/README.ru.md)

`gcviz` is a terminal UI (TUI) visualizer for Go GC behavior: GC cycles, STW pauses, heap live/goal dynamics, and GC pacer signals, live.

It is meant for "fast feedback" during performance work:

- spot bad STW spikes (p99/max) under load
- see GC rate changes across runs
- see heap live approaching heap goal and getting more aggressive pacing
- compare two runs via snapshots (`diff`)

## How It Works

`gcviz` has two data sources:

- `run` (primary): starts your binary, ensures `GODEBUG` contains `gctrace=1,gcpacertrace=1`, parses the target `stderr`.
- `attach` (secondary): polls an HTTP endpoint that exports `runtime/metrics` in a `gcviz`-friendly JSON format (via `pkg/reporter`).

No code changes are required for `run`. For `attach`, you add a small HTTP endpoint to the target service.

## Quickstart (1 minute)

![gcviz launch](docs/assets/demo_launch.readme.gif)

Prereqs: Go 1.22+, a reasonably sized terminal.

### 1) Try the built-in demo workload

From source (no install):

```bash
go run ./cmd/gcviz lab churn
```

Or via Makefile:

```bash
make lab-churn
```

Open in-app help any time: `?` / `h` / `f1`.

### 2) Run on your binary (step-by-step)

1. Build your target app as a binary:

```bash
go build -o ./myapp ./cmd/myapp
```

2. Run it under observation (note the `--` separator for passing args to your program):

```bash
go run ./cmd/gcviz run ./myapp -- --your-flag value
```

3. In the UI:

- press `?` to see all hotkeys
- press `space` to pause/resume
- when paused, use `left/right` (and `home/end`) to scrub history
- press `s` to write a snapshot to `tmp/snapshots` (by default)

### 3) Install (optional)

Install the `gcviz` CLI into your `GOBIN`:

```bash
go install github.com/timur-developer/gcviz/cmd/gcviz@latest
```

Then:

```bash
gcviz lab churn
```

## Usage

### run (primary)

Run your Go program under observation:

```bash
gcviz run ./path/to/your-binary -- --your-flag value
```

From source (no install):

```bash
go run ./cmd/gcviz run ./path/to/your-binary -- --your-flag value
```

Makefile shortcut:

```bash
make run TARGET=./path/to/your-binary ARGS="-- --your-flag value"
```

### lab

Built-in demo presets:

```bash
gcviz lab alloc
gcviz lab churn
gcviz lab idle
gcviz lab spike
```

### attach (secondary)

Attach to a running service that exposes `runtime/metrics` in `gcviz` JSON format.

1. Add `pkg/reporter` to your service:

```go
package main

import (
	"log"
	"net/http"

	"github.com/timur-developer/gcviz/pkg/reporter"
)

func main() {
	rep := reporter.New()

	mux := http.NewServeMux()
	mux.Handle(rep.Path(), rep.Handler())

	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

2. Attach:

```bash
gcviz attach http://127.0.0.1:8080/gcviz/metrics
```

Notes (attach mode):

- the endpoint payload is based on `runtime/metrics`, so values differ from `run` mode
- the target process environment (`GOGC`, `GOMEMLIMIT`, `GODEBUG`) is not available; UI shows `n/a`

### diff

Compare two snapshots:

```bash
gcviz diff ./a.json ./b.json
```

## What You See (Metrics & Panels)

`gcviz` keeps a sliding window of recent GC events (`--window-size`, default: 200) and shows both raw per-cycle values and derived stats over that window.

### Current Values

- `GC cycles total`: current GC cycle number
- `last STW (us)`: last cycle STW pause (sweep term + mark term, converted to microseconds)
- `heap live (MB)` / `heap goal (MB)`: live heap and current goal
- `heap: live/goal`: a compact live-vs-goal indicator

### Information (window stats)

- `max STW (us)`: max STW over the visible window
- `gc`: rate as `GCs/min` and/or average GC interval
- `stw`: bad STW count/percent (based on thresholds) and forced GC count
- `time since last GC`, `uptime`
- `stw thresholds`: `warn` / `bad` (see Configuration)
- snapshot status and snapshot directory
- env context (`GOGC`, `GOMEMLIMIT`, `GODEBUG`) in `run`/`lab` (not available in `attach`)

### Charts

- **Heap live over time (MB)**: time-series of heap live
- **STW p50/p99/max over time (us)**: time-series derived from the sliding window
- **STW per cycle**: per-cycle bar chart; labels can show STW or heap live (toggle with `l`)

### Cycle Details (selected GC event)

For the currently selected cycle (live cursor, or scrubbed history when paused):

- GC #, time since start, forced
- STW total (us) + breakdown: sweep term / mark term
- heap (MB): start/end and live/goal
- gc cpu (%)
- pacer signals (when available): assist ratio, assist workers, pages swept

## Controls

![gcviz features](docs/assets/demo_features.readme.gif)

Full list of hotkeys is always available in the in-app Help (`?` / `h` / `f1`).

Core:

- `?` / `h` / `f1` toggle Help
- `q` / `ctrl+c` quit
- `space` pause/resume live updates
- `left` / `right` scrub history when paused
- `home` / `end` jump to first/last event when paused
- `s` write a snapshot

Layout and labels:

- `g` toggle layout (spaced/tight)
- `l` toggle STW bar labels mode (GC+STW -> GC+Heap -> GC-only)

Charts:

- `z` switch focused chart (Heap/STW). Zoom/pan applies to the focused chart.
- `+` / `-` zoom Y for focused chart
- `0` reset Y zoom/pan for focused chart
- `shift+up` / `shift+down` pan Y for focused chart
- `[` / `]` zoom time span (X axis): all -> 1h -> 15m -> 5m -> 1m (and back)
- `r` reset focus, zoom/pan, and time span

## Configuration

Global flags (and their env overrides):

- `--window-size` (`GCVIZ_WINDOW_SIZE`) samples kept in memory (default: 200)
- `--snapshot-path` (`GCVIZ_SNAPSHOT_PATH`) snapshot directory (default: `tmp/snapshots`)
- `--exit-snapshot` (`GCVIZ_EXIT_SNAPSHOT`) write a snapshot on exit (default: true)
- `--no-alt-screen` (`GCVIZ_NO_ALT_SCREEN`) disable alt screen buffer
- `--stw-warn-us` (`GCVIZ_STW_WARN_US`) STW warning threshold (default: 200)
- `--stw-bad-us` (`GCVIZ_STW_BAD_US`) STW bad threshold (default: 1000)

Mode-specific env vars:

- `GCVIZ_RUN_TARGET`
- `GCVIZ_ATTACH_URL`, `GCVIZ_POLL_INTERVAL`
- `GCVIZ_LAB_PRESET`
- `GCVIZ_DIFF_A`, `GCVIZ_DIFF_B`

## Snapshots

- Default directory: `tmp/snapshots`
- Manual snapshot: press `s`
- Exit snapshot: enabled by default; skipped if a manual snapshot was created recently

## Notes / FAQ

- `attach` mode cannot see the target process env (`GOGC`, `GOMEMLIMIT`, `GODEBUG`), so UI shows `n/a` for these fields.
- If you see no updates, your program might simply not be hitting GC cycles yet (try a workload that allocates more, or use `lab churn`).
- If your terminal behaves oddly, try `--no-alt-screen` (or `GCVIZ_NO_ALT_SCREEN=true`).
- Very small STW values may display as 0 due to `gctrace` formatting.

## Development

```bash
make ci
make lint
make test
make build
```

Maintainers:

```bash
make testbin
make release-snapshot
```

## License

MIT. See [LICENSE](LICENSE).
