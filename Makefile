build:
	go build -o ./bind/kvstok ./cmd/cli/main.go

run: build
	./bin/kvstok

test:
	go test -v ./... -count=1
