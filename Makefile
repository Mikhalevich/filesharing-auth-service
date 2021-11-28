all: build

.PHONY: build
build:
	go build -mod=vendor -o ./bin/auth cmd/auth/main.go

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

