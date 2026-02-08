.PHONY: build build-cli build-web test test-verbose test-cover vet clean

build: build-cli build-web

build-cli:
	go build -o bin/adr ./cmd/adr-cli

build-web:
	go build -o bin/adr-web ./cmd/adr-web

test:
	go test ./...

test-verbose:
	go test -v ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

vet:
	go vet ./...

clean:
	rm -rf bin/ coverage.*
