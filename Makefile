TAG=v0.0.7

.PHONY: all

all: dep build test

build: gonzales

dep:
	go get github.com/tools/godep
	godep restore

gonzales: dep *.go
	godep go build

test:
	godep go test

run:
	godep go run

container: build
	./prepare_docker.sh

push-container: container
	docker tag gonzales gcr.io/noroutine/gonzales:${TAG}
	./deploy_gcp.sh ${TAG}
	
