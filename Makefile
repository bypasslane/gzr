.PHONY: build clean doc gen run test vet install_deps build_web vendor_install

DEPEND=\
		github.com/bypasslane/boxedRice/boxedRice


excluding_vendor := $(shell go list ./... | grep -v /vendor/)

default: build

build:
	go build -i -o gzr

build_web: build
	cd gozer-web; yarn global add webpack; yarn; yarn build;
	boxedRice append -b=public --exec=./gzr

install_build_deps:
	go get -u $(DEPEND)

watch_web:
	cd gozer-web; npm run watch;

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
