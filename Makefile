SHELL := /usr/bin/env bash

FZ_VERSION := $(shell cat cmd/fz/fz.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')
FZCTL_VERSION := $(shell cat cmd/fzctl/fzctl.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')
VERSION := $(FZ_VERSION)

BINS = $(addprefix dist/,fz_linux_amd64 fzctl_linux_amd64 fz_openbsd_amd64 fzctl_openbsd_amd64)
TASKS = dist/task/ping

CMDS = $(addprefix dist/,fzctl fz task/diskfree task/ping task/swapfree)

build: $(CMDS)
$(CMDS): src = ./cmd/$(subst dist/,,$@)
$(CMDS): .FORCE
	mkdir -p $(dir $@)
	go build -o $@ $(src)


release: prerelease_tests release_notes.txt $(BINS) dist/plugins.tar.gz
	gh release create $(VERSION) -F release_notes.txt
	gh release upload $(VERSION) dist/fz_* dist/fzctl_* dist/plugins.tar.gz

devrelease: gitsha := $(shell git rev-parse HEAD)
devrelease: clean test $(BINS) dist/plugins.tar.gz
	scp -r man/ janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/
	scp $(BINS) dist/plugins.tar.gz janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/
	scp scripts/openbsd_rc janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/
	scp scripts/openrc janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/

devplugins: clean shellcheck dist/plugins.tar.gz
	scp dist/plugins.tar.gz janx:/var/www/htdocs/artifacts.altos/flamingzombies/dev/

$(BINS) dist/plugins.tar.gz:
	make -C dist $(notdir $@)

release_notes.txt: CHANGELOG.md
	sed -n '/^## $(VERSION)$$/,/##/ { /^#/d; /^\w*$$/d; p }' $< > $@

prerelease_tests: test
	[[ "$(FZCTL_VERSION)" == "$(FZ_VERSION)" ]]
	git status | grep -q "nothing to commit"
	git push
	git fetch --tags
	! git rev-parse $(VERSION) &>/dev/null
	git status | grep -q "On branch main"
	git status | grep -q "working tree clean"
	grep -q "^## $(VERSION)$$" CHANGELOG.md
	$(MAKE) -C ./man test

test: gotest shellcheck

gotest:
	go test ./lib/fz

shellcheck:
	shellcheck -e SC1091 -x -s sh \
		 libexec/helpers.inc libexec/{task,notifier,gate}/*
	#[[ $$(find libexec/ ! -executable ! -name README.md ! -name \*.inc) = "" ]]

clean:
	$(MAKE) -C dist clean
	rm -f release_notes.txt

.FORCE:
