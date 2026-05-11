# Justfile for kvstok

BIN := "./bin/kvstok"

# Default recipe: build
default: build

# Run tests
test:
    go test -v -timeout=1s -race -covermode=atomic -count=1 ./...

# Build the binary (depends on test)
build: test
    mkdir -p bin
    go build -o {{BIN}} .

# Run the binary (depends on build)
run: build
    {{BIN}}

# Update dependencies
update:
    go get -u all

# Clean build artifacts
clean:
    rm -rf bin/

# Format code
fmt:
    go fmt ./...

# Vet code
vet:
    go vet ./...

# Lint code (assuming golangci-lint is available)
lint:
    golangci-lint run

# Install dependencies
deps:
    go mod download

# Tidy modules
tidy:
    go mod tidy