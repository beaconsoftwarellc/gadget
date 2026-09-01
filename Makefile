mocks: ## Generate mocks using go generate
	go generate ./...

lint: ## Run linter
	golangci-lint run -v

format: ## Format Go code
	gofmt -w .

unit: format ## Run unit tests
	env go mod tidy
	go test -cover -p 1 ./...
