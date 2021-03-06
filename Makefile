PKG_LIST := $(shell go list ./...)

default: help

.PHONY: help
help: ## Show this help
	@echo
	@echo "Available commands:"
	@echo
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF}' $(MAKEFILE_LIST)

.PHONY: setup
setup: ## Setup development environment
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@go get golang.org/x/tools/cmd/goimports
	@pip3 install pre-commit
	@pre-commit install -c githooks/.pre-commit-config.yaml -t pre-commit
	@pre-commit install -c githooks/.pre-commit-config.yaml -t pre-push
	@pre-commit install -c githooks/.pre-commit-config.yaml -t commit-msg
	@git config --local commit.template config/.gitmessage

.PHONY: build
build: clean ## Build the standup-reporter executable for the current OS and place it in local bin/ directory
	@go build -o bin/standup-reporter github.com/jeremy-miller/standup-reporter/cmd/standup-reporter

.PHONY: check
check: ## Try building all packages without producing binaries (i.e. check for errors)
	@go build $(PKG_LIST)

.PHONY: modd
modd: ## Run modd
	@./modd --file=config/modd.conf

.PHONY: lint
lint: ## Lint files
	@golangci-lint run --config config/.golangci.yml ./...

.PHONY: test
test: ## Run all tests with data race detection
	@go test -v -race $(PKG_LIST)

.PHONY: coverage
coverage: ## Run all tests with data race detection and generate code coverage
	@go test -v -race $(PKG_LIST) -coverprofile .testCoverage.txt
	@go tool cover -func=.testCoverage.txt

.PHONY: run
run: build ## Build and run the standup-reporter; assumes ASANA_TOKEN env var exists
	@bin/standup-reporter --asana=$(ASANA_TOKEN)

.PHONY: update-deps
update-deps: ## Update dependencies
	@go get -u ./...

.PHONY: tidy
tidy: ## Remove unused dependencies
	@go mod tidy

.PHONY: clean
clean: ## Remove generated/compiled files
	@go clean $(PKG_LIST)
	@rm -rf bin

.PHONY: update-hooks
update-hooks: ## Update pre-commit hook versions
	@pre-commit autoupdate -c githooks/.pre-commit-config.yaml
