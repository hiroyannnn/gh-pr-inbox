GOFMT_TARGETS := ./cmd ./internal ./main.go
MODULE_FILES := go.mod go.sum
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod
GOENV := env GOCACHE="$(GOCACHE)" GOMODCACHE="$(GOMODCACHE)"
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/hiroyannnn/gh-pr-inbox/internal/buildinfo.Version=$(VERSION)

.PHONY: all clean help fmt fmt-check tidy tidy-check vet test lint build ci install-local reinstall-local reinstall-prod release delete-tag

all: build ## Default target: build the binary

clean: ## Remove built artifacts
	rm -f gh-pr-inbox

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

fmt: ## Format Go files
	gofmt -w $(GOFMT_TARGETS)

fmt-check: ## Check Go formatting
	@fmt_out=$$(gofmt -l $(GOFMT_TARGETS)); \
	if [ -n "$$fmt_out" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$fmt_out"; \
		exit 1; \
	fi

tidy: ## Run go mod tidy
	$(GOENV) go mod tidy

tidy-check: ## Ensure go.mod/go.sum are tidy
	$(GOENV) go mod tidy
	@git diff --quiet -- $(MODULE_FILES) || (echo "go.mod/go.sum are not tidy"; git diff -- $(MODULE_FILES); exit 1)

vet: ## Run go vet
	$(GOENV) go vet ./...

test: ## Run tests
	$(GOENV) go test ./...

lint: fmt-check vet ## Run linters

build: ## Build the binary
	$(GOENV) go build -ldflags "$(LDFLAGS)" -o gh-pr-inbox .

ci: tidy-check lint test ## Run CI tasks

install-local: build ## Install extension from local checkout
	gh extension install .

reinstall-local: ## Reinstall extension from local checkout
	@echo "Reinstalling extension from local checkout..."
	@gh extension remove gh-pr-inbox 2>/dev/null || true
	@$(MAKE) install-local
	@echo "Done"

reinstall-prod: ## Reinstall extension from GitHub
	@echo "Reinstalling extension from GitHub..."
	@gh extension remove gh-pr-inbox 2>/dev/null || true
	@gh extension install hiroyannnn/gh-pr-inbox
	@echo "Done"

release: ## Create a new release tag (make release version=1.0.0)
	@if [ -z "$(version)" ]; then \
		echo "Specify a version (e.g. make release version=1.0.0)"; \
		exit 1; \
	fi
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working tree is dirty; commit or stash changes first"; \
		exit 1; \
	fi
	@if git rev-parse "v$(version)" >/dev/null 2>&1; then \
		echo "Tag v$(version) already exists"; \
		echo "Use a new version, or run 'make delete-tag version=$(version)'"; \
		exit 1; \
	fi
	@echo "Running CI tasks..."
	@if ! $(MAKE) ci; then \
		echo "CI tasks failed; run 'make ci' for details"; \
		exit 1; \
	fi
	@echo "Releasing v$(version)..."
	@git tag "v$(version)"
	@git push origin "v$(version)"
	@echo "GitHub Actions release build started:"
	@echo "https://github.com/hiroyannnn/gh-pr-inbox/actions"

delete-tag: ## Delete a tag (make delete-tag version=1.0.0)
	@if [ -z "$(version)" ]; then \
		echo "Specify a version (e.g. make delete-tag version=1.0.0)"; \
		exit 1; \
	fi
	@if ! git rev-parse "v$(version)" >/dev/null 2>&1; then \
		echo "Tag v$(version) does not exist"; \
		exit 1; \
	fi
	@echo "Deleting tag v$(version)..."
	@git push origin --delete "v$(version)"
	@git tag -d "v$(version)"
	@echo "Done"
