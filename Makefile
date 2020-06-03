build:
	$(MAKE) -j2 multi-build

build-node:
	$(MAKE) -C react-material-ui build

build-upcoming-sprint-stats:
	go build ./cmd/upcoming-sprint-stats

build-get-all-bugs:
	go build ./cmd/get-all-bugs

build-bugs-per-team:
	go build ./cmd/bugs-per-team

build-bugtool:
	go build ./

build-smartsheet:
	go build ./cmd/shartsheets

build-go: build-upcoming-sprint-stats build-get-all-bugs build-bugs-per-team build-bugtool

multi-build: build-go build-node

run-go: build-go
	./bugtool --test-team-data=testTeamData.yml

run-node:
	$(MAKE) -C react-material-ui run

run-prometheus:
	podman run --network=host -v ./data:/prometheus:z -v ./prometheus.yml:/etc/prometheus/prometheus.yml:z quay.io/prometheus/prometheus

run-bugs-per-team: build-bugs-per-team
	./bugs-per-team --test-team-data=testTeamData.yml

run: build-node run-go

apply-config:
	oc create secret generic bugzilla-api-key --from-file=bugzillaKey --dry-run=client -o yaml | oc apply -f -
	oc create secret generic github-api-key --from-file=githubKey --dry-run=client -o yaml | oc apply -f -
	oc create configmap prometheus-config --from-file=prometheus.yml --dry-run=client -o yaml | oc apply -f -
	oc create configmap test-team-data --from-file=testTeamData.yml  --dry-run=client -o yaml | oc apply -f -
	oc apply -f deployment/bugtool.deployment.yml
	oc apply -f deployment/bugtool.service.yml
	oc apply -f deployment/prometheus.deployment.yml
	oc apply -f deployment/prometheus.service.yml
	oc apply -f deployment/prometheus.route.yml

container: build-bugtool
	podman build -t quay.io/$(USER)/bugtool:latest .

container-push: container
	podman push quay.io/$(USER)/bugtool:latest

clean:
	rm ./main
