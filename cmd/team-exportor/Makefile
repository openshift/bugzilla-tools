NAME=team-exportor

build:
	go build ./

run: build
	./$(NAME) --test-team-data=overwriteTestTeamData.yml

apply-config:
	oc create secret generic github-api-key --from-file=githubKey --dry-run=client -o yaml | oc apply -f -
	oc apply -f deployment/$(NAME).deployment.yml
	oc apply -f deployment/$(NAME).service.yml

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./bugzillaKey:/etc/bugzilla/bugzillaKey:z quay.io/${USER}/$(NAME):latest
clean:
	rm ./$(NAME)
