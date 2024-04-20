release: export GITHUB_TOKEN="$(shell pass api/github.com/goreleaser)"
release:
	goreleaser release --clean
