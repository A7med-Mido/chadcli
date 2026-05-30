BINARY   := chad
CMD_PATH := ./cmd/
BUILD_DIR := ./dist

.PHONY: all build run install clean deps tidy

all: build

## Download dependencies
deps:
	go mod download

## Tidy go.mod / go.sum
tidy:
	go mod tidy

## Build the binary
build: tidy
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)
	@echo "Binary: $(BUILD_DIR)/$(BINARY)"

## Run without installing
run: build
	$(BUILD_DIR)/$(BINARY) .

## Install to $GOPATH/bin (or ~/go/bin)
install: tidy
	go install $(CMD_PATH)
	@echo "Installed: $$(go env GOPATH)/bin/$(BINARY)"

## Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## Format all Go source files
fmt:
	gofmt -w .

## Run go vet
vet:
	go vet ./...