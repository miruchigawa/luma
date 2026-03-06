.PHONY: run build test lint tidy clean

run:
	go run ./cmd/bot

dev:
	air

build:
	go build -o bin/bot ./cmd/bot

test:
	go test ./... -v -cover

lint:
	golangci-lint run

tidy:
	go mod tidy

clean:
	rm -rf bin/
