.PHONY: build test fmt lint run clean

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
