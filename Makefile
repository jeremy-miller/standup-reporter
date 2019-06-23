PKG_LIST := $(shell go list ./...)

build:
	@go install -i $(PKG_LIST)

run:
	@go run $(PKG_LIST)

clean:
	@go clean $(PKG_LIST)
