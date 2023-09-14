SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

# use the binaries install from the tools.go
BIN_KO = go run github.com/google/ko
BIN_MDTOGO = go run github.com/GoogleContainerTools/kpt/mdtogo
BIN_LINT = go run github.com/golangci/golangci-lint/cmd/golangci-lint
BIN_YAMLFMT = go run github.com/google/yamlfmt/cmd/yamlfmt -conf=../../.yamlfmt

##@ Options

# # use the config file to set ldflags
KO_CONFIG_PATH=/tmp/ko.yaml
export KO_CONFIG_PATH

# this is where images live
KO_DOCKER_REPO ?= docker.io/bluebrown
export KO_DOCKER_REPO

# never use cgo
CGO_ENABLED=0
export CGO_ENABLED

# the tags are dervied from the version set in the build-meta.txt
# and the revision is the current git revision
include build-meta.txt
tags := $(shell echo $(version) | cut -d'.' -f1) $(shell echo $(version) | cut -d'.' -f1-2) $(version) latest
revision := $(shell git rev-parse main)

# find all commands in this module
cmds := $(wildcard ./cmd/*/)

# used build images locally
loca_bin_dir := ../../bin
buildcmd := go build -ldflags '-w -s -X main.version=$(version) -X main.revision=$(revision)' -o $(loca_bin_dir)/

# generate the image build command, using ko
imgcmd = VERSION="$(version)" REVISION="$(revision)" $(BIN_KO) build --base-import-paths $\
	--image-label org.opencontainers.image.version="$(version)" $\
	--image-label org.opencontainers.image.revision="$(revision)" $\

# run make with push=true to push image
push ?= false
ifneq ($(push), true)
	imgcmd += --local
endif


##@ Commands

.PHONY: help
help: ## Display this help text
	$(loca_bin_dir)/makehelp Makefile

###@ Build

.PHONY: build
build: docs data ## Build the commands
	$(foreach cmd, $(cmds), $(buildcmd) $(cmd);)

.PHONY: image
image: docs data ## Build the container images. (use push=true to push)
	$(foreach cmd,$(cmds), TARGET="$(cmd)" envsubst < ../ko.yaml > $(KO_CONFIG_PATH); \
		$(imgcmd) $(foreach tag,$(tags),--tags $(tag)) $(cmd);)

.PHONY: data
data: ## Run all the scripts in hack.d
	$(foreach script, $(sort $(wildcard hack.d/*)), bash $(script);)

.PHONY: docs
docs: ## Generate the docs.go file from the README.md
	$(BIN_MDTOGO) $(CURDIR) $(CURDIR)/generated/ \
		--license=none --strategy=cmdDocs

###@ Development

.PHONY: lint
lint: ## Lint the code
	$(BIN_LINT) run

.PHONY: test
test: ## Test the code and generate a coverage report
	mkdir -p .coverage
	go test -timeout 30s -coverprofile=./.coverage/profile ./...
	go tool cover -html ./.coverage/profile -o ./.coverage/index.html

.PHONY: view-cov
view-cov: test ## View the coverage report on http://localhost:8080
	docker run --rm -p 127.0.0.1:8080:8080 \
	-v "$(CURDIR)/.coverage/index.html:/usr/share/nginx/html/index.html:ro" \
	nginxinc/nginx-unprivileged

.PHONY: kustomize
kustomize: build ## Build and test with kustomize
	kustomize build --enable-alpha-plugins --enable-exec examples/kustomize/

.PHONY: kpt
kpt: build ## Build and test with kpt
	kpt fn render --allow-exec examples/kpt/ --truncate-output=false

.PHONY: yaml-lint
yaml-lint: ## Lint yaml files
	$(BIN_YAMLFMT) -lint

.PHONY: yaml-fmt
yaml-fmt: ## Format yaml files
	$(BIN_YAMLFMT)

