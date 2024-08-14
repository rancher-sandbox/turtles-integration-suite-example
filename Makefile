# Copyright © 2023 - 2024 SUSE LLC
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

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

CURL_RETRIES=3

# Directories

ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
BIN_DIR := bin
TEST_DIR := test
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/$(BIN_DIR))

$(TOOLS_BIN_DIR):
	mkdir -p $@

export PATH := $(abspath $(TOOLS_BIN_DIR)):$(PATH)

# Dependencies

# Helper function to get dependency version from go.mod
get_go_version = $(shell go list -m $1 | awk '{print $$2}')

GO_INSTALL := ./scripts/go-install.sh

GINKGO_BIN := ginkgo
GINGKO_VER := $(call get_go_version,github.com/onsi/ginkgo/v2)
GINKGO := $(abspath $(TOOLS_BIN_DIR)/$(GINKGO_BIN)-$(GINGKO_VER))
GINKGO_PKG := github.com/onsi/ginkgo/v2/ginkgo

HELM_VER := v3.14.0
HELM_BIN := helm
HELM := $(TOOLS_BIN_DIR)/$(HELM_BIN)-$(HELM_VER)

CLUSTERCTL_VER := v1.4.6
CLUSTERCTL_BIN := clusterctl
CLUSTERCTL := $(TOOLS_BIN_DIR)/$(CLUSTERCTL_BIN)-$(CLUSTERCTL_VER)

.PHONY: $(GINKGO_BIN)
$(GINKGO_BIN): $(GINKGO) ## Build a local copy of ginkgo.

$(GINKGO): # Build ginkgo from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(GINKGO_PKG) $(GINKGO_BIN) $(GINGKO_VER)

$(CLUSTERCTL): $(TOOLS_BIN_DIR) ## Download and install clusterctl
	curl --retry $(CURL_RETRIES) -fsSL -o $(CLUSTERCTL) https://github.com/kubernetes-sigs/cluster-api/releases/download/$(CLUSTERCTL_VER)/clusterctl-linux-amd64
	chmod +x $(CLUSTERCTL)

$(HELM): ## Put helm into tools folder.
	mkdir -p $(TOOLS_BIN_DIR)
	rm -f "$(TOOLS_BIN_DIR)/$(HELM_BIN)*"
	curl --retry $(CURL_RETRIES) -fsSL -o $(TOOLS_BIN_DIR)/get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
	chmod 700 $(TOOLS_BIN_DIR)/get_helm.sh
	USE_SUDO=false HELM_INSTALL_DIR=$(TOOLS_BIN_DIR) DESIRED_VERSION=$(HELM_VER) BINARY_NAME=$(HELM_BIN)-$(HELM_VER) $(TOOLS_BIN_DIR)/get_helm.sh
	ln -sf $(HELM) $(TOOLS_BIN_DIR)/$(HELM_BIN)
	rm -f $(TOOLS_BIN_DIR)/get_helm.sh

kubectl: # Download kubectl cli into tools bin folder
	scripts/ensure-kubectl.sh \
		-b $(TOOLS_BIN_DIR) \
		$(KUBECTL_VERSION)

#
# Ginkgo configuration.
#
GINKGO_FOCUS ?=
GINKGO_SKIP ?=
GINKGO_NODES ?= 1
GINKGO_TIMEOUT ?= 2h
GINKGO_POLL_PROGRESS_AFTER ?= 60m
GINKGO_POLL_PROGRESS_INTERVAL ?= 5m
E2E_CONF_FILE ?= $(ROOT_DIR)/config/config.yaml
GINKGO_ARGS ?=
GINKGO_NOCOLOR ?= false
GINKGO_LABEL_FILTER ?= 
GINKGO_TESTS ?= $(ROOT_DIR)/suites/...

MANAGEMENT_CLUSTER_ENVIRONMENT ?= kind
SKIP_RESOURCE_CLEANUP ?= false
USE_EXISTING_CLUSTER ?= false
GITEA_CUSTOM_INGRESS ?= false

E2ECONFIG_VARS ?= MANAGEMENT_CLUSTER_ENVIRONMENT=$(MANAGEMENT_CLUSTER_ENVIRONMENT) \
ARTIFACTS=$(ARTIFACTS) \
HELM_BINARY_PATH=$(HELM) \
CLUSTERCTL_BINARY_PATH=$(CLUSTERCTL) \
SKIP_RESOURCE_CLEANUP=$(SKIP_RESOURCE_CLEANUP) \
USE_EXISTING_CLUSTER=$(USE_EXISTING_CLUSTER) \

.PHONY: test
test: $(GINKGO) $(HELM) $(CLUSTERCTL) kubectl
	$(E2ECONFIG_VARS) $(GINKGO) -v --trace -poll-progress-after=$(GINKGO_POLL_PROGRESS_AFTER) \
		-poll-progress-interval=$(GINKGO_POLL_PROGRESS_INTERVAL) --tags=e2e --focus="$(GINKGO_FOCUS)" --label-filter="$(GINKGO_LABEL_FILTER)" \
		$(_SKIP_ARGS) --nodes=$(GINKGO_NODES) --timeout=$(GINKGO_TIMEOUT) --no-color=$(GINKGO_NOCOLOR) \
		--output-dir="$(ARTIFACTS)" --junit-report="junit.e2e_suite.1.xml" $(GINKGO_ARGS) $(GINKGO_TESTS) -- \
	    -e2e.config="$(E2E_CONF_FILE)" \