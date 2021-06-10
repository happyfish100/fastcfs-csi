FROM golang:1.16.3 as builder
WORKDIR /go/src/vazmin.github.io/fastcfs-csi
COPY . .
RUN make

FROM centos:centos8

RUN rpm -ivh http://www.fastken.com/yumrepo/el8/x86_64/FastOSrepo-1.0.0-1.el8.x86_64.rpm \
 && yum remove fuse -y \
 && yum install FastCFS-fused -y

ENV TZ Asia/Shanghai

LABEL maintainers="vazmin"
LABEL description="FastCFS Driver"


COPY --from=builder /go/src/vazmin.github.io/fastcfs-csi/bin/fcfsplugin  /fcfsplugin
ENTRYPOINT ["/fcfsplugin"]