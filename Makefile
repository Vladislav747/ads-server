APP_NAME := simple-ads-server
BIN_DIR := bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)
MAIN_PKG := ./cmd/server
GEOIP_DB := GeoLite2-Country.mmdb

.PHONY: all build run clean test tidy deps docker-up docker-down check-geoip

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

check-geoip:
	@if [ ! -f "$(GEOIP_DB)" ]; then \
		echo "Отсутствует $(GEOIP_DB). Скачайте GeoLite2-Country.mmdb и положите в корень проекта."; \
		exit 1; \
	fi
