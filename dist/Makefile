UID = $(shell id -u)
GID = $(shell id -g)

GOOS = $(shell uname -s | tr '[:upper:]' '[:lower:]')
GOARCH = $(shell uname -m | sed 's/x86_64/amd64/')

DIR = $(GOOS)_$(GOARCH)


BINS = $(DIR)/fz             \
	$(DIR)/fzctl         \
	$(DIR)/task/ping     \
	$(DIR)/task/diskfree \
	$(DIR)/task/swapfree \

build: $(BINS)

publish: DEST := git@janx.altos:/var/www/htdocs/artifacts.altos/flamingzombies/$(shell git rev-parse HEAD)
publish:
	rsync -av $(DIR) $(DEST)

$(BINS): export CGO_ENABLED = 1
$(BINS):
	mkdir -p $(dir $@)
	go build -o $@ -buildvcs=false ../cmd/$(subst $(DIR)/,,$@)

.PHONY: clean
clean:
	uname -s | tr '[:upper:]' '[:lower:]'
	#rm -rf $(DIR)
