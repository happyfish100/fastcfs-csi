ARG FASTCFS_IMAGE

FROM golang:1.16.3 as builder
WORKDIR /go/src/vazmin.github.io/fastcfs-csi
COPY . .
RUN make

FROM ${FASTCFS_IMAGE}

ENV TZ Asia/Shanghai

LABEL maintainers="vazmin"
LABEL description="The FastCFS Container Storage Interface (CSI) Driver"


COPY --from=builder /go/src/vazmin.github.io/fastcfs-csi/bin/fcfsplugin  /fcfsplugin
ENTRYPOINT ["/fcfsplugin"]