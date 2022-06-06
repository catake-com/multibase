.PHONY: lint
lint:
	@golangci-lint run

.PHONY: dev
dev:
	@wails dev

.PHONY: generate
generate:
	@wails generate module

.PHONY: update-wails
update-wails:
	@go install github.com/wailsapp/wails/v2/cmd/wails@latest

.PHONY: cleanup-state
cleanup-state:
	@rm ~/Library/Application\ Support/multibase/project.json
	@rm ~/Library/Application\ Support/multibase/grpc.json
