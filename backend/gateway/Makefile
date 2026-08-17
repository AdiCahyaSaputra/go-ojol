# Import .env file
ifneq (,$(wildcard ./.env))
		include .env
		export $(shell sed 's/=.*//' .env)
endif

# Variables
CONTAINER_NAME=${APP_NAME}

# Commands
dep:
	go mod tidy

run:
	go run cmd/main.go

build:
	go build -o main cmd/main.go

run-build: build
	./main

test:
	go test -v ./...

module:
	@if [ -z "$(name)" ]; then echo "Usage: make module name=<module_name>"; exit 1; fi
	@./create_module.sh $(name)
