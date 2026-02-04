.PHONY: build run test clean docker-build docker-up docker-down lint

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOVET=$(GOCMD) vet

# Binary names
CURRENCY_BINARY=currency
GATEWAY_BINARY=gateway

# Build directories
BUILD_DIR=bin

all: build

build: build-currency build-gateway

build-currency:
	$(GOBUILD) -o $(BUILD_DIR)/$(CURRENCY_BINARY) ./currency/cmd/currency

build-gateway:
	$(GOBUILD) -o $(BUILD_DIR)/$(GATEWAY_BINARY) ./gateway/cmd/gateway

run-currency:
	$(GOCMD) run ./currency/cmd/currency

run-gateway:
	$(GOCMD) run ./gateway/cmd/gateway

test:
	$(GOTEST) -v -race -cover ./...

test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

tidy:
	$(GOMOD) tidy

lint:
	$(GOVET) ./...

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart: docker-down docker-up

# Development helpers
dev-db:
	docker-compose up -d postgres

dev-currency: dev-db
	sleep 2
	DB_HOST=localhost $(GOCMD) run ./currency/cmd/currency

dev-gateway:
	CURRENCY_SERVICE_URL=http://localhost:8081 $(GOCMD) run ./gateway/cmd/gateway
