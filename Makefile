# Nutty's build, test and release recipes.
#
# Targets:
#   make            build (alias of build)
#   make run        run the app from source
#   make fmt        gofmt every source file
#   make vet        go vet
#   make lint       golangci-lint over the whole module
#   make test       run the tests
#   make snapshot   render preview images to /tmp
#   make icon       regenerate resources/icon.png from the svg
#   make package    build Nutty.app and the DMG into out/
#   make clean      remove build outputs

.PHONY: run build fmt vet lint test snapshot icon package clean

# GOLANGCI is the linter, wherever it lives: the GOPATH copy when the
# PATH one is missing.
GOLANGCI := $(shell command -v golangci-lint 2>/dev/null || echo $(shell go env GOPATH)/bin/golangci-lint)

run:
	go run ./cmd/nutty

build:
	go build ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

# Lint runs golangci-lint over the whole module. Install it with
# `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`
# if the target cannot find it.
lint:
	$(GOLANGCI) run ./...

test:
	go test ./...

# TestSnapshot renders the window to /tmp/nutty-preview.png, the site
# shots to /tmp/nutty-site-*.png, and the files panel to /tmp/nutty-files.png.
snapshot:
	NUTTY_SNAPSHOT=1 go test -run 'TestSnapshot|TestSiteShots|TestFilesPanel' -v ./internal/app/

# Icon regenerates resources/icon.png, which MyGo packages into the app
# bundle. Run it after changing internal/app/assets/nutty.svg.
icon:
	go run ./tools/icon

# Package builds Nutty.app and a DMG under out/ with MyGo's own build,
# as the release workflow does. The app is ad hoc signed: mygo.json names
# no signing identity, and nothing is notarized.
package:
	go run github.com/egoist/mygo/cmd/mygo@v0.4.2 build -platform darwin/universal -o out

clean:
	rm -rf out dist
