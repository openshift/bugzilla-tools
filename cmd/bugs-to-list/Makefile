NAME=bugs-to-list

build:
	go build ./

run: build
	./$(NAME) --org-data-url=http://localhost:8000/teams

save-bug-data: build
	./$(NAME) --org-data-url=http://localhost:8000/teams > testBugData

container: build
	podman build -t quay.io/$(USER)/$(NAME):latest .

container-push: container
	podman push quay.io/$(USER)/$(NAME):latest

container-run: container
	podman run -ti -v ./bugzillaKey:/bugzillaKey:z -v ./githubKey:/githubKey:z -v ./overwriteTeamData.yml:/overwriteTeamData.yml:z quay.io/$(USER)/$(NAME):latest /$(NAME) --data-from-github --overwrite-team-data=overwriteTeamData.yml
clean:
	rm ./$(NAME)
