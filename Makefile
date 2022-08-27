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
	@rm ~/Library/Application\ Support/multibase/grpc
	@rm ~/Library/Application\ Support/multibase/thrift
	@rm ~/Library/Application\ Support/multibase/kafka

.PHONY: show-state
show-state:
	@jq . ~/Library/Application\ Support/multibase/project.json
	@jq . ~/Library/Application\ Support/multibase/grpc
	@jq . ~/Library/Application\ Support/multibase/thrift
	@jq . ~/Library/Application\ Support/multibase/kafka

.PHONY: release
release:
	$(eval VERSION = ${v})
	@git tag ${VERSION}
	@git push origin ${VERSION}
