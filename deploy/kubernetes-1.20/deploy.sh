#!/usr/bin/env bash

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

# Modifications Copyright 2021 vazmin.
# Licensed under the Apache License, Version 2.0.


# This script captures the steps required to successfully
# deploy the fcfs plugin driver.  This should be considered
# authoritative and all updates for this process should be
# done here and referenced elsewhere.

# The script assumes that kubectl is available on the OS path
# where it is executed.

set -e
set -o pipefail

BASE_DIR=$(dirname "$0")


# Some images are not affected by *_REGISTRY/*_TAG and IMAGE_* variables.
# The default is to update unless explicitly excluded.
update_image() {
  case "$1" in socat) return 1 ;; esac
}

run() {
  echo "$@" >&2
  "$@"
}

for i in $(ls ${BASE_DIR}/rbac/*.yaml | sort); do
  run kubectl apply -f "$i"
done

# deploy fcfs plugin and registrar sidecar
echo "deploying fcfs components"
for i in $(ls ${BASE_DIR}/fcfs/*.yaml | sort); do
  run kubectl apply -f "$i"
#  echo "   $i"
#  modified="$(cat "$i" | while IFS= read -r line; do
#    nocomments="$(echo "$line" | sed -e 's/ *#.*$//')"
#    if echo "$nocomments" | grep -q '^[[:space:]]*image:[[:space:]]*'; then
#      # Split 'image: quay.io/k8scsi/csi-attacher:v1.0.1'
#      # into image (quay.io/k8scsi/csi-attacher:v1.0.1),
#      # registry (quay.io/k8scsi),
#      # name (csi-attacher),
#      # tag (v1.0.1).
#      image=$(echo "$nocomments" | sed -e 's;.*image:[[:space:]]*;;')
#      registry=$(echo "$image" | sed -e 's;\(.*\)/.*;\1;')
#      name=$(echo "$image" | sed -e 's;.*/\([^:]*\).*;\1;')
#      tag=$(echo "$image" | sed -e 's;.*:;;')
#
#      # Variables are with underscores and upper case.
#      varname=$(echo $name | tr - _ | tr a-z A-Z)
#
#      # Now replace registry and/or tag, if set as env variables.
#      # If not set, the replacement is the same as the original value.
#      # Only do this for the images which are meant to be configurable.
#      if update_image "$name"; then
#        prefix=$(eval echo \${${varname}_REGISTRY:-${IMAGE_REGISTRY:-${registry}}}/ | sed -e 's;none/;;')
#        if [ "$IMAGE_TAG" = "canary" ] &&
#          [ -f ${BASE_DIR}/canary-blacklist.txt ] &&
#          grep -q "^$name\$" ${BASE_DIR}/canary-blacklist.txt; then
#          # Ignore IMAGE_TAG=canary for this particular image because its
#          # canary image is blacklisted in the deployment blacklist.
#          suffix=$(eval echo :\${${varname}_TAG:-${tag}})
#        else
#          suffix=$(eval echo :\${${varname}_TAG:-${IMAGE_TAG:-${tag}}})
#        fi
#        line="$(echo "$nocomments" | sed -e "s;$image;${prefix}${name}${suffix};")"
#      fi
#      echo "        using $line" >&2
#    fi
#    echo "$line"
#  done)"
#  if ! echo "$modified" | kubectl apply -f -; then
#    echo "modified version of $i:"
#    echo "$modified"
#    exit 1
#  fi
done

# Wait until all pods are running. We have to make some assumptions
# about the deployment here, otherwise we wouldn't know what to wait
# for: the expectation is that we run attacher, provisioner,
# resizer, socat and fcfs plugin in the default namespace.
expected_running_pods=5
cnt=0
while [ $(kubectl get pods 2>/dev/null | grep '^csi-fcfs.* Running ' | wc -l) -lt ${expected_running_pods} ]; do
  if [ $cnt -gt 30 ]; then
    echo "$(kubectl get pods 2>/dev/null | grep '^csi-fcfs.* Running ' | wc -l) running pods:"
    kubectl describe pods

    echo >&2 "ERROR: fcfs deployment not ready after over 5min"
    exit 1
  fi
  echo $(date +%H:%M:%S) "waiting for fcfs deployment to complete, attempt #$cnt"
  cnt=$(($cnt + 1))
  sleep 10
done

# Create a test driver configuration in the place where the prow job
# expects it?
#if [ "${CSI_PROW_TEST_DRIVER}" ]; then
#    cp "${BASE_DIR}/test-driver.yaml" "${CSI_PROW_TEST_DRIVER}"
#
#    # When testing late binding, pods must be forced to run on the
#    # same node as the fcfs driver. external-provisioner currently
#    # doesn't handle the case when the "wrong" node is chosen and gets
#    # stuck permanently with:
#    # error generating accessibility requirements: no topology key found on CSINode csi-prow-worker2
#    echo >>"${CSI_PROW_TEST_DRIVER}" "ClientNodeName: $(kubectl get pods/csi-fcfs-provisioner-0  -o jsonpath='{.spec.nodeName}')"
#fi
