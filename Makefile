.PHONY: build clean doc gen run test vet install_deps build_web vendor_install

DEPEND=\
		github.com/GeertJohan/go.rice/rice


excluding_vendor := $(shell go list ./... | grep -v /vendor/)

default: build

vendor_install:
	glide i

build:
	go build -i -o gzr

build_web: build
	cd gozer-web; npm i -g webpack; npm i; npm run build;
	rice -i=github.com/bypasslane/gzr/controllers append --exec=./gzr

install_deps:
	go get -u $(DEPEND)
	go install $(DEPEND)

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
