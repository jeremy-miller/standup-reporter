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
	@pip3 install pre-commit
	@pre-commit install -c githooks/.pre-commit-config.yaml -t pre-commit
	@pre-commit install -c githooks/.pre-commit-config.yaml -t pre-push

.PHONY: setup-ci
setup-ci: ## Setup CI/CD environment
	go mod download
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
	curl https://pre-commit.com/install-local.py | python -

.PHONY: build
build: clean ## Build the standup-reporter executables and place them in local build/ directory
	@build/build.sh

.PHONY: check
check: ## Try building all packages without producing binaries (i.e. check for errors)
	@go build $(PKG_LIST)

.PHONY: modd
modd: ## Run modd
	@./modd --file=config/modd.conf

.PHONY: lint
lint: ## Lint files
	@golangci-lint run --config config/.golangci.yml ./...

.PHONY: lint-ci
lint-ci: ## Lint files during CI/CD
	git diff-tree --no-commit-id --name-only -r $(TRAVIS_COMMIT) | xargs pre-commit run -c githooks/.pre-commit-config.yaml --files
	golangci-lint run --config config/.golangci.yml ./...

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
