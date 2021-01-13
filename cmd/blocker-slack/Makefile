mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
NAME := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

build:
	go build ./

run: build
	curl http://localhost:8000/teams > /dev/null 2>&1 || ( echo "You likely need to run 'cmd/team-exportor' before you can run this" && /bin/false )
	./$(NAME) --debug --org-data-url=http://localhost:8000/teams

apply-manifest: container-push
	oc create secret generic slack-key --from-file=slackKey --dry-run=client -o yaml | oc apply -f -
	oc create secret generic bugzilla-api-key --from-file=bugzillaKey --dry-run=client -o yaml | oc apply -f -
	oc create configmap blocker-slack-config --from-file=config.yaml --dry-run=client -o yaml | oc apply -f -
	oc apply -f manifests/
	oc delete pod -l app=$(NAME)

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./slackKey:/slackKey:z -v ./overwriteTeamData.yml:/overwriteTeamData.yml:z -v ./bugzillaKey:/bugzillaKey:z -v ./githubKey:/githubKey:z -v ./config.yaml:/config.yaml:z quay.io/${USER}/$(NAME):latest /$(NAME) --data-from-github

clean:
	rm ./$(NAME)
