SHELL := /bin/bash

VERSION = $(shell cat cmd/fz/fz.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')

artifacts := dist/linux/amd64/flamingzombies-$(VERSION).tar.gz \
             dist/openbsd/amd64/flamingzombies-$(VERSION).tar.gz

gitsha := $(shell git rev-parse HEAD)

release: prerelease_tests release_notes.txt $(artifacts)
	gh release create $(VERSION) -F release_notes.txt
	cp dist/linux/amd64/flamingzombies-$(VERSION).tar.gz dist/flamingzombies-$(VERSION)-linux-amd64.tar.gz
	cp dist/openbsd/amd64/flamingzombies-$(VERSION).tar.gz dist/flamingzombies-$(VERSION)-openbsd-amd64.tar.gz
	gh release upload $(VERSION) dist/flamingzombies-$(VERSION)-*.tar.gz

release_notes.txt: CHANGELOG.md
	sed -n '/^## $(VERSION)$$/,/##/ { /^#/d; /^\w*$$/d; p }' $< > $@

dist/%.tar.gz: dist/%/bin/fz dist/man/man1/fz.1.gz
	mkdir -p dist/$*/share
	mkdir -p dist/etc
	cp -r dist/man dist/$*/share
	cp -r libexec/ dist/$*
	touch dist/etc/flamingzombies.sample.toml
	tar -zvc \
		-C $(dir dist/$*) \
		-f $@ \
		flamingzombies-$(VERSION)/

dist/linux/amd64/flamingzombies-%/bin/fz:
	mkdir -p $$(dirname $@)
	go build ./cmd/fz/fz.go
	mv fz $@

dist/openbsd/amd64/flamingzombies-%/bin/fz:
	git push
	mkdir -p $$(dirname $@)
	ssh janx build_flamingzombies/build $(gitsha)
	wget http://artifacts.altos/flamingzombies/openbsd/fz-$*-amd64 -O $@

dist/man/%.gz: export BUILD_DATE = $(shell date --iso-8601)
dist/man/%.gz: man/% | dist/man/man1
	cat $< | envsubst '$${BUILD_DATE}' > dist/man/$*
	gzip -f dist/man/$*

dist/man/man1 doc/man1:
	mkdir -p $@

prerelease_tests: test
	git fetch --tags
	! git rev-parse $(VERSION) &>/dev/null
	git status | grep -q "On branch master"
	git status | grep -q "working tree clean"

test: gotest shellcheck

gotest:
	go test ./lib/fz

shellcheck:
	shellcheck -s sh libexec/{task,notifier,gates}/*

clean:
	rm -Rf ./dist
