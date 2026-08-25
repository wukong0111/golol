.PHONY: run test tidy

ADDR ?= :8080

run:
	ADDR=$(ADDR) go run ./cmd/golol

test:
	go test ./...

tidy:
	go mod tidy
