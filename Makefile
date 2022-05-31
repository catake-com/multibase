.PHONY: lint
lint:
	@golangci-lint run

.PHONY: dev
dev:
	@wails dev

.PHONY: generate
generate:
	@wails generate module
