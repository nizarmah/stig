include .env
export

.PHONY: run record

run:
	@go run cmd/stig/main.go

record:
	@go run cmd/recorder/main.go
