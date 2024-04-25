SHELL := /bin/bash

VERSION = $(shell cat cmd/fz/fz.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')

release: prerelease_tests release_notes.txt dist/fz_openbsd_amd64 dist/fz_linux_amd64 dist/plugins.tar.gz
	gh release create $(VERSION) -F release_notes.txt
	gh release upload $(VERSION) dist/fz_* dist/plugins.tar.gz

devrelease: clean predevrelease_tests dist/fz_linux_amd64 dist/plugins.tar.gz
	ssh janx build_flamingzombies/build $(gitsha)
	scp dist/fz_* janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/
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

dist/fz_openbsd_amd64: gitsha := $(shell git rev-parse HEAD)
dist/fz_openbsd_amd64: prerelease_tests | dist
	ssh janx build_flamingzombies/build $(gitsha)
	wget http://artifacts.altos/flamingzombies/openbsd/$(VERSION)/fz \
		-O $@
	chmod 755 $@

dist/man/%.gz: export BUILD_DATE = $(shell date --iso-8601)
dist/man/%.gz: man/% | dist/man/man1
	cat $< | envsubst '$${BUILD_DATE}' > dist/man/$*
	gzip -f dist/man/$*

dist/man/man1 doc/man1 dist:
	mkdir -p $@

prerelease_tests: test
	git status | grep -q "nothing to commit"
	git push
	git fetch --tags
	! git rev-parse $(VERSION) &>/dev/null
	git status | grep -q "On branch master"
	git status | grep -q "working tree clean"
	grep -q "^## $(VERSION)$$" CHANGELOG.md

predevrelease_tests: test
	git status | grep -q "nothing to commit"
	git push
	git status | grep -q "working tree clean"

test: gotest shellcheck

gotest:
	go test ./lib/fz

shellcheck:
	shellcheck -s sh libexec/{task,notifier,gates}/*

clean:
	rm -Rf ./dist release_notes.txt
