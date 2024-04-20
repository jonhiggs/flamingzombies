SHELL := /bin/bash


ifndef MESSAGE
  $(error MESSAGE was not provided)
endif
ifndef VERSION
  $(error VERSION was not provided)
endif

TAG := v$(VERSION)

artifacts := $(addsuffix .tar.gz, dist/fz_openbsd_arm64 dist/fz_openbsd_amd64 dist/fz_linux_arm64 dist/fz_linux_amd64 dist/fz_darwin_arm64 dist/fz_darwin_amd64)

release: $(artifacts)
	gh release create $(TAG) --notes "${MESSAGE}"
	gh release upload $(TAG) $(artifacts)

dist/fz_darwin_amd64.tar.gz:  DIR := dist/fz_darwin_amd64_v1
dist/fz_darwin_arm64.tar.gz:  DIR := dist/fz_darwin_arm64
dist/fz_linux_amd64.tar.gz:   DIR := dist/fz_linux_amd64_v1
dist/fz_linux_arm64.tar.gz:   DIR := dist/fz_linux_arm64
dist/fz_openbsd_amd64.tar.gz: DIR := dist/fz_openbsd_amd64_v1
dist/fz_openbsd_arm64.tar.gz: DIR := dist/fz_openbsd_arm64
dist/%.tar.gz: gorelease_build
	mkdir -p $(DIR)/bin
	mkdir -p $(DIR)/share/man1
	mkdir -p $(DIR)/share/flamingzombies
	mv $(DIR)/fz $(DIR)/bin/fz
	cp -r plugins $(DIR)/share/flamingzombies
	tar -C dist -zcvf $@ $(notdir $(DIR))

dist/plugins.tar.bz:
	tar xvf $@ plugins/

gorelease_build:
	git tag -a $(TAG) -m "$(MESSAGE)"
	git push origin $(TAG)
	goreleaser build --clean
	goreleaser build --snapshot --clean

clean:
	rm -Rf ./dist
