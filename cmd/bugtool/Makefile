build-and-run: build run

build:
	go build ./

run: build
	./bugtool

container: build
	podman build -t quay.io/${USER}/bugtool:latest .

container-run: container
	podman run -ti -v ./apikey:/apikey:z quay.io/${USER}/bugtool:latest
