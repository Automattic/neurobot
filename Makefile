default: build

build:
	go build -o neurobot

test:
	go test -v ./...

coverage:
	go test ./... -coverprofile=c.out
	go tool cover -html=c.out
	rm c.out
