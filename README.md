# gcviz

TUI visualizer for Go GC behavior.

## Install

```bash
go install github.com/timur-developer/gcviz/cmd/gcviz@latest
```

## Usage

### Lab mode

Run a built-in demo workload (no external services required):

```bash
gcviz lab alloc
gcviz lab churn
gcviz lab idle
gcviz lab spike
```

### Run mode

Run a target binary under observation (stderr is parsed from `GODEBUG=gctrace=1,gcpacertrace=1`):

```bash
gcviz run ./myservice -- --config ./config.yml
```

### Attach mode

Attach to a running service exposing runtime metrics:

```bash
gcviz attach http://127.0.0.1:6060/debug/metrics/v1
```

### Snapshot & diff

```bash
gcviz diff ./a.json ./b.json
```
