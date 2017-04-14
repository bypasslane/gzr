.PHONY: build clean doc gen run test vet

excluding_vendor := $(shell go list ./... | grep -v /vendor/)

default: build

vendor_install:
	glide i
build:
	go build -i -o gzr


clean:
	rm gzr

run:
	go build -o gzr && ./gzr

test:
	go test -v $(excluding_vendor)

local_install:
	go install `go list | grep -v /vendor/`

install:
	glide install

doc:
	godoc -http=:8080 -index

vet:
	go vet ./..
