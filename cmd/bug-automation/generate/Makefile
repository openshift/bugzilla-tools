all: build run

build:
	go build ./

run: build
	./generate --test-bugs=$(TEST_BUGS)

.PHONY: all build run
