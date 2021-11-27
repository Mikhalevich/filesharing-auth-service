all: build

build:
	go build -mod=vendor -o ./bin/auth cmd/auth/main.go

vendor:
	go mod tidy
	go mod vendor

