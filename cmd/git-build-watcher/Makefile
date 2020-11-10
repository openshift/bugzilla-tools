mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
NAME := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

build:
	go build ./

run: build
	./$(NAME)

apply-manifest: container-push
	oc apply -f manifests/
	oc delete pod -l app=$(NAME)

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

clean:
	rm ./$(NAME)
