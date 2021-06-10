#!/bin/bash

# Copyright 2021 vazmin.
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

set -euo pipefail
BASE_DIR=$(dirname "$(realpath "${BASH_SOURCE[0]}")")

source "${BASE_DIR}"/util.sh


DRIVER_NAME=${DRIVER_NAME:-fcfs-csi-driver}
CONTROLLER_POD_NAME=${CONTROLLER_POD_NAME:-$(kubectl get po | grep fcfs-csi-controller | awk '{print $1}')}
CONTAINER_NAME=${CONTAINER_NAME:-fcfs-plugin}

TEST_ID=${TEST_ID:-$RANDOM}

TEST_DIR=${BASE_DIR}/csi-test-artifacts
BIN_DIR=${TEST_DIR}/bin


ZONES=${AVAILABILITY_NODE_NAME:-kind-control-plane}
FIRST_ZONE=$(echo "${ZONES}" | cut -d, -f1)

FCFS_ADMIN_NAME=${FCFS_ADMIN_NAME:-admin}
FCFS_USER_NAME=${FCFS_USER_NAME:-admin}
TEST_EXTRA_FLAGS="-config-url=${FCFS_CONFIG_URL} -username=${FCFS_USER_NAME} -admin-name=${FCFS_ADMIN_NAME}"
TEST_EXTRA_FLAGS+=" -controller-pod-name=${CONTROLLER_POD_NAME} -container-name=${CONTAINER_NAME} -allowed-topology-values=${ZONES}"

TEST_PATH=${TEST_PATH:-"./tests/e2e/..."}
ARTIFACTS=${ARTIFACTS:-"${TEST_DIR}/artifacts"}
GINKGO_FOCUS=${GINKGO_FOCUS:-"\[fcfs-csi-e2e\]"}
GINKGO_SKIP=${GINKGO_SKIP:-"\[Disruptive\]"}
GINKGO_NODES=${GINKGO_NODES:-1}
TEST_EXTRA_FLAGS=${TEST_EXTRA_FLAGS:-}


loudecho "Testing in  ${ZONES}"
mkdir -p "${BIN_DIR}"
export PATH=${PATH}:${BIN_DIR}


loudecho "Installing ginkgo to ${BIN_DIR}"
GINKGO_BIN=${BIN_DIR}/ginkgo
if [[ ! -e ${GINKGO_BIN} ]]; then
  pushd /tmp
  GOPATH=${TEST_DIR} GOBIN=${BIN_DIR} GO111MODULE=on go get github.com/onsi/ginkgo/ginkgo@v1.12.0
  popd
fi



loudecho "Testing focus ${GINKGO_FOCUS}"
#eval "EXPANDED_TEST_EXTRA_FLAGS=$TEST_EXTRA_FLAGS"
set -x
set +e
${GINKGO_BIN} -p -nodes="${GINKGO_NODES}" -v --focus="${GINKGO_FOCUS}" --skip="${GINKGO_SKIP}" "${TEST_PATH}" -- -report-dir="${ARTIFACTS}" ${TEST_EXTRA_FLAGS}
TEST_PASSED=$?
set -e
set +x
loudecho "TEST_PASSED: ${TEST_PASSED}"


