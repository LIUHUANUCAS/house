# Define the binary name
BINARY_NAME=myapp
VERSION=1.0.0

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=`date +%Y-%m-%d_%H:%M:%S`"

.PHONY: all build clean test run install uninstall fmt lint vet

all: build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

clean:
	go clean
	rm -f $(BINARY_NAME)

test:
	go test -v ./...

run: build
	./$(BINARY_NAME)

install: build
	cp $(BINARY_NAME) /usr/local/bin

uninstall:
	rm -f /usr/local/bin/$(BINARY_NAME)

fmt:
	go fmt ./...

lint:
	golint ./...

vet:
	go vet ./...

# Cross-compilation targets
build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux ./cmd/myapp

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows.exe ./cmd/myapp

build-darwin:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin ./cmd/myapp

# Build all platforms
build-all: build-linux build-windows build-darwin

# Docker targets
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

docker-run:
	docker run --rm -p 8080:8080 $(BINARY_NAME):$(VERSION)