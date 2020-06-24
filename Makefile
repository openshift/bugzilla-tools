IMAGE_REGISTRY?=quay.io

dirs := $(wildcard cmd/*)
names := $(notdir $(dirs))

all: build
.PHONY: all

# Include the library makefile
include $(addprefix ./vendor/github.com/openshift/build-machinery-go/make/, \
	golang.mk \
	targets/openshift/deps-gomod.mk \
	targets/openshift/images.mk \
)


define build-image-internal
image-$(1):
	podman build -t $(IMAGE_REGISTRY)/$(USER)/$(1):dev -f ./cmd/$(1)/Dockerfile .
.PHONY: image-$(1)
endef

define build-image
$(eval $(call build-image-internal,$(1)))
endef

$(foreach name,$(names),$(call build-image,$(name)))


TOPTARGETS := apply-config
$(TOPTARGETS): $(dirs)
$(dirs):
	$(MAKE) -C "$@" $(MAKECMDGOALS)
.PHONY: $(TOPTARGETS) $(dirs)

