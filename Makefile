#!/usr/bin/env make -f

test:
	go test -v -timeout=1s -race -covermode=atomic -count=1 ./...

build: test
	go build -o ./bin/kvstok ./cmd/cli/main.go

run: build
	./bin/kvstok

update:
	go get -u all


.PHONY: test build
