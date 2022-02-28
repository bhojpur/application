# Docker image build and push setting
DOCKER:=docker
DOCKERFILE_DIR?=./docker

APP_SYSTEM_IMAGE_NAME=$(RELEASE_NAME)
APP_RUNTIME_IMAGE_NAME=appsvr
APP_PLACEMENT_IMAGE_NAME=placement
APP_SENTRY_IMAGE_NAME=sentry

# build docker image for linux
BIN_PATH=$(OUT_DIR)/$(TARGET_OS)_$(TARGET_ARCH)

ifeq ($(TARGET_OS), windows)
  DOCKERFILE:=Dockerfile-windows
  BIN_PATH := $(BIN_PATH)/release
else ifeq ($(origin DEBUG), undefined)
  DOCKERFILE:=Dockerfile
  BIN_PATH := $(BIN_PATH)/release
else ifeq ($(DEBUG),0)
  DOCKERFILE:=Dockerfile
  BIN_PATH := $(BIN_PATH)/release
else
  DOCKERFILE:=Dockerfile-debug
  BIN_PATH := $(BIN_PATH)/debug
endif

ifeq ($(TARGET_ARCH),arm)
  DOCKER_IMAGE_PLATFORM:=$(TARGET_OS)/arm/v7
else ifeq ($(TARGET_ARCH),arm64)
  DOCKER_IMAGE_PLATFORM:=$(TARGET_OS)/arm64/v8
else
  DOCKER_IMAGE_PLATFORM:=$(TARGET_OS)/amd64
endif

# Supported docker image architecture
DOCKERMUTI_ARCH=linux-amd64 linux-arm linux-arm64 windows-amd64

################################################################################
# Target: docker-build, docker-push                                            #
################################################################################

LINUX_BINS_OUT_DIR=$(OUT_DIR)/linux_$(GOARCH)
DOCKER_IMAGE_TAG=$(APP_REGISTRY)/$(APP_SYSTEM_IMAGE_NAME):$(APP_TAG)
APP_RUNTIME_DOCKER_IMAGE_TAG=$(APP_REGISTRY)/$(APP_RUNTIME_IMAGE_NAME):$(APP_TAG)
APP_PLACEMENT_DOCKER_IMAGE_TAG=$(APP_REGISTRY)/$(APP_PLACEMENT_IMAGE_NAME):$(APP_TAG)
APP_SENTRY_DOCKER_IMAGE_TAG=$(APP_REGISTRY)/$(APP_SENTRY_IMAGE_NAME):$(APP_TAG)

ifeq ($(LATEST_RELEASE),true)
DOCKER_IMAGE_LATEST_TAG=$(APP_REGISTRY)/$(APP_SYSTEM_IMAGE_NAME):$(LATEST_TAG)
APP_RUNTIME_DOCKER_IMAGE_LATEST_TAG=$(APP_REGISTRY)/$(APP_RUNTIME_IMAGE_NAME):$(LATEST_TAG)
APP_PLACEMENT_DOCKER_IMAGE_LATEST_TAG=$(APP_REGISTRY)/$(APP_PLACEMENT_IMAGE_NAME):$(LATEST_TAG)
APP_SENTRY_DOCKER_IMAGE_LATEST_TAG=$(APP_REGISTRY)/$(APP_SENTRY_IMAGE_NAME):$(LATEST_TAG)
endif


# To use buildx: https://github.com/docker/buildx#docker-ce
export DOCKER_CLI_EXPERIMENTAL=enabled

# check the required environment variables
check-docker-env:
ifeq ($(APP_REGISTRY),)
	$(error APP_REGISTRY environment variable must be set)
endif
ifeq ($(APP_TAG),)
	$(error APR_TAG environment variable must be set)
endif

check-arch:
ifeq ($(TARGET_OS),)
	$(error TARGET_OS environment variable must be set)
endif
ifeq ($(TARGET_ARCH),)
	$(error TARGET_ARCH environment variable must be set)
endif


docker-build: check-docker-env check-arch
	$(info Building $(DOCKER_IMAGE_TAG) docker image ...)
ifeq ($(TARGET_ARCH),amd64)
	$(DOCKER) build --build-arg PKG_FILES=* -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	$(DOCKER) build --build-arg PKG_FILES=appsvr -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(APP_RUNTIME_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	$(DOCKER) build --build-arg PKG_FILES=placement -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(APP_PLACEMENT_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	$(DOCKER) build --build-arg PKG_FILES=sentry -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(APP_SENTRY_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
else
	-$(DOCKER) buildx create --use --name appsvrbuild
	-$(DOCKER) run --rm --privileged multiarch/qemu-user-static --reset -p yes
	$(DOCKER) buildx build --build-arg PKG_FILES=* --platform $(DOCKER_IMAGE_PLATFORM) -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	$(DOCKER) buildx build --build-arg PKG_FILES=appsvr --platform $(DOCKER_IMAGE_PLATFORM) -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(APP_RUNTIME_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	$(DOCKER) buildx build --build-arg PKG_FILES=placement --platform $(DOCKER_IMAGE_PLATFORM) -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(APP_PLACEMENT_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	$(DOCKER) buildx build --build-arg PKG_FILES=sentry --platform $(DOCKER_IMAGE_PLATFORM) -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(APP_SENTRY_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
endif

# push docker image to the registry
docker-push: docker-build
	$(info Pushing $(DOCKER_IMAGE_TAG) docker image ...)
ifeq ($(TARGET_ARCH),amd64)
	$(DOCKER) push $(DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	$(DOCKER) push $(APP_RUNTIME_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	$(DOCKER) push $(APP_PLACEMENT_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	$(DOCKER) push $(APP_SENTRY_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
else
	-$(DOCKER) buildx create --use --name appsvrbuild
	-$(DOCKER) run --rm --privileged multiarch/qemu-user-static --reset -p yes
	$(DOCKER) buildx build --build-arg PKG_FILES=* --platform $(DOCKER_IMAGE_PLATFORM) -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH) --push
	$(DOCKER) buildx build --build-arg PKG_FILES=appsvr --platform $(DOCKER_IMAGE_PLATFORM) -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(APP_RUNTIME_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH) --push
	$(DOCKER) buildx build --build-arg PKG_FILES=placement --platform $(DOCKER_IMAGE_PLATFORM) -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(APP_PLACEMENT_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH) --push
	$(DOCKER) buildx build --build-arg PKG_FILES=sentry --platform $(DOCKER_IMAGE_PLATFORM) -f $(DOCKERFILE_DIR)/$(DOCKERFILE) $(BIN_PATH) -t $(APP_SENTRY_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH) --push
endif

# push docker image to kind cluster
docker-push-kind: docker-build
	$(info Pushing $(DOCKER_IMAGE_TAG) docker image to kind cluster...)
	kind load docker-image $(DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	kind load docker-image $(APP_RUNTIME_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	kind load docker-image $(APP_PLACEMENT_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)
	kind load docker-image $(APP_SENTRY_DOCKER_IMAGE_TAG)-$(TARGET_OS)-$(TARGET_ARCH)

# publish muti-arch docker image to the registry
docker-manifest-create: check-docker-env
	$(DOCKER) manifest create $(DOCKER_IMAGE_TAG) $(DOCKERMUTI_ARCH:%=$(DOCKER_IMAGE_TAG)-%)
	$(DOCKER) manifest create $(APP_RUNTIME_DOCKER_IMAGE_TAG) $(DOCKERMUTI_ARCH:%=$(APP_RUNTIME_DOCKER_IMAGE_TAG)-%)
	$(DOCKER) manifest create $(APP_PLACEMENT_DOCKER_IMAGE_TAG) $(DOCKERMUTI_ARCH:%=$(APP_PLACEMENT_DOCKER_IMAGE_TAG)-%)
	$(DOCKER) manifest create $(APP_SENTRY_DOCKER_IMAGE_TAG) $(DOCKERMUTI_ARCH:%=$(APP_SENTRY_DOCKER_IMAGE_TAG)-%)
ifeq ($(LATEST_RELEASE),true)
	$(DOCKER) manifest create $(DOCKER_IMAGE_LATEST_TAG) $(DOCKERMUTI_ARCH:%=$(DOCKER_IMAGE_TAG)-%)
	$(DOCKER) manifest create $(APP_RUNTIME_DOCKER_IMAGE_LATEST_TAG) $(DOCKERMUTI_ARCH:%=$(APP_RUNTIME_DOCKER_IMAGE_TAG)-%)
	$(DOCKER) manifest create $(APP_PLACEMENT_DOCKER_IMAGE_LATEST_TAG) $(DOCKERMUTI_ARCH:%=$(APP_PLACEMENT_DOCKER_IMAGE_TAG)-%)
	$(DOCKER) manifest create $(APP_SENTRY_DOCKER_IMAGE_LATEST_TAG) $(DOCKERMUTI_ARCH:%=$(APP_SENTRY_DOCKER_IMAGE_TAG)-%)
endif

docker-publish: docker-manifest-create
	$(DOCKER) manifest push $(DOCKER_IMAGE_TAG)
	$(DOCKER) manifest push $(APP_RUNTIME_DOCKER_IMAGE_TAG)
	$(DOCKER) manifest push $(APP_PLACEMENT_DOCKER_IMAGE_TAG)
	$(DOCKER) manifest push $(APP_SENTRY_DOCKER_IMAGE_TAG)
ifeq ($(LATEST_RELEASE),true)
	$(DOCKER) manifest push $(DOCKER_IMAGE_LATEST_TAG)
	$(DOCKER) manifest push $(APP_RUNTIME_DOCKER_IMAGE_LATEST_TAG)
	$(DOCKER) manifest push $(APP_PLACEMENT_DOCKER_IMAGE_LATEST_TAG)
	$(DOCKER) manifest push $(APP_SENTRY_DOCKER_IMAGE_LATEST_TAG)
endif

check-windows-version:
ifeq ($(WINDOWS_VERSION),)
	$(error WINDOWS_VERSION environment variable must be set)
endif

docker-windows-base-build: check-windows-version
	$(DOCKER) build --build-arg WINDOWS_VERSION=$(WINDOWS_VERSION) -f $(DOCKERFILE_DIR)/$(DOCKERFILE)-base . -t $(APP_REGISTRY)/windows-base:$(WINDOWS_VERSION)
	$(DOCKER) build --build-arg WINDOWS_VERSION=$(WINDOWS_VERSION) -f $(DOCKERFILE_DIR)/$(DOCKERFILE)-java-base . -t $(APP_REGISTRY)/windows-java-base:$(WINDOWS_VERSION)
	$(DOCKER) build --build-arg WINDOWS_VERSION=$(WINDOWS_VERSION) -f $(DOCKERFILE_DIR)/$(DOCKERFILE)-php-base . -t $(APP_REGISTRY)/windows-php-base:$(WINDOWS_VERSION)
	$(DOCKER) build --build-arg WINDOWS_VERSION=$(WINDOWS_VERSION) -f $(DOCKERFILE_DIR)/$(DOCKERFILE)-python-base . -t $(APP_REGISTRY)/windows-python-base:$(WINDOWS_VERSION)

docker-windows-base-push: check-windows-version
	$(DOCKER) push $(APP_REGISTRY)/windows-base:$(WINDOWS_VERSION)
	$(DOCKER) push $(APP_REGISTRY)/windows-java-base:$(WINDOWS_VERSION)
	$(DOCKER) push $(APP_REGISTRY)/windows-php-base:$(WINDOWS_VERSION)
	$(DOCKER) push $(APP_REGISTRY)/windows-python-base:$(WINDOWS_VERSION)

################################################################################
# Target: build-dev-container, push-dev-container                              #
################################################################################

# Update whenever you upgrade dev container image
DEV_CONTAINER_VERSION_TAG?=0.1.6

# Use this to pin a specific version of the Bhojpur Application CLI to a devcontainer
DEV_CONTAINER_CLI_TAG?=1.6.0

# Bhojpur Application container image name
DEV_CONTAINER_IMAGE_NAME=app-dev

DEV_CONTAINER_DOCKERFILE=Dockerfile-dev
DOCKERFILE_DIR=./docker

check-docker-env-for-dev-container:
ifeq ($(APP_REGISTRY),)
	$(error APP_REGISTRY environment variable must be set)
endif

build-dev-container:
ifeq ($(APP_REGISTRY),)
	$(info APP_REGISTRY environment variable not set, tagging image without registry prefix.)
	$(info `make tag-dev-container` should be run with APP_REGISTRY before `make push-dev-container.)
	$(DOCKER) build --build-arg APP_CLI_VERSION=$(DEV_CONTAINER_CLI_TAG) -f $(DOCKERFILE_DIR)/$(DEV_CONTAINER_DOCKERFILE) $(DOCKERFILE_DIR)/. -t $(DEV_CONTAINER_IMAGE_NAME):$(DEV_CONTAINER_VERSION_TAG)
else
	$(DOCKER) build --build-arg APP_CLI_VERSION=$(DEV_CONTAINER_CLI_TAG) -f $(DOCKERFILE_DIR)/$(DEV_CONTAINER_DOCKERFILE) $(DOCKERFILE_DIR)/. -t $(APP_REGISTRY)/$(DEV_CONTAINER_IMAGE_NAME):$(DEV_CONTAINER_VERSION_TAG)
endif

tag-dev-container: check-docker-env-for-dev-container
	$(DOCKER) tag $(DEV_CONTAINER_IMAGE_NAME):$(DEV_CONTAINER_VERSION_TAG) $(APP_REGISTRY)/$(DEV_CONTAINER_IMAGE_NAME):$(DEV_CONTAINER_VERSION_TAG)

push-dev-container: check-docker-env-for-dev-container
	$(DOCKER) push $(APP_REGISTRY)/$(DEV_CONTAINER_IMAGE_NAME):$(DEV_CONTAINER_VERSION_TAG)