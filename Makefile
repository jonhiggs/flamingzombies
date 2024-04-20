release: export GITHUB_TOKEN="$(shell pass api/github.com/ghcli)"
release:
	goreleaser release --clean
