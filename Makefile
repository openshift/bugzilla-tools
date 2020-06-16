mains := $(wildcard cmd/*/Makefile)
dirs = $(dir $(mains))

TOPTARGETS := build clean container container-push apply-config

$(TOPTARGETS): $(dirs)
$(dirs):
	$(MAKE) -C "$@" $(MAKECMDGOALS)

.PHONY: $(TOPTARGETS) $(dirs)
