# =========================
# Project configuration
# =========================

APP_NAME        := munchin-api
CMD_DIR         := ./cmd/server
MAIN_FILE       := $(CMD_DIR)/...
BIN_DIR         := ./bin
BIN_FILE        := $(BIN_DIR)/$(APP_NAME)

GO              := go
GOFLAGS         :=
ENV             ?= dev

SWAG            := swag
SWAG_MAIN       := ./cmd/server/main.go

# =========================
# Default target
# =========================

.DEFAULT_GOAL := help

# =========================
# Help
# =========================

.PHONY: help
help:
	@echo ""
	@echo "Available targets:"
	@echo "  make build        Build the API binary"
	@echo "  make run          Run the API locally"
	@echo "  make dev          Run with live reload (if available)"
	@echo "  make test         Run tests"
	@echo "  make lint         Run golangci-lint"
	@echo "  make fmt          Format Go code"
	@echo "  make swagger      Generate Swagger docs"
	@echo "  make clean        Clean build artifacts"
	@echo ""

# =========================
# Build & Run
# =========================

.PHONY: build
build:
	@echo ">> Building $(APP_NAME)"
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_FILE) $(MAIN_FILE)

.PHONY: run
run:
	@echo ">> Running $(APP_NAME) (ENV=$(ENV))"
	ENV=$(ENV) $(GO) run $(MAIN_FILE)

.PHONY: dev
dev:
	@echo ">> Running $(APP_NAME) in dev mode"
	ENV=dev $(GO) run $(MAIN_FILE)

# =========================
# Quality
# =========================

.PHONY: test
test:
	@echo ">> Running tests"
	$(GO) test ./... -race -cover

.PHONY: fmt
fmt:
	@echo ">> Formatting code"
	$(GO) fmt ./...

.PHONY: lint
lint:
	@echo ">> Running golangci-lint"
	golangci-lint run

# =========================
# Swagger / OpenAPI
# =========================

.PHONY: swagger
swagger:
	@echo ">> Generating Swagger docs"
	swag init -g cmd/server/main.go --parseInternal	

# =========================
# Cleanup
# =========================

.PHONY: clean
clean:
	@echo ">> Cleaning build artifacts"
	rm -rf $(BIN_DIR)
	rm -rf ./docs
