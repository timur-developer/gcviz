# reporter

Пакет `reporter` позволяет легко экспонировать `runtime/metrics` по HTTP, чтобы `gcviz attach` мог подключиться к работающему сервису.

## Быстрый старт

По умолчанию endpoint: `GET /gcviz/metrics`.

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

Подключение:

```bash
gcviz attach http://localhost:8080/gcviz/metrics
```

## Кастомный путь

```go
rep := reporter.New(reporter.WithPath("/debug/runtime/metrics"))
```
