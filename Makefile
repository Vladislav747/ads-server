APP_NAME := simple-ads-server
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)
MAIN_PKG := ./cmd/server

.PHONY: all build run clean test tidy deps docker-up docker-down

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_PATH) $(MAIN_PKG)

run: check-geoip
	go run $(MAIN_PKG)

clean:
	rm -rf $(BIN_DIR)

test:
	go test ./...

tidy:
	go mod tidy

deps:
	go mod download

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down