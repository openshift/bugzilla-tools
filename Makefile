build:
	$(MAKE) -j2 multi-build

build-node:
	$(MAKE) -C react-material-ui build

multi-build: build-go build-node

run-node:
	$(MAKE) -C react-material-ui run

run: build-node run-go

clean:
	rm ./main
