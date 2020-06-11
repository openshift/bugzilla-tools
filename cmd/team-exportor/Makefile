NAME=team-exportor

build:
	go build ./

run: build
	./$(NAME) --overwrite-team-data=overwriteTeamData.yml --data-from-github

apply-config: container-push
	oc create secret generic github-api-key --from-file=githubKey --dry-run=client -o yaml | oc apply -f -
	oc create configmap overwrite-team-data --from-file=overwriteTeamData.yml  --dry-run=client -o yaml | oc apply -f -
	oc apply -f deployment/$(NAME).deployment.yml
	oc apply -f deployment/$(NAME).service.yml
	oc delete pod -l app=$(NAME)

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./overwriteTeamData.yml:/var/run/overwriteTeamData.yml:z -v ./githubKey:/etc/github/githubKey:z --network=host quay.io/${USER}/$(NAME):latest

clean:
	rm ./$(NAME)
