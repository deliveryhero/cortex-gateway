BINARY_NAME=cortex-gateway

build:
	go build -o ${BINARY_NAME} main.go

run:
	go run main.go

test:
	go test -v ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run
