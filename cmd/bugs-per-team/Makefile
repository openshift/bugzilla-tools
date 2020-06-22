NAME=bugs-per-team

build:
	go build ./

run: build
	./$(NAME)

apply-config: container-push
	#oc create secret generic bugzilla-api-key --from-file=bugzillaKey --dry-run=client -o yaml | oc apply -f -
	#oc apply -f deployment/$(NAME).cronjob.yml

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./bugzillaKey:/etc/bugzilla/bugzillaKey:z -v ./overwriteTeamData.yml:/var/run/overwriteTeamData.yml:z -v ./githubKey:/etc/github/githubKey:z quay.io/eparis/bugs-per-team:latest /bugs-per-team --bugzilla-key=/etc/bugzilla/bugzillaKey --github-key=/etc/github/githubKey --overwrite-team-data=/var/run/overwriteTeamData.yml --data-from-github

clean:
	rm ./$(NAME)
