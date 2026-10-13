# Nutty's build, test and release recipes.
#
# Targets:
#   make            build (alias of build)
#   make run        run the app from source
#   make test       run the tests
#   make snapshot   render preview images to /tmp
#   make icon       regenerate resources/icon.png from the svg
#   make package    build Nutty.app and the DMG into out/
#   make clean      remove build outputs

.PHONY: run test snapshot icon package vet clean

run:
	go run ./cmd/nutty

build:
	go build ./...

vet:
	go vet ./...

test:
	go test ./...

# TestSnapshot renders the window to /tmp/nutty-preview.png, the site
# shots to /tmp/nutty-site-*.png, and the files panel to /tmp/nutty-files.png.
snapshot:
	NUTTY_SNAPSHOT=1 go test -run 'TestSnapshot|TestSiteShots|TestFilesPanel' -v ./internal/app/

# Icon regenerates resources/icon.png, which MyGo packages into the app
# bundle. Run it after changing assets/nutty.svg.
icon:
	go run ./tools/icon

# Package builds a signed Nutty.app and a DMG under out/ with MyGo's
# own build, as the release workflow does.
package:
	go run github.com/egoist/mygo/cmd/mygo@v0.4.2 build -platform darwin/universal -o out

clean:
	rm -rf out dist
