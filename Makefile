include .env
export

.PHONY: run

run:
	@go run cmd/stig/main.go
