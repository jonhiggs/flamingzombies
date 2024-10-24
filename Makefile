SHELL := /usr/bin/env bash

CMDS = dist/fzctl dist/fz

GO_TASKS = $(addprefix dist/libexec/task/,diskfree ping swapfree)
SH_TASKS = $(addprefix dist/,$(wildcard libexec/task/*))

SH_GATES = $(addprefix dist/,$(wildcard libexec/gate/*))

SH_NOTIFIERS = $(addprefix dist/,$(wildcard libexec/notifier/*))

MAN1_PAGES = $(addprefix dist/,$(wildcard man/man1/*.1))
MAN5_PAGES = $(addprefix dist/,$(wildcard man/man5/*.5))
MAN7_PAGES = $(addprefix dist/,$(wildcard man/man7/*.7))
MAN_PAGES = $(MAN1_PAGES) $(MAN5_PAGES) $(MAN7_PAGES)

build: $(CMDS) $(GO_TASKS) $(SH_TASKS) $(SH_GATES) $(SH_NOTIFIERS) $(MAN_PAGES)

$(CMDS): src = ./cmd/$(subst dist/,,$@)
$(CMDS): .FORCE | dist
	go build -o $@ ./cmd/$(notdir $@)

$(GO_TASKS):
	go build -o $@ ./cmd/task/$(notdir $@)

$(SH_TASKS): | dist/libexec/task
	cp ./libexec/task/$(notdir $@) $@

$(SH_GATES): | dist/libexec/gate
	cp ./libexec/gate/$(notdir $@) $@

$(SH_NOTIFIERS): | dist/libexec/notifier
	cp ./libexec/notifier/$(notdir $@) $@

$(MAN1_PAGES): src = ./man/man1/$(notdir $@)
$(MAN5_PAGES): src = ./man/man5/$(notdir $@)
$(MAN7_PAGES): src = ./man/man7/$(notdir $@)
$(MAN_PAGES): file_ts = $(shell date -r $$(git log -1 --pretty="format:%ct" $(src)) +%Y-%m-%d)
$(MAN_PAGES): content_ts = $(shell awk '/.Dd/ { print $$2 }' $(src))
$(MAN_PAGES): | dist/man/man1 dist/man/man5 dist/man/man7
	echo testing timestamp of $@
	[[ $(file_ts) = $(content_ts) ]]
	cp $(src) $@

dist dist/libexec/task dist/libexec/gate dist/libexec/notifier dist/man/man1 dist/man/man5 dist/man/man7:
	mkdir -p $@

test: dirs = $(shell find . -name \*_test.go | xargs -I{} dirname {})
test: shellcheck
	go test $(dirs)

shellcheck:
	shellcheck -e SC1091 -x -s sh \
		 libexec/helpers.inc libexec/{task,notifier,gate}/*
	#[[ $$(find libexec/ ! -executable ! -name README.md ! -name \*.inc) = "" ]]

clean:
	rm -Rf ./dist/*

.FORCE:
