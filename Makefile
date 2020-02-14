.PHONY: all help unit

all: unit

help:
	@echo "Please use \`make <target>' where <target> is one of"
	@echo "  unit                    to run unit tests"

unit:
	go mod tidy
	go test -cover -p 1 ./...
