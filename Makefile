mains := $(wildcard cmd/*/Makefile)
dirs = $(dir $(mains))

TOPTARGETS := build clean container container-push apply-config

pull-ubi:
	podman pull registry.access.redhat.com/ubi8/ubi-minimal

$(TOPTARGETS): pull-ubi $(dirs)
$(dirs):
	$(MAKE) -C "$@" $(MAKECMDGOALS)

.PHONY: $(TOPTARGETS) $(dirs)
