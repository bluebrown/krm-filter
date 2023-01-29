SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

##@ Options

dir_build ?= build
dir_deploy ?= deploy


##@ Commands

.PHONY: help
help: ## Display this help text
	@../../bin/makehelp $(MAKEFILE_LIST)

image:
	@$(MAKE) -C ../../ all
	@docker build -f $(dir_build)/Dockerfile -t bluebrown/argocd-cmp-krm:v0.1.0 ../../

kind: image
	@kind create cluster
	@kind load docker-image bluebrown/argocd-cmp-krm:v0.1.0
	@kubectl rollout restart deploy/argocd-repo-server -n argocd || true
	@kubectl create ns argocd || true


kzcmd := kustomize build --enable-alpha-plugins --enable-exec $(dir_deploy)

config: $(dir_deploy)/components/aad-token/etc/aad-token.env $(dir_deploy)/localized-argocd
	@$(kzcmd)

.PHONY: deploy
deploy: $(dir_deploy)/components/aad-token/etc/aad-token.env $(dir_deploy)/localized-argocd
	@$(kzcmd) | kubectl apply -n argocd -f -

.PHONY: undeploy
undeploy: $(dir_deploy)/components/aad-token/etc/aad-token.env $(dir_deploy)/localized-argocd
	@$(kzcmd) | kubectl delete -n argocd -f - --ignore-not-found

$(dir_deploy)/components/aad-token/etc/aad-token.env:
	@printf "AAD_ACCESS_TOKEN=%s\n" \
		"$(shell az account get-access-token \
  			--resource https://vault.azure.net \
			--scope https://vault.azure.net/.default \
  			--query accessToken --output tsv)" \
  		> $(dir_deploy)/components/aad-token/etc/aad-token.env

$(dir_deploy)/localized-argocd:
	@kustomize localize \
		$(dir_deploy)/.remote/argocd \
		$(dir_deploy)/localized-argocd
