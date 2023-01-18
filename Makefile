# Include variables from the .envrc file
include .envrc

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run/service: run the cmd/service application
.PHONY: run/service
run/service:
	docker-compose up --build

## run/example: run the cmd/example application to submit a log
.PHONY: run/example
run/example:
	go run ./cmd/example/

## test: run the test suite
.PHONY: test
test:
	go test ./... -v