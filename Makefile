SHELL := /bin/bash

ifeq ($(findstring release,$(MAKECMDGOALS)),release)
  ifndef MESSAGE
    $(error MESSAGE was not provided)
  endif
endif

VERSION = $(shell cat cmd/fz/fz.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')
artifacts := $(addsuffix .tar.gz, dist/fz_openbsd_arm64 dist/fz_openbsd_amd64 dist/fz_linux_arm64 dist/fz_linux_amd64 dist/fz_darwin_arm64 dist/fz_darwin_amd64)

release: $(artifacts)
	gh release create $(VERSION) --notes "${MESSAGE}"
	gh release upload $(VERSION) $(artifacts)

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
