FROM centos:centos8
ARG FASTCFS_VERSION
RUN rpm -ivh http://www.fastken.com/yumrepo/el8/x86_64/FastOSrepo-1.0.0-1.el8.x86_64.rpm \
 && yum remove fuse -y \
 && yum install ${FASTCFS_VERSION} -y