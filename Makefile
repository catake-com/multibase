.PHONY: lint
lint:
	@golangci-lint run

.PHONY: dev
dev:
	@wails dev -skipbindings -noreload

.PHONY: generate
generate:
	@wails generate module

.PHONY: update-wails
update-wails:
	@go install github.com/wailsapp/wails/v2/cmd/wails@latest

.PHONY: cleanup-state
cleanup-state:
	@rm -rf ~/Library/Application\ Support/multibase/state

.PHONY: release
release:
	$(eval VERSION = ${v})
	@git tag ${VERSION}
	@git push origin ${VERSION}
