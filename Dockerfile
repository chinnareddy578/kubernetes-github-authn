FROM golang:1.9.4-alpine as build

ARG SRC_REPO=github.com/jjo/kubernetes-github-authn
ARG SRC_TAG=master
ARG ARCH=amd64
ARG BINARY=main

RUN apk --no-cache --update add ca-certificates make git

#RUN go get ${SRC_REPO}
COPY . /go/src/${SRC_REPO}
WORKDIR ${GOPATH}/src/${SRC_REPO}
RUN make
RUN cp -p _output/main /kubernetes-github-authn

FROM alpine:3.7
RUN apk --no-cache --update add ca-certificates
MAINTAINER JuanJo Ciarlante <juanjosec@gmail.com>

COPY --from=build kubernetes-github-authn /boot

USER 1001
EXPOSE 3000
CMD ["/boot"]
