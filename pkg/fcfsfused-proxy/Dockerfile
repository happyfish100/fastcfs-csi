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

#
# Modifications Copyright 2021 vazmin.
# Licensed under the Apache License, Version 2.0.
#

ARG FASTCFS_IMAGE

FROM golang:1.16.3-alpine3.13 as builder
WORKDIR /go/src/vazmin.github.io/fastcfs-csi
ADD . .
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -mod vendor -o _output/fcfsfused-proxy ./pkg/fcfsfused-proxy

FROM ${FASTCFS_IMAGE}
COPY --from=builder /go/src/vazmin.github.io/fastcfs-csi/_output/fcfsfused-proxy /fcfsfused-proxy

ENTRYPOINT ["/fcfsfused-proxy"]
