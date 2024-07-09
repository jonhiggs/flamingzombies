SHELL = /bin/bash
UID = $(shell id -u)
GID = $(shell id -g)

BINS = fz_linux_amd64              \
	   fzctl_linux_amd64           \
	   fz_openbsd_amd64            \
	   fzctl_openbsd_amd64         \
	   task/ping_linux_amd64       \
	   task/ping_openbsd_amd64     \
	   task/diskfree_linux_amd64   \
	   task/diskfree_openbsd_amd64

CMDS = 	fz            \
		fzctl         \
		task/diskfree \
		task/ping     \
		task/swapfree

BINS = $(addprefix openbsd_amd64/,$(CMDS)) $(addprefix linux_amd64/,$(CMDS))

build: $(BINS)

$(BINS): TYPE = $(word 1,$(subst /, ,$(dir $@)))
$(BINS): BIN = $(subst $(TYPE)/,,$@)
$(BINS): GOOS = $(word 1,$(subst _, ,$(TYPE)))
$(BINS): GOARCH = $(word 2,$(subst _, ,$(TYPE)))
$(BINS):
	mkdir -p $(dir $@)
	docker run                                                        \
		--volume "$(shell pwd)/../:/go"                               \
		--env "CGO_ENABLED=1"                                         \
		--env "GOOS=$(GOOS)"                                          \
		--env "GOARCH=$(GOARCH)"                                      \
		-it \
		golang:1.22-bookworm                                          \
		bash -c '                                                     \
			bash; \
			go build -o ./dist/$@ -buildvcs=false ./cmd/$(BIN) \
				&& chown $(UID):$(GID) ./dist/$@    \
		'

.PHONY: clean
clean:
	rm -rf linux_amd64 openbsd_amd64
