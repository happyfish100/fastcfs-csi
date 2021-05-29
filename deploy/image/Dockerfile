FROM centos:centos8

RUN rpm -ivh http://www.fastken.com/yumrepo/el8/x86_64/FastOSrepo-1.0.0-1.el8.x86_64.rpm \
 && yum remove fuse -y \
 && yum install FastCFS-fused -y

ENV TZ Asia/Shanghai

LABEL maintainers="vazmin"
LABEL description="FastCFS Driver"
ARG binary=./bin/fcfsplugin


COPY ${binary} /fcfsplugin
ENTRYPOINT ["/fcfsplugin"]