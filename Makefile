
.PHONY: test build

test:
	@go test -v ./...

build:
	@docker build -t thomasnyambati/drone-pr-checker .

run:
	@go run main.go

lint:
	@golangci-lint run
