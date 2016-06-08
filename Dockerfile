FROM golang:alpine

RUN apk update && apk add curl git mercurial bzr && rm -rf /var/cache/apk/*

RUN mkdir -p /go/src/watbot
WORKDIR /go/src/watbot

COPY . /go/src/watbot

RUN go build

CMD ./watbot
