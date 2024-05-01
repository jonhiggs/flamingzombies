SHELL := /bin/bash

FZ_VERSION := $(shell cat cmd/fz/fz.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')
FZCTL_VERSION := $(shell cat cmd/fzctl/fzctl.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')
VERSION := $(FZ_VERSION)

release: prerelease_tests release_notes.txt dist/fz_openbsd_amd64 dist/fz_linux_amd64 dist/fzctl_openbsd_amd64 dist/fzctl_linux_amd64 dist/plugins.tar.gz
	gh release create $(VERSION) -F release_notes.txt
	gh release upload $(VERSION) dist/fz_* dist/fzctl_* dist/plugins.tar.gz

devrelease: gitsha := $(shell git rev-parse HEAD)
devrelease: clean predevrelease_tests dist/fz_linux_amd64 dist/fzctl_linux_amd64 dist/plugins.tar.gz
	ssh janx build_flamingzombies/build_dev $(gitsha)
	scp dist/fz_* janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/
	scp dist/fzctl_* janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/
	scp dist/plugins.tar.gz janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/
	scp scripts/openbsd_rc janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/

release_notes.txt: CHANGELOG.md
	sed -n '/^## $(VERSION)$$/,/##/ { /^#/d; /^\w*$$/d; p }' $< > $@

dist/plugins.tar.gz: libexec | dist
	mkdir -p dist/flamingzombies
	cp -aux $</* dist/flamingzombies
	tar -C ./dist -zcvf $@ flamingzombies
	rm -Rf dist/flamingzombies

dist/fz_linux_amd64: | dist
	mkdir -p $$(dirname $@)
	CGO_ENABLED=0 go build ./cmd/fz/fz.go
	mv fz $@

dist/fzctl_linux_amd64: | dist
	mkdir -p $$(dirname $@)
	CGO_ENABLED=0 go build ./cmd/fzctl/fzctl.go
	mv fzctl $@

dist/fz_openbsd_amd64: gitsha := $(shell git rev-parse HEAD)
dist/fz_openbsd_amd64: prerelease_tests | dist
	ssh janx build_flamingzombies/build $(gitsha)
	wget http://artifacts.altos/flamingzombies/openbsd/$(VERSION)/fz \
		-O $@
	chmod 755 $@

dist/fzctl_openbsd_amd64: gitsha := $(shell git rev-parse HEAD)
dist/fzctl_openbsd_amd64: prerelease_tests | dist
	ssh janx build_flamingzombies/build $(gitsha)
	wget http://artifacts.altos/flamingzombies/openbsd/$(VERSION)/fzctl \
		-O $@
	chmod 755 $@

dist:
	mkdir -p $@

prerelease_tests: test
	[[ "$(FZCTL_VERSION)" == "$(FZ_VERSION)" ]]
	git status | grep -q "nothing to commit"
	git push
	git fetch --tags
	! git rev-parse $(VERSION) &>/dev/null
	git status | grep -q "On branch master"
	git status | grep -q "working tree clean"
	grep -q "^## $(VERSION)$$" CHANGELOG.md
	./man/test.sh

predevrelease_tests: test
	git status | grep -q "nothing to commit"
	git push
	git status | grep -q "working tree clean"

test: gotest shellcheck

gotest:
	go test ./lib/fz

shellcheck:
	shellcheck -s sh libexec/{task,notifier,gate}/*
	[[ $$(find libexec/ ! -executable ! -name README.md) = "" ]]

clean:
	rm -Rf ./dist release_notes.txt
