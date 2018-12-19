# Go parameters
GOCMD=go
DEPCMD=dep
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GODEP=$(DEPCMD) ensure
BINARY_NAME=creepycrawler
BINARY_UNIX=$(BINARY_NAME)_unix

all: test build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v cmd/creepycrawler/main.go
test:
	$(GOTEST) -v ./...
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
deps:
	$(GODEP)