build-and-run: build run

build:
	go build ./

run: build
	./bugautomation

container: build
	podman build -t quay.io/${USER}/bugautomation:latest .

container-run: container
	podman run -ti -v ./bugzillaKey:/bugzillaKey:z quay.io/${USER}/bugautomation:latest
