SHELL := /usr/bin/env bash

build clean fz.tar.bz2 install:
	$(MAKE) -f dist.mk $@

test: gotest shellcheck

gotest govet: dirs = $(shell find . -name \*_test.go | xargs -I{} dirname {})
gotest:
	go test $(dirs)

govet:
	go vet $(dirs)

shellcheck:
	shellcheck -e SC1091 -x -s sh \
		 libexec/{task,notifier,gate}/*
	#[[ $$(find libexec/ ! -executable ! -name README.md ! -name \*.inc) = "" ]]
