-include .env
export

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: dev
dev:
	@wails dev -skipbindings -noreload

.PHONY: generate
generate:
	@wails generate module

.PHONY: build
build:
	@wails build -skipbindings -clean -o Multibase

.PHONY: release
release:
	@wails build -skipbindings -clean -o Multibase
	@envsubst < ./build/darwin/gon-sign.template.json > ./build/darwin/gon-sign.json
	@envsubst < ./build/darwin/gon-notarize.template.json > ./build/darwin/gon-notarize.json
	@gon -log-level=info ./build/darwin/gon-sign.json
	@npx create-dmg ./build/bin/Multibase.app --dmg-title=Multibase --overwrite ./build/bin
	@mv ./build/bin/Multibase*.dmg ./build/bin/multibase.dmg
	@gon -log-level=info ./build/darwin/gon-notarize.json

.PHONY: tag
tag:
	$(eval VERSION = ${v})
	@git tag ${VERSION}
	@git push origin ${VERSION}

.PHONY: update-wails
update-wails:
	@go install github.com/wailsapp/wails/v2/cmd/wails@latest

.PHONY: cleanup-state
cleanup-state:
	@rm -rf ~/Library/Application\ Support/multibase/state
