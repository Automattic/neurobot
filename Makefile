default: build

build:
	go build -o bin/neurobot

test:
	go test -v ./engine

coverage:
	cd engine
	go test ./... -coverprofile=c.out
	go tool cover -html=c.out
	rm c.out
