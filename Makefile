.PHONY: build test fmt lint run clean security

build:
	go build -o bin/netmon ./cmd/netmon

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

security:
	govulncheck ./...
	gosec ./...
	trufflehog git file://. --only-verified --fail
	gitleaks detect --source . -v
