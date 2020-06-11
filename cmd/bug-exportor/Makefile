NAME=bug-exportor

build:
	go build ./

run: build
	./$(NAME)

apply-config: container-push
	oc create secret generic bugzilla-api-key --from-file=bugzillaKey --dry-run=client -o yaml | oc apply -f -
	oc create secret generic github-api-key --from-file=githubKey --dry-run=client -o yaml | oc apply -f -
	oc create configmap test-team-data --from-file=testTeamData.yml  --dry-run=client -o yaml | oc apply -f -
	oc create configmap prometheus-config --from-file=deployment/prometheus.yml --dry-run=client -o yaml | oc apply -f -
	oc apply -f deployment/$(NAME).deployment.yml
	oc apply -f deployment/$(NAME).service.yml
	oc apply -f deployment/prometheus.deployment.yml
	oc apply -f deployment/prometheus.service.yml
	oc apply -f deployment/prometheus.route.yml

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./bugzillaKey:/etc/bugzilla/bugzillaKey:z quay.io/${USER}/$(NAME):latest
clean:
	rm ./$(NAME)
