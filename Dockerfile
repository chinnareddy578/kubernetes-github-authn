FROM golang:1.13-alpine as build

ARG SRC_REPO=github.com/chinnareddy578/kubernetes-github-authn
ARG SRC_TAG=master
ARG ARCH=amd64

RUN apk --no-cache --update add ca-certificates make git

#RUN go get ${SRC_REPO}
COPY . /go/src/${SRC_REPO}
WORKDIR ${GOPATH}/src/${SRC_REPO}
RUN make
RUN cp -p _output/main /main

FROM alpine:3.7
RUN apk --no-cache --update add ca-certificates

COPY --from=build /main /kubernetes-github-authn

USER 1001
EXPOSE 3000
CMD ["/kubernetes-github-authn"]
