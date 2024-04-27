MANS = $(wildcard man*/*.[0-9])

generate_pdfs: $(addsuffix .pdf,$(MANS))

man1/%.pdf man5/%.pdf man7/%.pdf:
	man -Tpdf $(subst .pdf,,$@) > $@

clean:
	find . -iname \*.pdf -delete
