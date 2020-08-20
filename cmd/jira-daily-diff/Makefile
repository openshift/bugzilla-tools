mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
NAME := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

build:
	go build ./

run: build
	./$(NAME)

apply-manifests: container-push
	oc create secret generic jira-key --from-file=jiraKey --dry-run=client -o yaml | oc apply -f -
	oc apply -f manifest/

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./bugzillaKey:/bugzillaKey:z -v ./githubKey:/githubKey:z -v ./overwriteTeamData.yml:/overwriteTeamData.yml:z quay.io/eparis/$(NAME):latest /$(NAME) --data-from-github --overwrite-team-data=overwriteTeamData.yml

clean:
	rm ./$(NAME)
