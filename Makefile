MAKEFLAGS += -j6

build: preprocess
	go build

preprocess: build-assets build-templ

preprocess-watch: build-assets-watch build-templ-watch

build-assets:
	node ./build.mjs

build-assets-watch:
	node ./build.mjs --watch

build-templ:
	templ generate

build-templ-watch:
	templ generate -watch
