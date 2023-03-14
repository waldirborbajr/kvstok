build:
  go build -o bin/kvstok cmd/cli/main.go

run: build
	./bin/kvstok

test:
	go test -v ./... -count=1
