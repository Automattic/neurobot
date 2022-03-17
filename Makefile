default: build

build:
	go build -o neurobot

test:
	go test -v ./...

coverage:
	cd engine
	go test ./... -coverprofile=c.out
	go tool cover -html=c.out
	rm c.out
