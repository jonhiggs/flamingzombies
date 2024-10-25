FZ_VERSION := $(shell cat cmd/fz/fz.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')
FZCTL_VERSION := $(shell cat cmd/fzctl/fzctl.go | awk '/const VERSION/ { gsub(/"/,"",$$NF); print $$NF }')
VERSION := $(FZ_VERSION)

# make and publish a release
# I don't know if this won't be moved to a post_recieve hook.

release: prerelease_tests release_notes.txt
	gh release create $(VERSION) -F release_notes.txt
	gh release upload $(VERSION) dist/fz_* dist/fzctl_* dist/plugins.tar.gz

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

clean:
	rm -f release_notes.txt
