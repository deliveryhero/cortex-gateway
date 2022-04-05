BINARY_NAME=cortex-gateway

build:
	go build -o ${BINARY_NAME} main.go

run:
	go run main.go
