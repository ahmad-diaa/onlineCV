FROM golang:latest

ADD .   /go/src/server/
WORKDIR /go/src/server/
ENTRYPOINT /go/src/server/app -configfile=myconfig.json
EXPOSE 3000
