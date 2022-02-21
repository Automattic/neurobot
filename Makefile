default: build

build:
	go build -o bin/neurobot

test:
	go test

coverage:
	cd engine
	go test -coverprofile=c.out
	go tool cover -html=c.out
	rm c.out
