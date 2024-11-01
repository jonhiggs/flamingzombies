DIST_GO = dist/bin/fzctl dist/bin/fz
DIST_LIBEXEC_GO = $(addprefix dist/libexec/flamingzombies/task/,diskfree ping swapfree)
DIST_LIBEXEC = $(subst libexec/,dist/libexec/flamingzombies/,$(wildcard libexec/task/* libexec/gate/* libexec/notifier/*))
DIST_MAN = $(addprefix dist/,$(wildcard man/man1/*.1) $(wildcard man/man5/*.5) $(wildcard man/man7/*.7))
DIST_SCRIPTS = $(addprefix dist/,$(wildcard scripts/*))
DIST_CONF = dist/example_config.toml

build: $(DIST_GO) $(DIST_LIBEXEC_GO) $(DIST_LIBEXEC) $(DIST_MAN) $(DIST_SCRIPTS) $(DIST_CONF)

$(DIST_GO): src = ./cmd/$(subst dist/,,$@)
$(DIST_GO): .FORCE | dist/bin
	go build -o $@ ./cmd/$(notdir $@)

$(DIST_LIBEXEC_GO): src = ./$(subst dist/libexec/flamingzombies/,cmd/,$@)
$(DIST_LIBEXEC_GO): .FORCE | $(addprefix dist/libexec/flamingzombies/, task gate notifier)
	go build -o $@ $(src)

$(DIST_LIBEXEC): src = $(subst dist/libexec/flamingzombies/,libexec/,$@)
$(DIST_LIBEXEC): | $(addprefix dist/libexec/flamingzombies/, task gate notifier)
	cp $(src) $@

$(DIST_MAN): src = $(subst dist/,,$@)
$(DIST_MAN): | $(addprefix dist/man/, man1 man5 man7)
	cp $(src) $@

$(DIST_SCRIPTS) $(DIST_CONF): | dist/scripts
	cp $(subst dist/,,$@) $@

dist/bin dist/libexec/flamingzombies/task dist/libexec/flamingzombies/gate dist/libexec/flamingzombies/notifier dist/man/man1 dist/man/man5 dist/man/man7 dist/scripts:
	mkdir -p $@

clean:
	rm -Rf ./dist/*

.FORCE:
