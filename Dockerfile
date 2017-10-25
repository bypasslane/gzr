FROM alpine:latest

ADD gzr /usr/local/bin/gzr
ADD docker-config/gzr.json /root/.gzr.json
ADD start.sh /usr/local/bin/start.sh

RUN apk update && \
    apk add bash

ENTRYPOINT ["start.sh"]
