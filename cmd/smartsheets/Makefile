NAME=smartsheets

build:
	go build ./

run: build
	./$(NAME) --overwrite-team-data=overwriteTeamData.yml --data-from-github

apply-config: container-push
	oc create secret generic bugzilla-api-key --from-file=bugzillaKey --dry-run=client -o yaml | oc apply -f -
	oc create secret generic smartsheets-api-key --from-file=smartsheetKey --dry-run=client -o yaml | oc apply -f -
	oc apply -f deployment/$(NAME).cronjob.yml

run-job: container-push
	oc delete job -l app=$(NAME),type=oneshot
	oc apply -f deployment/$(NAME).job.yml

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./overwriteTeamData.yml:/overwriteTeamData.yml:z -v ./githubKey:/githubKey:z -v ./bugzillaKey:/bugzillaKey:z -v ./smartsheetKey:/smartsheetKey:z quay.io/eparis/smartsheets:latest /smartsheets --overwrite-team-data=overwriteTeamData.yml --data-from-github

clean:
	rm ./$(NAME)
