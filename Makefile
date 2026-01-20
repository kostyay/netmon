.PHONY: build test fmt lint run clean security security-full

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags="-X main.Version=$(VERSION)" -o bin/netmon ./cmd/netmon

test:
	go test ./...

fmt:
	gofmt -w .

lint:
	golangci-lint run

run:
	go run ./cmd/netmon

clean:
	rm -rf bin/
	rm -f netmon

security:
	govulncheck ./...
	gosec ./...
	trufflehog git file://. --only-verified --fail
	gitleaks detect --source . -v

security-full:
	gitleaks detect --source . -v --log-opts="--all"
	trufflehog git file://. --no-verification
