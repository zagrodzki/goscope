FROM circleci/golang:latest

USER root

RUN apt-get update && apt-get install libusb-1.0-0-dev gawk

USER circleci

RUN go get github.com/golang/lint/golint

RUN go get github.com/jstemmer/go-junit-report
