PKG_LIST := $(shell go list ./...)

default: help

.PHONY: help
help: ## Show this help
	@echo
	@echo "Choose a command run:"
	@echo
	@awk -F ':|##' '/^[^\t].+?:.*?##/ {printf "\033[36m%-30s\033[0m %s\n", $$1, $$NF}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the standup-reporter executable
	@go install -i $(PKG_LIST)

.PHONY: run
run: ## Run the standup-reporter
	@go run $(PKG_LIST)

.PHONY: clean
clean: ## Remove compiled executable
	@go clean $(PKG_LIST)
