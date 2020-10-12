.PHONY: all help start stop unit

all: unit

help:
	@echo "Please use \`make <target>' where <target> is one of"
	@echo "  start                   start the unit test database"
	@echo "  stop                    stop the unit test database"
	@echo "  unit                    to run unit tests"

start:
	cd docker; docker-compose up --no-recreate -d

stop:
	cd docker; docker-compose down

unit:
	go mod tidy
	go test -cover -p 1 ./...
