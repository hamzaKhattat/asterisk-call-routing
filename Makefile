.PHONY: build run clean test

build:
	go build -o bin/router cmd/router/main.go

run: build
	./bin/router

clean:
	rm -rf bin/

test:
	go test -v ./...

deps:
	go mod download
	go mod tidy
