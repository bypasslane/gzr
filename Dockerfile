FROM golang:1.9-alpine as BUILD

WORKDIR /go/src/github.com/bypasslane/gzr
RUN apk update && \
    apk add git nodejs nodejs-npm yarn make ca-certificates && \
    update-ca-certificates && \
    git config --global url."https://".insteadOf git:// && \
    git config --global url."https://".insteadOf ssh:// && \
    git config --global url."https://github.com/".insteadOf git@github.com: && \
    go get github.com/Masterminds/glide

ADD glide.lock .
ADD glide.yaml .
ADD Makefile .

RUN make install_build_deps && \
    make install

ADD . /go/src/github.com/bypasslane/gzr

RUN make build_web

FROM alpine:latest

COPY --from=BUILD /go/src/github.com/bypasslane/gzr/gzr /usr/local/bin/gzr
ADD docker/gzr.json /root/.gzr.json
ADD docker/entrypoint.sh /usr/local/bin/entrypoint.sh

RUN apk update && \
    apk add bash && \
    mkdir -p /root/.gzr && \
    touch /root/.gzr/gzr.db

ENTRYPOINT ["entrypoint.sh"]
