# Copyright 2021 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


FROM golang:1.16.3-alpine3.13 as builder
WORKDIR /go/src/vazmin.github.io/fastcfs-csi
ADD . .
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -mod vendor -o _output/fcfsfused-proxy ./pkg/fcfsfused-proxy

FROM centos:centos8
COPY --from=builder /go/src/sigs.k8s.io/blob-csi-driver/_output/fcfsfused-proxy /fcfsfused-proxy

RUN rpm -ivh http://www.fastken.com/yumrepo/el8/x86_64/FastOSrepo-1.0.0-1.el8.x86_64.rpm \
 && yum remove fuse -y \
 && yum install FastCFS-fused -y

ENTRYPOINT ["/fcfsfused-proxy"]
