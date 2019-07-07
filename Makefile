default: help

.PHONY: help
help: ## Show this help
	@echo
	@echo "Available commands:"
	@echo
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF}' $(MAKEFILE_LIST)

.PHONY: build
build: clean ## Build the standup-reporter executables and place them in local build/ directory
	@scripts/build.sh

.PHONY: update-deps
update-deps: ## Update dependencies
	@go get -u ./...

.PHONY: remove-deps
remove-deps: ## Remove unused dependencies
	@go mod tidy

.PHONY: clean
clean: ## Remove generated/compiled files
	@go clean
	@rm -rf build

.PHONY: update-hooks
update-hooks: ## Update pre-commit hook versions
	@pre-commit autoupdate -c githooks/.pre-commit-config.yaml
