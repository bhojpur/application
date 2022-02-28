################################################################################
# Variables                                                                    #
################################################################################

export GO111MODULE ?= on
export GOPROXY ?= https://proxy.golang.org
export GOSUMDB ?= sum.golang.org

GIT_COMMIT  = $(shell git rev-list -1 HEAD)
GIT_VERSION = $(shell git describe --always --abbrev=7 --dirty)
# By default, disable CGO_ENABLED. See the details on https://golang.org/cmd/cgo
CGO         ?= 0
CLI_BINARY  ?= appctl
BINARIES    ?= appsvr placement operator injector sentry
HA_MODE     ?= false
# Force in-memory log for placement
FORCE_INMEM ?= true

# Add latest tag if LATEST_RELEASE is true
LATEST_RELEASE ?=

PROTOC ?=protoc
# name of protoc-gen-go when protoc-gen-go --version is run.
PROTOC_GEN_GO_NAME = "protoc-gen-go"
ifdef REL_VERSION
	CLI_VERSION := $(REL_VERSION)
else
	CLI_VERSION := edge
endif

ifdef API_VERSION
	RUNTIME_API_VERSION = $(API_VERSION)
else
	RUNTIME_API_VERSION = 1.0
endif

ifdef REL_VERSION
	APP_VERSION := $(REL_VERSION)
else
	APP_VERSION := edge
endif

LOCAL_ARCH := $(shell uname -m)
ifeq ($(LOCAL_ARCH),x86_64)
	TARGET_ARCH_LOCAL=amd64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 5),armv8)
	TARGET_ARCH_LOCAL=arm64
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 4),armv)
	TARGET_ARCH_LOCAL=arm
else ifeq ($(shell echo $(LOCAL_ARCH) | head -c 5),arm64)
	TARGET_ARCH_LOCAL=arm64
else
	TARGET_ARCH_LOCAL=amd64
endif
export GOARCH ?= $(TARGET_ARCH_LOCAL)

ifeq ($(GOARCH),amd64)
	LATEST_TAG=latest
else
	LATEST_TAG=latest-$(GOARCH)
endif

LOCAL_OS := $(shell uname)
ifeq ($(LOCAL_OS),Linux)
   TARGET_OS_LOCAL = linux
else ifeq ($(LOCAL_OS),Darwin)
   TARGET_OS_LOCAL = darwin
else
   TARGET_OS_LOCAL ?= windows
   BINARY_EXT_LOCAL = .exe
   PROTOC_GEN_GO_NAME := "protoc-gen-go.exe"
endif
export GOOS ?= $(TARGET_OS_LOCAL)

PROTOC_GEN_GO_NAME+= "v1.26.0"

# Default docker container and e2e test targst.
TARGET_OS ?= linux
TARGET_ARCH ?= amd64
TEST_OUTPUT_FILE_PREFIX ?= ./test_report
TEST_OUTPUT_FILE ?= test_output.json

ifeq ($(GOOS),windows)
BINARY_EXT_LOCAL:=.exe
GOLANGCI_LINT:=golangci-lint.exe
export ARCHIVE_EXT = .zip
else
BINARY_EXT_LOCAL:=
GOLANGCI_LINT:=golangci-lint
export ARCHIVE_EXT = .tar.gz
endif

export BINARY_EXT ?= $(BINARY_EXT_LOCAL)

OUT_DIR := ./dist

# Helm template and install setting
HELM:=helm
RELEASE_NAME?=appctl
APP_NAMESPACE?=app-system
APP_MTLS_ENABLED?=true
HELM_CHART_ROOT:=./charts
HELM_CHART_DIR:=$(HELM_CHART_ROOT)/app
HELM_OUT_DIR:=$(OUT_DIR)/install
HELM_MANIFEST_FILE:=$(HELM_OUT_DIR)/$(RELEASE_NAME).yaml
HELM_REGISTRY?=registry.bhojpur.net


################################################################################
# Go build details                                                             #
################################################################################
BASE_PACKAGE_NAME := github.com/bhojpur/application
LOGGER_PACKAGE_NAME := github.com/bhojpur/service/pkg/utils/logger
OUT_DIR := ./dist

# For the Bhojpur Application CLI
BINS_OUT_DIR := $(OUT_DIR)/$(GOOS)_$(GOARCH)/$(BUILDTYPE_DIR)
LDFLAGS := "-X client.version=$(CLI_VERSION) -X client.apiVersion=$(RUNTIME_API_VERSION) \
 -X $(BASE_PACKAGE_NAME)/pkg/standalone.gitcommit=$(GIT_COMMIT) -X $(BASE_PACKAGE_NAME)/pkg/standalone.gitversion=$(GIT_VERSION)"

# For the Bhojpur Application runtime engine
DEFAULT_LDFLAGS:=-X $(BASE_PACKAGE_NAME)/pkg/version.gitcommit=$(GIT_COMMIT) \
  -X $(BASE_PACKAGE_NAME)/pkg/version.gitversion=$(GIT_VERSION) \
  -X $(BASE_PACKAGE_NAME)/pkg/version.version=$(APP_VERSION) \
  -X $(LOGGER_PACKAGE_NAME).appVersion=$(APP_VERSION)

# Use the variable H to add a header (equivalent to =>) to informational output
H = $(shell printf "\033[34;1m=>\033[0m")

ifeq ($(origin DEBUG), undefined)
  BUILDTYPE_DIR:=release
  LDFLAGS:="$(DEFAULT_LDFLAGS) -s -w"
else ifeq ($(DEBUG),0)
  BUILDTYPE_DIR:=release
  LDFLAGS:="$(DEFAULT_LDFLAGS) -s -w"
else
  BUILDTYPE_DIR:=debug
  GCFLAGS:=-gcflags="all=-N -l"
  LDFLAGS:="$(DEFAULT_LDFLAGS)"
  $(info Build with debugger information)
endif

APP_OUT_DIR := $(OUT_DIR)/$(GOOS)_$(GOARCH)/$(BUILDTYPE_DIR)
APP_LINUX_OUT_DIR := $(OUT_DIR)/linux_$(GOARCH)/$(BUILDTYPE_DIR)

################################################################################
# Target: build                                                                #
################################################################################
.PHONY: build
APP_BINS:=$(foreach ITEM,$(BINARIES),$(APP_OUT_DIR)/$(ITEM)$(BINARY_EXT))
build: $(CLI_BINARY) $(APP_BINS)

$(CLI_BINARY):
	CGO_ENABLED=$(CGO) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(GCFLAGS) -ldflags $(LDFLAGS) \
	-o $(BINS_OUT_DIR)/$(CLI_BINARY)$(BINARY_EXT);

# Generate builds for Bhojpur Application runtime binaries for the target
# Params:
# $(1): the binary name for the target
# $(2): the binary main directory
# $(3): the target os
# $(4): the target arch
# $(5): the output directory
define genBinariesForTarget
.PHONY: $(5)/$(1)
$(5)/$(1):
	CGO_ENABLED=$(CGO) GOOS=$(3) GOARCH=$(4) go build $(GCFLAGS) -ldflags=$(LDFLAGS) \
	-o $(5)/$(1) $(2)/;
endef

# Generate binary targets
$(foreach ITEM,$(BINARIES),$(eval $(call genBinariesForTarget,$(ITEM)$(BINARY_EXT),./cmd/$(ITEM),$(GOOS),$(GOARCH),$(APP_OUT_DIR))))

################################################################################
# Target: build-linux                                                          #
################################################################################
BUILD_LINUX_BINS:=$(foreach ITEM,$(BINARIES),$(APP_LINUX_OUT_DIR)/$(ITEM))
build-linux: $(BUILD_LINUX_BINS)

# Generate linux binaries targets to build linux docker image
ifneq ($(GOOS), linux)
$(foreach ITEM,$(BINARIES),$(eval $(call genBinariesForTarget,$(ITEM),./cmd/$(ITEM),linux,$(GOARCH),$(APP_LINUX_OUT_DIR))))
endif

################################################################################
# Target: archive                                                              #
################################################################################
ARCHIVE_OUT_DIR ?= $(APP_OUT_DIR)
ARCHIVE_FILE_EXTS:=$(foreach ITEM,$(BINARIES),archive-$(ITEM)$(ARCHIVE_EXT))

archive: $(ARCHIVE_FILE_EXTS)

# Generate archive files for each binary
# $(1): the binary name to be archived
# $(2): the archived file output directory
define genArchiveBinary
ifeq ($(GOOS),windows)
archive-$(1).zip:
	7z.exe a -tzip "$(2)\\$(1)_$(GOOS)_$(GOARCH)$(ARCHIVE_EXT)" "$(APP_OUT_DIR)\\$(1)$(BINARY_EXT)"
else
archive-$(1).tar.gz:
	tar czf "$(2)/$(1)_$(GOOS)_$(GOARCH)$(ARCHIVE_EXT)" -C "$(APP_OUT_DIR)" "$(1)$(BINARY_EXT)"
endif
endef

# Generate archive-*.[zip|tar.gz] targets
$(foreach ITEM,$(BINARIES),$(eval $(call genArchiveBinary,$(ITEM),$(ARCHIVE_OUT_DIR))))


################################################################################
# Target: manifest-gen                                                         #
################################################################################

# Generate helm chart manifest
manifest-gen: app.yaml

app.yaml: check-docker-env
	$(info Generating helm manifest $(HELM_MANIFEST_FILE)...)
	@mkdir -p $(HELM_OUT_DIR)
	$(HELM) template \
		--include-crds=true  --set global.ha.enabled=$(HA_MODE) --set app_config.app_config_chart_included=false --set-string global.tag=$(APP_TAG) --set-string global.registry=$(APP_REGISTRY) $(HELM_CHART_DIR) > $(HELM_MANIFEST_FILE)

################################################################################
# Target: upload-helmchart
################################################################################

# Upload helm charts to Helm Registry
upload-helmchart:
	export HELM_EXPERIMENTAL_OCI=1; \
	$(HELM) chart save ${HELM_CHART_ROOT}/${RELEASE_NAME} ${HELM_REGISTRY}/${HELM}/${RELEASE_NAME}:${APP_VERSION}; \
	$(HELM) chart push ${HELM_REGISTRY}/${HELM}/${RELEASE_NAME}:${APP_VERSION}

################################################################################
# Target: docker-deploy-k8s                                                    #
################################################################################

PULL_POLICY?=Always
docker-deploy-k8s: check-docker-env check-arch
	$(info Deploying ${APP_REGISTRY}/${RELEASE_NAME}:${APP_TAG} to the current K8S context...)
	$(HELM) install \
		$(RELEASE_NAME) --namespace=$(APP_NAMESPACE) --wait --timeout 5m0s \
		--set global.ha.enabled=$(HA_MODE) --set-string global.tag=$(APP_TAG)-$(TARGET_OS)-$(TARGET_ARCH) \
		--set-string global.registry=$(APP_REGISTRY) --set global.logAsJson=true \
		--set global.appControlPlaneOs=$(TARGET_OS) --set global.appControlPlaneArch=$(TARGET_ARCH) \
		--set app_placement.logLevel=debug --set app_sidecar_injector.sidecarImagePullPolicy=$(PULL_POLICY) \
		--set global.imagePullPolicy=$(PULL_POLICY) --set global.imagePullSecrets=${APP_TEST_REGISTRY_SECRET} \
		--set global.mtls.enabled=${APP_MTLS_ENABLED} \
		--set app_placement.cluster.forceInMemoryLog=$(FORCE_INMEM) $(HELM_CHART_DIR)

################################################################################
# Target: release                                                              #
################################################################################
release: build archive

################################################################################
# Target: test                                                                 #
################################################################################
.PHONY: test
test: test-deps
	gotestsum --jsonfile $(TEST_OUTPUT_FILE_PREFIX)_unit.json --format standard-quiet -- ./pkg/... ./utils/... ./cmd/... $(COVERAGE_OPTS)
	go test ./test/...

################################################################################
# Target: lint                                                                 #
################################################################################
# Due to https://github.com/golangci/golangci-lint/issues/580, we need to add --fix for windows
.PHONY: lint
lint:
	$(GOLANGCI_LINT) run --timeout=20m

################################################################################
# Target: modtidy                                                              #
################################################################################
.PHONY: modtidy
modtidy:
	go mod tidy

################################################################################
# Target: init-proto                                                            #
################################################################################
.PHONY: init-proto
init-proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1

################################################################################
# Target: gen-proto                                                            #
################################################################################
GRPC_PROTOS:=common internals operator placement runtime sentry
PROTO_PREFIX:=github.com/bhojpur/application/pkg/api

# Generate archive files for each binary
# $(1): the binary name to be archived
define genProtoc
.PHONY: gen-proto-$(1)
gen-proto-$(1):
	$(PROTOC) --go_out=./pkg/api --go_opt=module=$(PROTO_PREFIX) --go-grpc_out=./pkg/api --go-grpc_opt=require_unimplemented_servers=false,module=$(PROTO_PREFIX) ./pkg/api/v1/$(1)/*.proto
endef

$(foreach ITEM,$(GRPC_PROTOS),$(eval $(call genProtoc,$(ITEM))))

GEN_PROTOS:=$(foreach ITEM,$(GRPC_PROTOS),gen-proto-$(ITEM))

.PHONY: gen-proto
gen-proto: check-proto-version $(GEN_PROTOS) modtidy

################################################################################
# Target: get-service-integration                                              #
################################################################################
.PHONY: get-service-integration
get-service-integration:
	go get github.com/bhojpur/service@master

################################################################################
# Target: check-diff                                                           #
################################################################################
.PHONY: check-diff
check-diff:
	git diff --exit-code ./go.mod # check no changes
	git diff --exit-code ./go.sum # check no changes

################################################################################
# Target: check-proto-version                                                  #
################################################################################
.PHONY: check-proto-version
check-proto-version: ## Checking the version of proto related tools
	@test "$(shell protoc --version)" = "libprotoc 3.19.4" \
	|| { echo "please use protoc 3.19.4 to generate proto, see https://github.com/bhojpur/application/blob/master/pkg/api/README.md#proto-client-generation"; exit 1; }

	@test "$(shell protoc-gen-go-grpc --version)" = "protoc-gen-go-grpc 1.1.0" \
	|| { echo "please use protoc-gen-go-grpc 1.1.0 to generate proto, see https://github.com/bhojpur/application/blob/master/pkg/api/README.md#proto-client-generation"; exit 1; }

	@test "$(shell protoc-gen-go --version 2>&1)" = "$(PROTOC_GEN_GO_NAME)" \
	|| { echo "please use protoc-gen-go v1.26.0 to generate proto, see https://github.com/bhojpur/application/blob/master/pkg/api/README.md#proto-client-generation"; exit 1; }

################################################################################
# Target: check-proto-diff                                                           #
################################################################################
.PHONY: check-proto-diff
check-proto-diff:
	git diff --exit-code ./pkg/api/v1/common/common.pb.go # check no changes
	git diff --exit-code ./pkg/api/v1/internals/status.pb.go # check no changes
	git diff --exit-code ./pkg/api/v1/operator/operator.pb.go # check no changes
	git diff --exit-code ./pkg/api/v1/operator/operator_grpc.pb.go # check no changes
	git diff --exit-code ./pkg/api/v1/runtime/appcallback.pb.go # check no changes
	git diff --exit-code ./pkg/api/v1/runtime/appcallback_grpc.pb.go # check no changes
	git diff --exit-code ./pkg/api/v1/runtime/app.pb.go # check no changes
	git diff --exit-code ./pkg/api/v1/runtime/appr_grpc.pb.go # check no changes
	git diff --exit-code ./pkg/api/v1/sentry/sentry.pb.go # check no changes


################################################################################
# Target: codegen                                                              #
################################################################################
include tools/codegen.mk

################################################################################
# Target: docker                                                               #
################################################################################
include docker/docker.mk

################################################################################
# Target: tests                                                                #
################################################################################
include test/app_tests.mk