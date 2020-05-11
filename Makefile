build-node:
	$(MAKE) -C react-material-ui build

build-go:
	go build ./

multi-build: build-go build-node

build:
	$(MAKE) -j2 multi-build

run-go: build-go
	./main

run-node:
	$(MAKE) -C react-material-ui run

run-prometheus:
	podman run --network=host -v ./prometheus.yml:/etc/prometheus/prometheus.yml:Z quay.io/prometheus/prometheus

run: build-node run-go

clean:
	rm ./main
