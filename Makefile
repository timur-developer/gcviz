.PHONY: help lint test build ci install run lab lab-alloc lab-churn lab-idle lab-spike attach diff testbin release-snapshot

GOCMD ?= go
GCSCOPE_RUN := $(GOCMD) run ./cmd/gcscope

PRESET ?= alloc
URL ?= http://127.0.0.1:8080/gcscope/metrics

help:
	@echo "Targets:"
	@echo "  make lint                Run golangci-lint"
	@echo "  make test                Run go tests"
	@echo "  make build               Build all packages (sanity check)"
	@echo "  make ci                  Lint + test + build"
	@echo "  make install             Install gcscope into GOPATH/bin"
	@echo ""
	@echo "Run modes (zero-guess):"
	@echo "  make lab                 Run lab preset (default PRESET=alloc)"
	@echo "  make lab-alloc            "
	@echo "  make lab-churn            "
	@echo "  make lab-idle             "
	@echo "  make lab-spike            "
	@echo "  make run TARGET=./app ARGS='-- --config ./cfg.yml'"
	@echo "  make attach               (default URL=$(URL))"
	@echo "  make attach URL=http://127.0.0.1:8080/gcscope/metrics"
	@echo "  make diff A=./a.json B=./b.json"
	@echo ""
	@echo "Maintainers:"
	@echo "  make testbin             Rebuild embedded testbin binaries"
	@echo "  make release-snapshot     Local goreleaser build (no publish)"

lint:
	golangci-lint run

test:
	$(GOCMD) test ./...

build:
	$(GOCMD) build ./...

ci: lint test build

install:
	$(GOCMD) install ./cmd/gcscope

lab:
	$(GCSCOPE_RUN) lab $(PRESET)

lab-alloc:
	$(GCSCOPE_RUN) lab alloc

lab-churn:
	$(GCSCOPE_RUN) lab churn

lab-idle:
	$(GCSCOPE_RUN) lab idle

lab-spike:
	$(GCSCOPE_RUN) lab spike

run:
	$(GCSCOPE_RUN) run $(TARGET) $(ARGS)

attach:
	$(GCSCOPE_RUN) attach $(URL)

diff:
	$(GCSCOPE_RUN) diff $(A) $(B)

testbin:
	$(GOCMD) run ./internal/devtools/testbinbuild

release-snapshot:
	goreleaser release --snapshot --clean
