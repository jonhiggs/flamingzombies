SHELL = /bin/bash
UID = $(shell id -u)
GID = $(shell id -g)

BINS = fz_linux_amd64 fzctl_linux_amd64 fz_openbsd_amd64 fzctl_openbsd_amd64

build: $(BINS) plugins.tar.gz

$(BINS): BIN = $(word 1,$(subst _, ,$@))
$(BINS): GOOS = $(word 2,$(subst _, ,$@))
$(BINS): GOARCH = $(word 3,$(subst _, ,$@))
$(BINS):
	docker run                                            \
		--volume "$(shell pwd)/../:/go"                   \
		--env "CGO_ENABLED=0"                             \
		--env "GOOS=$(GOOS)"                              \
		--env "GOARCH=$(GOARCH)"                          \
		golang:1.22-bookworm                              \
		bash -c '                                         \
			go build -o ./dist/$@ ./cmd/$(BIN)/$(BIN)*.go \
				&& chown $(UID):$(GID) ./dist/$@          \
		'

plugins.tar.gz:
	rm -f $@
	mkdir -p flamingzombies
	cp -aux ../libexec/* flamingzombies
	tar -zcvf $@ flamingzombies
	rm -Rf flamingzombies

.PHONY: clean
clean:
	rm -f $(BINS)

.FORCE:
