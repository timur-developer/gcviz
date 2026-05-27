# gcviz - Go Garbage Collector Visualizer

![gcviz demo](assets/demo_hero.readme.gif)

Read this in other languages: [English](../README.md)

`gcviz` - TUI-инструмент для live-наблюдения за поведением Go GC: GC-циклы, STW-паузы, динамика heap (live/goal) и сигналы GC pacer, прямо в терминале.

Инструмент сделан для "быстрой обратной связи" при перформанс-работе:

- быстро заметить STW пики (p99/max) под нагрузкой
- увидеть изменение частоты GC между запусками
- увидеть, как heap live приближается к heap goal и ужесточается pacing
- сравнить два запуска через snapshots (`diff`)

## Как это работает

У `gcviz` два источника данных:

- `run` (основной): запускает ваш бинарник, гарантирует `GODEBUG` с `gctrace=1,gcpacertrace=1`, парсит `stderr` таргета.
- `attach` (вторичный): опрашивает HTTP endpoint, который отдает `runtime/metrics` в JSON-формате, понятном `gcviz` (через `pkg/reporter`).

Для `run` не нужно менять код приложения. Для `attach` потребуется добавить небольшой HTTP endpoint в сервис.

## Быстрый старт (1 минута)

![gcviz launch](assets/demo_launch.readme.gif)

Требования: Go 1.22+, достаточно большой терминал.

### 1) Запуск демо-нагрузки

Из исходников (без установки):

```bash
go run ./cmd/gcviz lab churn
```

Или через Makefile:

```bash
make lab-churn
```

Help в приложении: `?` / `h` / `f1`.

### 2) Запуск на вашем бинарнике (пошагово)

1. Соберите ваш сервис/приложение в бинарник:

```bash
go build -o ./myapp ./cmd/myapp
```

2. Запустите под наблюдением (обратите внимание на разделитель `--` для аргументов вашей программы):

```bash
go run ./cmd/gcviz run ./myapp -- --your-flag value
```

3. В UI:

- нажмите `?`, чтобы увидеть все хоткеи
- `space` - пауза/продолжить
- в паузе `left/right` (и `home/end`) листают историю
- `s` сохраняет snapshot в `tmp/snapshots` (по умолчанию)

### 3) Установка (опционально)

Установить `gcviz` в `GOBIN`:

```bash
go install github.com/timur-developer/gcviz/cmd/gcviz@latest
```

## Использование

### run (основной)

```bash
gcviz run ./path/to/your-binary -- --your-flag value
```

Из исходников:

```bash
go run ./cmd/gcviz run ./path/to/your-binary -- --your-flag value
```

Через Makefile:

```bash
make run TARGET=./path/to/your-binary ARGS="-- --your-flag value"
```

### lab

Встроенные демо-пресеты:

```bash
gcviz lab alloc
gcviz lab churn
gcviz lab idle
gcviz lab spike
```

### attach (вторичный)

Подключение к уже работающему сервису, который отдает `runtime/metrics` в JSON-формате `gcviz`.

1. Добавьте `pkg/reporter` в сервис:

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

2. Подключитесь:

```bash
gcviz attach http://127.0.0.1:8080/gcviz/metrics
```

Примечания:

- данные основаны на `runtime/metrics`, поэтому они отличаются от `run` режима
- env таргета (`GOGC`, `GOMEMLIMIT`, `GODEBUG`) недоступен, UI покажет `n/a`

### diff

Сравнение двух snapshots:

```bash
gcviz diff ./a.json ./b.json
```

## Что показывает UI (метрики и панели)

`gcviz` хранит скользящее окно последних GC событий (`--window-size`, по умолчанию: 200) и показывает как значения по циклам, так и агрегаты по окну.

### Current Values

- `GC cycles total`: текущий номер GC цикла
- `last STW (us)`: STW пауза последнего цикла (sweep term + mark term, в микросекундах)
- `heap live (MB)` / `heap goal (MB)`: live heap и цель
- `heap: live/goal`: компактный индикатор соотношения live/goal

### Information (агрегаты по окну)

- `max STW (us)`: максимум STW по окну
- `gc`: `GCs/min` и/или средний интервал GC
- `stw`: число/процент "bad" STW (по порогам) и количество forced GC
- `time since last GC`, `uptime`
- `stw thresholds`: `warn` / `bad` (см. настройки)
- состояние snapshot и директория snapshot
- env контекст (`GOGC`, `GOMEMLIMIT`, `GODEBUG`) в `run`/`lab` (в `attach` недоступен)

### Графики

- **Heap live over time (MB)**: heap live во времени
- **STW p50/p99/max over time (us)**: p50/p99/max STW по окну
- **STW per cycle**: bar chart по циклам; подписи можно переключать на STW или heap live (`l`)

### Cycle Details (детали выбранного цикла)

- GC #, time since start, forced
- STW total (us) + разбивка: sweep term / mark term
- heap (MB): start/end и live/goal
- gc cpu (%)
- pacer сигналы (если доступны): assist ratio, assist workers, pages swept

## Хоткеи

![gcviz features](assets/demo_features.readme.gif)

Информацию о всех горячих клавишах всегда можно посмотреть в Help (`?` / `h` / `f1`).

База:

- `?` / `h` / `f1` показать/скрыть Help
- `q` / `ctrl+c` выход
- `space` пауза/продолжить обновления
- `left` / `right` листать историю (в паузе)
- `home` / `end` в начало/конец истории (в паузе)
- `s` сохранить snapshot

Интерфейс:

- `g` переключить layout (spaced/tight)
- `l` режим подписей STW bar chart (GC+STW -> GC+Heap -> GC-only)

Графики:

- `z` выбрать активный график (Heap/STW). Zoom/pan применяется к активному графику.
- `+` / `-` Y-zoom активного графика
- `0` сброс Y zoom/pan активного графика
- `shift+up` / `shift+down` Y-pan активного графика
- `[` / `]` X-zoom (масштаб по времени): all -> 1h -> 15m -> 5m -> 1m (и обратно)
- `r` полный сброс focus, zoom/pan и time span

## Настройки

Глобальные флаги (и env-оверрайды):

- `--window-size` (`GCVIZ_WINDOW_SIZE`) количество событий в памяти (default: 200)
- `--snapshot-path` (`GCVIZ_SNAPSHOT_PATH`) директория snapshots (default: `tmp/snapshots`)
- `--exit-snapshot` (`GCVIZ_EXIT_SNAPSHOT`) snapshot на выходе (default: true)
- `--no-alt-screen` (`GCVIZ_NO_ALT_SCREEN`) отключить alt screen buffer
- `--stw-warn-us` (`GCVIZ_STW_WARN_US`) порог warn для STW (default: 200)
- `--stw-bad-us` (`GCVIZ_STW_BAD_US`) порог bad для STW (default: 1000)

Режимные env:

- `GCVIZ_RUN_TARGET`
- `GCVIZ_ATTACH_URL`, `GCVIZ_POLL_INTERVAL`
- `GCVIZ_LAB_PRESET`
- `GCVIZ_DIFF_A`, `GCVIZ_DIFF_B`

## Snapshots

- Директория по умолчанию: `tmp/snapshots`
- Ручной snapshot: `s`
- Snapshot на выходе включен по умолчанию; пропускается, если недавно делали ручной

## Notes / FAQ

- В `attach` режиме нельзя узнать env таргета (`GOGC`, `GOMEMLIMIT`, `GODEBUG`), поэтому UI показывает `n/a`.
- Если вы не видите обновлений, возможно, приложение пока не делает GC (попробуйте более аллоцирующую нагрузку или `lab churn`).
- Если терминал ведет себя странно, попробуйте `--no-alt-screen` (или `GCVIZ_NO_ALT_SCREEN=true`).
- Очень маленькие STW значения могут отображаться как 0 из-за форматирования `gctrace`.
