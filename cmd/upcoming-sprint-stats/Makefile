NAME=upcoming-sprint-stats

build:
	go build ./

run: build
	./$(NAME)

apply-config: container-push
	oc create secret generic bugzilla-api-key --from-file=bugzillaKey --dry-run=client -o yaml | oc apply -f -
	oc apply -f deployment/$(NAME).cronjob.yml

run-job: container-push
	oc delete job -l app=$(NAME),type=oneshot
	oc apply -f deployment/$(NAME).job.yml

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./bugzillaKey:/bugzillaKey:z -v ./githubKey:/githubKey:z -v ./overwriteTeamData.yml:/overwriteTeamData.yml:z quay.io/eparis/$(NAME):latest /$(NAME) --data-from-github --overwrite-team-data=overwriteTeamData.yml

clean:
	rm ./$(NAME)
