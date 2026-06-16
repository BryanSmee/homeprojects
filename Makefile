.PHONY: run build test vet fmt tidy docker

run:
	go run ./cmd/server

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w cmd internal

tidy:
	go mod tidy

docker:
	docker build -t homeprojects:latest .
