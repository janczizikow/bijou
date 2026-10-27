build:
	go build -o=./bin/bijou ./cmd/bijou

vendor:
	@go mod tidy
	@go mod verify
	@go mod vendor

.PHONY: test
test:
	go test -race -cover ./...
