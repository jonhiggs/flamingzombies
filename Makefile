SHELL := /bin/bash

VERSION = $(shell cat cmd/fz/fz.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')
artifacts := $(addsuffix .tar.gz, dist/fz_openbsd_arm64 dist/fz_openbsd_amd64 dist/fz_linux_arm64 dist/fz_linux_amd64 dist/fz_darwin_arm64 dist/fz_darwin_amd64)

release: release_notes.txt $(artifacts)
		| gh release create $(VERSION) -F release_notes.txt
	gh release upload $(VERSION) $(artifacts)

release_notes.txt: CHANGELOG.md
	sed -n '/^## $(VERSION)$$/,/##/ { /^#/d; /^\w*$$/d; p }' $< > $@

dist/fz_darwin_amd64.tar.gz:  DIR := dist/fz_darwin_amd64_v1
dist/fz_darwin_arm64.tar.gz:  DIR := dist/fz_darwin_arm64
dist/fz_linux_amd64.tar.gz:   DIR := dist/fz_linux_amd64_v1
dist/fz_linux_arm64.tar.gz:   DIR := dist/fz_linux_arm64
dist/fz_openbsd_amd64.tar.gz: DIR := dist/fz_openbsd_amd64_v1
dist/fz_openbsd_arm64.tar.gz: DIR := dist/fz_openbsd_arm64
dist/%.tar.gz: gorelease_build
	mkdir -p $(DIR)/bin
	mkdir -p $(DIR)/share/man1
	mkdir -p $(DIR)/libexec/flamingzombies
	mv $(DIR)/fz $(DIR)/bin/fz
	cp -r libexec/* $(DIR)/libexec/flamingzombies
	tar -C dist -zcvf $@ $(notdir $(DIR))

dist/plugins.tar.bz:
	tar xvf $@ plugins/

gorelease_build: test
	git status status 2>&1 | grep -q "working tree clean"
	git branch | grep -q "* master"
	grep -q '^## $(VERSION)$$' CHANGELOG.md
	git tag -a $(VERSION) -m "$(MESSAGE)"
	git push origin $(VERSION)
	goreleaser build --clean
	goreleaser build --snapshot --clean

test: gotest shellcheck

gotest:
	go test ./lib/fz

shellcheck:
	shellcheck -s sh libexec/{task,notifier,gates}/*

clean:
	rm -Rf ./dist
