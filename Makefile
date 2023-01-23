# Include variables from the .envrc file
include .envrc

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run/service: run the cmd/service command to start service in background
.PHONY: run/service
run/service:
	docker-compose up --build

## run/stop: run the cmd/stop command to stop service
run/stop: 
	docker-compose down

## run/example: run the cmd/example command to submit a log
.PHONY: run/example
run/example:
	go run ./cmd/example/

## test: run the test suite
.PHONY: test
test:
	go test -race ./... -v