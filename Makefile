COMMIT_HASH := $(shell git rev-parse HEAD)
PROJECT := gobrowser

GOCMD := go
GOTEST := $(GOCMD) test
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOINSTALL := $(GOCMD) install
GOPATH := $(shell $(GOCMD) env GOPATH)
GOIMPORTS := $(GOPATH)/bin/goimports

# ------------------------------------------------------------
# Tool Installation Targets
# ------------------------------------------------------------
.PHONY: install-tools
install-tools:
	$(GOINSTALL) gioui.org/cmd/gogio@latest

.PHONY: linter
linter:
	@echo "Installing linter..."
	$(GOINSTALL) github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

# ------------------------------------------------------------
# Service Management Targets
# ------------------------------------------------------------
.PHONY: run-app
run-app:
	@echo "Running $(PROJECT) service..."
	$(GOCMD) run ./cmd/main.go --debug

.PHONY: build-app
build-app:
	@echo "Building $(PROJECT) service..."
	$(GOBUILD) -ldflags "-s -w" -o ./build/$(PROJECT) ./cmd/main.go

.PHONY: run-build
run-build: build-app
	@echo "Running $(PROJECT) service..."
	./build/$(PROJECT)

# ------------------------------------------------------------
# Testing & Quality Assurance Targets
# ------------------------------------------------------------
.PHONY: test
test:
	$(GOTEST) $(TEST_PACKAGES) -coverprofile=coverage.out

.PHONY: coverage
coverage:
	$(GOCMD) tool cover -func=coverage.out

.PHONY: lint
lint: linter
	$(shell $(GOCMD) env GOPATH)/bin/golangci-lint run ./...


.PHONY: fix-lint
fix-lint: linter
	@echo "Running golangci-lint --fix..."
	$(shell $(GOCMD) env GOPATH)/bin/golangci-lint run --fix  ./...