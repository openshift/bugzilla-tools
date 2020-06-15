build:
	go build ./

run: build
	./bugautomation

apply-config: container-push
	oc create secret generic bugzilla-api-key --from-file=bugzillaKey --dry-run=client -o yaml | oc apply -f -
	oc apply -f deployment/bugautomation.cronjob.yml

container: build
	podman build -t quay.io/$(USER)/bugautomation:latest .

container-push: container
	podman push quay.io/$(USER)/bugautomation:latest

container-run: container
	podman run -ti -v ./bugzillaKey:/etc/bugzilla/bugzillaKey:z quay.io/${USER}/bugautomation:latest
clean:
	rm ./bugautomation
