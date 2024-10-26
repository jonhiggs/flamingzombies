SHELL := /usr/bin/env bash

build clean:
	$(MAKE) -f dist.mk $@

test: dirs = $(shell find . -name \*_test.go | xargs -I{} dirname {})
test: shellcheck
	go test $(dirs)

shellcheck:
	shellcheck -e SC1091 -x -s sh \
		 libexec/helpers.inc libexec/{task,notifier,gate}/*
	#[[ $$(find libexec/ ! -executable ! -name README.md ! -name \*.inc) = "" ]]
