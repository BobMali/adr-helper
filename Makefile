.PHONY: build build-cli build-web build-frontend test test-verbose test-cover vet clean

build: build-cli build-web

build-cli:
	go build -o bin/adr ./cmd/adr-cli

build-frontend:
	cd web && npm ci && npm run build

build-web: build-frontend
	go build -tags embed -o bin/adr-web ./cmd/adr-web

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
