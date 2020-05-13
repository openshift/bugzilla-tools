build:
	$(MAKE) -j2 multi-build

build-node:
	$(MAKE) -C react-material-ui build

build-go:
	go build ./
	go build ./cmd/stats
	go build ./cmd/get-all-bugs
	go build ./cmd/bugs-per-team

multi-build: build-go build-node

run-go: build-go
	./react-material

run-node:
	$(MAKE) -C react-material-ui run

run-prometheus:
	podman run --network=host -v ./prometheus.yml:/etc/prometheus/prometheus.yml:Z quay.io/prometheus/prometheus

run: build-node run-go

clean:
	rm ./main
