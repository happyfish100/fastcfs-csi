
CMDS=fcfsplugin

CONTAINER_CMD=$(shell docker version >/dev/null 2>&1 && echo docker)
ifeq ($(CONTAINER_CMD),)
	CONTAINER_CMD?=$(shell podman version >/dev/null 2>&1 && echo podman)
endif

all: build

BIN_OUTPUT=bin


FASTCFS_VERSION=5.3.0
FASTCFS_NAME=vazmin/fastcfs-fused
FASTCFS_IMAGE=$(FASTCFS_NAME):v$(FASTCFS_VERSION)

CSI_IMAGE_VERSION=v0.4.6-fastcfs5.3.0-1
CSI_IMAGE_NAME=$(if $(ENV_CSI_IMAGE_NAME),$(ENV_CSI_IMAGE_NAME),vazmin/fcfs-csi)
CSI_IMAGE=$(CSI_IMAGE_NAME):$(CSI_IMAGE_VERSION)

PKG=vazmin.github.io/fastcfs-csi
GOOS=$(shell go env GOOS)

ifndef REV
# Revision that gets built into each binary via the main.version
# string. Uses the `git describe` output based on the most recent
# version tag with a short revision suffix or, if nothing has been
# tagged yet, just the revision.
#
# Beware that tags may also be missing in shallow clones as done by
# some CI systems (like TravisCI, which pulls only 50 commits).
REV=$(shell git describe --long --tags --match='v*' --dirty 2>/dev/null || git rev-list -n1 HEAD)
endif
GIT_COMMIT?=$(shell git rev-parse HEAD)

# BUILD_PLATFORMS contains a set of <os> <arch> <suffix> triplets,
# separated by semicolon. An empty variable or empty entry (= just a
# semicolon) builds for the default platform of the current Go
# toolchain.
BUILD_PLATFORMS =
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# Add go ldflags using LDFLAGS at the time of compilation.
LDFLAGS ?=
# CSI_IMAGE_VERSION will be considered as the driver version
LDFLAGS += -X $(PKG)/pkg/common.DriverVersion=$(REV) -X ${PKG}/pkg/common.GitCommit=${GIT_COMMIT} -X ${PKG}/pkg/common.BuildDate=${BUILD_DATE}
FULL_LDFLAGS = $(LDFLAGS) $(EXT_LDFLAGS)

# This builds each command (= the sub-directories of ./cmd) for the target platform(s)
# defined by BUILD_PLATFORMS.
$(CMDS:%=build-%): build-%:
	mkdir -p bin
	echo '$(BUILD_PLATFORMS)' | tr ';' '\n' | while read -r os arch suffix; do \
		if ! (set -x; CGO_ENABLED=0 GOOS="$$os" GOARCH="$$arch" go build $(GOFLAGS_VENDOR) -a -ldflags '$(FULL_LDFLAGS)' -o "${BIN_OUTPUT}/$*$$suffix" ./cmd/$*); then \
			echo "Building $* for GOOS=$$os GOARCH=$$arch failed, see error(s) above."; \
			exit 1; \
		fi; \
	done

build: $(CMDS:%=build-%)

.PHONY: image-csi
# image-csi: GOARCH ?= $(shell go env GOARCH 2>/dev/null)
image-csi: .container-cmd
	$(CONTAINER_CMD) build --build-arg FASTCFS_IMAGE=$(FASTCFS_IMAGE) -t $(CSI_IMAGE) .

image-clean: .container-cmd
	$(CONTAINER_CMD) image ls | grep $(CSI_IMAGE_VERSION) | grep $(CSI_IMAGE_NAME) | awk '{print $$3}' | xargs -r $(CONTAINER_CMD) image rm -f

.container-cmd:
	@test -n "$(shell which $(CONTAINER_CMD) 2>/dev/null)" || { echo "Missing container support, install Podman or Docker"; exit 1; }
	@echo "$(CONTAINER_CMD)" > .container-cmd

kind-load-image:
	kind load docker-image $(CSI_IMAGE)

kind-clean:
	$(CONTAINER_CMD) exec kind-control-plane bash -c "crictl image | grep $(CSI_IMAGE_NAME) | grep $(CSI_IMAGE_VERSION) | awk '{print $$3}' | xargs -r crictl rmi"

local-deploy: image-clean image-csi kind-load-image

node-restart:
	kubectl get po | grep fcfs-csi-node |awk '{print $$1}'| xargs -i kubectl delete po {}

node-log:
	kubectl get po | grep fcfs-csi-node | awk '{print $$1}'| xargs -i kubectl logs {} fcfs-plugin -f

controller-restart:
	kubectl get po | grep fcfs-csi-controller |awk '{print $$1}'| xargs -i kubectl delete po {}

controller-log:
	kubectl get po | grep fcfs-csi-controller |awk '{print $$1}'| xargs -i kubectl logs {} fcfs-plugin -f

.PHONY: image-fastcfs
image-fastcfs: .container-cmd
	$(CONTAINER_CMD) build --build-arg FASTCFS_VERSION=FastCFS-fused-$(FASTCFS_VERSION)-1.el8.x86_64 -t $(FASTCFS_IMAGE) -f deploy/fastcfs-fused/Dockerfile .

.PHONY: fcfsfused-proxy
fcfsfused-proxy:
	CGO_ENABLED=0 GOOS=linux go build -mod vendor -ldflags="-s -w" -o _output/fcfsfused-proxy ./pkg/fcfsfused-proxy

.PHONY: fcfsfused-proxy-container
fcfsfused-proxy-container:
	docker build --build-arg FASTCFS_IMAGE=$(FASTCFS_IMAGE) -t vazmin/fcfsfused-proxy -f pkg/fcfsfused-proxy/Dockerfile .

.PHONY: install-fcfsfused-proxy
install-fcfsfused-proxy:
	kubectl apply -f ./deploy/fcfsfused-proxy/fcfsfused-proxy.yaml

.PHONY: uninstall-fcfsfused-proxy
uninstall-fcfsfused-proxy:
	kubectl delete -f ./deploy/fcfsfused-proxy/fcfsfused-proxy.yaml --ignore-not-found

clean:
	-rm -rf ${BIN_OUTPUT}


bin /tmp/helm /tmp/kubeval:
	@mkdir -p $@

bin/helm: | /tmp/helm bin
	@curl -o /tmp/helm/helm.tar.gz -sSL https://get.helm.sh/helm-v3.6.0-${GOOS}-amd64.tar.gz
	@tar -zxf /tmp/helm/helm.tar.gz -C bin --strip-components=1
	@rm -rf /tmp/helm/*


BASE_YAML = fastcfs-client-configmap.yaml
KUBE_YAML = csidriver.yaml controller.yaml node.yaml
RBAC_YAML = clusterrole-attacher.yaml clusterrole-csi-node.yaml clusterrole-provisioner.yaml clusterrole-resizer.yaml \
			clusterrolebinding-attacher.yaml clusterrolebinding-csi-node.yaml clusterrolebinding-provisioner.yaml clusterrolebinding-resizer.yaml \
			poddisruptionbudget-controller.yaml serviceaccount-csi-controller.yaml serviceaccount-csi-node.yaml

.PHONY: generate-kustomize

# if `WARNING: Kubernetes configuration file is group-readable. This is insecure.`
# exec `chmod go-r ~/.kube/config`
generate-kustomize: bin/helm $(BASE_YAML) $(KUBE_YAML) $(RBAC_YAML)

%.yaml:
	cd charts/fcfs-csi-driver && ../../bin/helm template kustomize . -s templates/$@ > ../../deploy/kubernetes/base/$@


.PHONY: test-e2e-single-nn
test-e2e-single-nn:
	AVAILABILITY_NODE_NAME=kind-control-plane \
	TEST_PATH=./tests/e2e/... \
	GINKGO_FOCUS="\[fcfs-csi-e2e\] \[single-nn\]" \
	FCFS_CONFIG_URL=http://192.168.99.180:8080 \
	./hack/e2e/run.sh

.PHONY: test-e2e-multi-nn
test-e2e-multi-nn:
	AVAILABILITY_NODE_NAME=kind-control-plane \
	TEST_PATH=./tests/e2e/... \
	GINKGO_FOCUS="\[fcfs-csi-e2e\] \[multi-nn\]" \
	FCFS_CONFIG_URL=http://192.168.99.180:8080 \
	./hack/e2e/run.sh
