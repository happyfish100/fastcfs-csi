
CMDS=fcfsplugin

CONTAINER_CMD=$(shell docker version >/dev/null 2>&1 && echo docker)
ifeq ($(CONTAINER_CMD),)
	CONTAINER_CMD?=$(shell podman version >/dev/null 2>&1 && echo podman)
endif

all: build

BIN_OUTPUT=bin
FCFS_CSI_VERSION=$(shell . $(CURDIR)/build.env ; echo $${FCFS_CSI_VERSION})

CSI_IMAGE_NAME=$(if $(ENV_CSI_IMAGE_NAME),$(ENV_CSI_IMAGE_NAME),hub.docker.com/r/vazmin/fcfs-csi)
CSI_IMAGE_VERSION=$(shell . $(CURDIR)/build.env ; echo $${CSI_IMAGE_VERSION})
CSI_IMAGE=$(CSI_IMAGE_NAME):$(CSI_IMAGE_VERSION)

GO_PROJECT=github.com/happyfish100/fastcfs-csi

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

# BUILD_PLATFORMS contains a set of <os> <arch> <suffix> triplets,
# separated by semicolon. An empty variable or empty entry (= just a
# semicolon) builds for the default platform of the current Go
# toolchain.
BUILD_PLATFORMS =

# Add go ldflags using LDFLAGS at the time of compilation.
LDFLAGS ?=
# CSI_IMAGE_VERSION will be considered as the driver version
LDFLAGS += -X $(GO_PROJECT)/pkg/common.DriverVersion=$(REV)
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

# image-csi: GOARCH ?= $(shell go env GOARCH 2>/dev/null)
image-csi: .container-cmd
	$(CONTAINER_CMD) build -t $(CSI_IMAGE) -f deploy/image/Dockerfile .

image-clean: .container-cmd
	$(CONTAINER_CMD) image ls | grep $(CSI_IMAGE_VERSION) | grep $(CSI_IMAGE_NAME) | awk '{print $$3}' | xargs -r $(CONTAINER_CMD) image rm -f


kind-load-image:
	kind load docker-image $(CSI_IMAGE)

kind-clean:
	$(CONTAINER_CMD) exec kind-control-plane bash -c "crictl image | grep $(CSI_IMAGE_NAME) | grep $(CSI_IMAGE_VERSION) | awk '{print $$3}' | xargs -r crictl rmi"

delete-plugin-po:
	kubectl delete po csi-fcfsplugin-0

local-deploy: image-clean build image-csi kind-load-image delete-plugin-po


.container-cmd:
	@test -n "$(shell which $(CONTAINER_CMD) 2>/dev/null)" || { echo "Missing container support, install Podman or Docker"; exit 1; }
	@echo "$(CONTAINER_CMD)" > .container-cmd


clean:
	-rm -rf ${BIN_OUTPUT}