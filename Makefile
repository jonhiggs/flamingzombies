release: export GITHUB_TOKEN="$(shell pass api/github.com/goreleaser)"
release:
ifndef MESSAGE
  $(error MESSAGE was not provided)
endif
ifndef VERSION
  $(error VERSION was not provided)
endif
	git tag -a v$(VERSION) -m "$(MESSAGE)"
	git push origin v$(VERSION)
	goreleaser build --clean
