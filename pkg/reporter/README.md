# reporter

Package `reporter` exposes `runtime/metrics` over HTTP in a format that `gcviz attach` can consume.

Use it when you want to connect `gcviz` to an already running service (attach mode). In contrast, `gcviz run` does not require any code changes in your application.

Default endpoint: `GET /gcviz/metrics`.

## Quickstart

`rep.Path()` returns the URL path to mount the endpoint (default: `/gcviz/metrics`).

`rep.Handler()` returns an `http.Handler` that serves a JSON payload with `runtime/metrics` samples.

### net/http

If you already have an HTTP server, register the handler in your `http.ServeMux`:

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

Attach:

```bash
gcviz attach http://127.0.0.1:8080/gcviz/metrics
```

### chi

The same handler can be mounted into `chi` router:

```go
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/timur-developer/gcviz/pkg/reporter"
)

func main() {
	rep := reporter.New()

	r := chi.NewRouter()
	
	r.Handle(rep.Path(), rep.Handler())

	log.Fatal(http.ListenAndServe(":8080", r))
}
```

## Custom Path

```go
rep := reporter.New(reporter.WithPath("/debug/gcviz/metrics"))
```
