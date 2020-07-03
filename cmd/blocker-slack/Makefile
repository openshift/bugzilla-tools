mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
NAME := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

build:
	go build ./

run: build
	./$(NAME) --overwrite-team-data=overwriteTeamData.yml --data-from-github

apply-manifest: container-push
	oc create secret slack-auth --from-file=slackKey --dry-run=client -o yaml | oc apply -f
	oc create configmap blocker-slack-config --from-file=config.yml --dry-run=client -o yaml | oc apply -f -
	oc create configmap overwrite-team-data --from-file=overwriteTeamData.yml --dry-run=client -o yaml | oc apply -f -
	oc apply -f manifests/
	oc delete pod -l app=$(NAME)

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./overwriteTeamData.yml:/var/run/overwriteTeamData.yml:z -v ./githubKey:/etc/github/githubKey:z --network=host quay.io/${USER}/$(NAME):latest

clean:
	rm ./$(NAME)
