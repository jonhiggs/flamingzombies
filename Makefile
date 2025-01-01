SHELL := /usr/bin/env bash

build clean fz.tar.bz2 install:
	$(MAKE) -f dist.mk $@

test: gotest shellcheck bats

gotest govet: dirs = $(shell find . -name \*_test.go | xargs -I{} dirname {})
gotest:
	go test $(dirs)

govet:
	go vet $(dirs)

shellcheck: files := $(shell find libexec -type f -not -name \*.bats -not -name \*.md)
shellcheck:
	shellcheck -e SC1091 -x -s sh $(files)

bats: files = $(shell find . -iname \*.bats -not -path ./dist\*)
bats:
	bats $(files)
