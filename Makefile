.PHONY: build run test integration lint

COMPOSE ?= docker compose
TOOLS   := $(COMPOSE) --profile tools run --rm tools

build:
	$(COMPOSE) build goboxd

run:
	$(COMPOSE) up goboxd

test:
	$(TOOLS) go test ./...

integration:
	$(TOOLS) go test -tags=integration ./tests/...

lint:
	$(TOOLS) golangci-lint run ./...

benchmark:
	$(COMPOSE) up -d goboxd
	$(TOOLS) hey -c 100 -n 2000 -m POST -T "application/json" -D ./tests/testdata/py3/benchmark.json http://goboxd:8080/run
	$(COMPOSE) stop goboxd
