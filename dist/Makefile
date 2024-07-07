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

build: $(BINS)

$(BINS): BIN = $(word 1,$(subst _, ,$@))
$(BINS): GOOS = $(word 2,$(subst _, ,$@))
$(BINS): GOARCH = $(word 3,$(subst _, ,$@))
$(BINS): TARGET_DIR = $(GOOS)_$(GOARCH)
$(BINS):
	mkdir -p $(TARGET_DIR)/bin $(TARGET_DIR)/libexec
	docker run                                                        \
		--volume "$(shell pwd)/../:/go"                               \
		--env "CGO_ENABLED=0"                                         \
		--env "GOOS=$(GOOS)"                                          \
		--env "GOARCH=$(GOARCH)"                                      \
		golang:1.22-bookworm                                          \
		bash -c '                                                     \
			go build -o ./dist/$(TARGET_DIR)/$(BIN) -buildvcs=false ./cmd/$(BIN) \
				&& chown $(UID):$(GID) ./dist/$(TARGET_DIR)/$(BIN)    \
		'

.PHONY: clean
clean:
	rm -rf linux_amd64 openbsd_amd64
