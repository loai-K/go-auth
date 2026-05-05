.PHONY: build test run docker

MODULE := ./...

build:
	go build $(MODULE)

test:
	go test ./...

run:
	go run ./cmd/auth-server

docker:
	docker build -t goauth:mvp .
