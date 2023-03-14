buildapi:
  go build -o bin/kvstokapi cmd/cli/main.go

runapi: buildapi
	./bin/kvstokapi

buildcli:
  go build -o bin/kvstok cmd/cli/main.go

runcli: buildcli
	./bin/kvstok

test:
	go test -v ./... -count=1
