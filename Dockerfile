FROM golang:1.5
MAINTAINER Hugo González Labrador

ENV CLAWIO_LOCALFS_DATA_DATADIR /tmp/localfs
ENV CLAWIO_LOCALFS_DATA_TMPDIR /tmp/localfs
ENV CLAWIO_LOCALFS_DATA_PORT 57002
ENV CLAWIO_LOCALFS_DATA_CHECKSUM md5
ENV CLAWIO_LOCALFS_DATA_PROP "service-localfs-prop:57003"
ENV CLAWIO_SHAREDSECRET secret

ADD . /go/src/github.com/clawio/service-localfs-data
WORKDIR /go/src/github.com/clawio/service-localfs-data

RUN go get -u github.com/tools/godep
RUN godep restore
RUN go install

ENTRYPOINT /go/bin/service-localfs-data

EXPOSE 57002

