NAME=bugs-to-list

build:
	go build ./

run: build
	./$(NAME)

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./bugzillaKey:/etc/bugzilla/bugzillaKey:z quay.io/${USER}/$(NAME):latest
clean:
	rm ./$(NAME)
