FROM golang:1.5
MAINTAINER Hugo González Labrador

ENV CLAWIO_LOCALSTOREDATA_DATADIR=/tmp
ENV CLAWIO_LOCALSTOREDATA_TMPDIR=/tmp
ENV CLAWIO_LOCALSTOREDATA_PORT=57002
ENV CLAWIO_SHAREDSECRET=secret

RUN go get -u github.com/clawio/service.localstore.data

ENTRYPOINT /go/bin/service.localstore.data

EXPOSE 57002

