build:
	$(MAKE) -j2 multi-build

build-node:
	$(MAKE) -C react-material-ui build

build-stats:
	go build ./cmd/stats

build-get-all-bugs:
	go build ./cmd/get-all-bugs

build-bugs-per-team:
	go build ./cmd/bugs-per-team

build-bugtool:
	go build ./cmd/bugtool

build-smartsheet:
	go build ./cmd/shartsheets

build-go: build-stats build-get-all-bugs build-bugs-per-team
	go build ./

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

clean:
	rm ./main
