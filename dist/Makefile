SHELL = /usr/bin/env bash
UID = $(shell id -u)
GID = $(shell id -g)

GOOS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH = $(shell uname -m | sed 's/x86_64/amd64/')

ifeq ($(GOOS),linux)
  LDD_VERSION = $(shell ldd --version 2>&1 | grep -q musl && echo musl || echo gnu)
  DIR = $(GOOS)_$(GOARCH)_$(LDD_VERSION)
endif

BINS = $(DIR)/fz             \
	$(DIR)/fzctl         \
	$(DIR)/task/ping     \
	$(DIR)/task/diskfree \
	$(DIR)/task/swapfree \

build: $(BINS)

publish: DEST := git@janx.altos:/var/www/htdocs/artifacts.altos/flamingzombies/$(shell git rev-parse HEAD)
publish: $(BINS)
	rsync -av $(DIR) $(DEST)

$(BINS): export CGO_ENABLED = 1
$(BINS):
	mkdir -p $(dir $@)
	go build -o $@ -buildvcs=false ../cmd/$(subst $(DIR)/,,$@)

.PHONY: clean
clean:
	rm -rf $(DIR)
