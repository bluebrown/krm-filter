SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

BINDIR = bin
BIN_YAMLFMT = go run github.com/google/yamlfmt/cmd/yamlfmt -conf=.yamlfmt

##@ Options

cmd ?= build
push ?= false


##@ Commands

.PHONY: help
help: ## Display this help text
	bin/makehelp $(MAKEFILE_LIST)

%: ## run the cmd against the named filter
	$(MAKE) $(cmd) -C filter/$* push=$(push)

.PHONY: all
all: ## run the cmd against all filters
	@$(foreach filter, $(wildcard filter/*/), $(MAKE) $(cmd) -C $(filter) push=$(push);)

.PHONY: clean
clean: ## clean up the local bin dir
	@rm -f $(filter-out $(BINDIR)/init, $(wildcard $(BINDIR)/*))

.PHONY: yaml-lint
yaml-lint: ## Lint yaml files
	$(BIN_YAMLFMT) -lint

.PHONY: yaml-fmt
yaml-fmt: ## Format yaml files
	$(BIN_YAMLFMT)

