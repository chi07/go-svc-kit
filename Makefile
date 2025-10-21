APP := go-svc-kit
PKG := ./...
BIN := ./bin/$(APP)

# Tools (installed into $GOBIN)
GOLANGCI_VERSION ?= v1.61.0

.PHONY: all build run tidy fmt vet lint test cover vuln verify tools clean

all: verify build

build:
	@mkdir -p bin
	@go build -trimpath -ldflags="-s -w" -o $(BIN) ./cmd/$(APP)

run:
	@go run ./cmd/$(APP)

tidy:
	@go mod tidy

fmt:
	@go fmt $(PKG)
	@# gofumpt provides stricter formatting; install: go install mvdan.cc/gofumpt@latest
	@if command -v gofumpt >/dev/null 2>&1; then gofumpt -l -w .; fi

vet:
	@go vet $(PKG)

lint: tools
	@golangci-lint run

test:
	@go test -race -shuffle=on $(PKG)

cover:
	@go test -race -covermode=atomic -coverprofile=coverage.out $(PKG)
	@go tool cover -func=coverage.out | tail -1

vuln:
	@if command -v govulncheck >/dev/null 2>&1; then govulncheck $(PKG); else echo "govulncheck not installed (go install golang.org/x/vuln/cmd/govulncheck@latest)"; fi

verify: tidy fmt vet lint test

tools:
	@command -v golangci-lint >/dev/null 2>&1 || \
		( echo "Installing golangci-lint $(GOLANGCI_VERSION)..." && \
		  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		  sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_VERSION) )

clean:
	@rm -rf bin coverage.out
