FROM alpine:3.3

RUN apk update \
 && apk add jq \
 && rm -rf /var/cache/apk/*

COPY script.sh ./

RUN chmod 777 ./script.sh

