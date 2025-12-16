GOFMT_TARGETS := ./cmd ./internal ./main.go
MODULE_FILES := go.mod go.sum

.PHONY: fmt fmt-check tidy tidy-check vet test lint build ci

fmt:
	gofmt -w $(GOFMT_TARGETS)

fmt-check:
	@fmt_out=$$(gofmt -l $(GOFMT_TARGETS)); \
	if [ -n "$$fmt_out" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$fmt_out"; \
		exit 1; \
	fi

tidy:
	go mod tidy

tidy-check:
	go mod tidy
	@git diff --quiet -- $(MODULE_FILES) || (echo "go.mod/go.sum are not tidy"; git diff -- $(MODULE_FILES); exit 1)

vet:
	go vet ./...

test:
	go test ./...

lint: fmt-check vet

build:
	go build ./...

ci: tidy-check lint test
