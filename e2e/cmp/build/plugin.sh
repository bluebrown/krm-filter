#!/bin/bash

set -euo pipefail

: "${ANNOTATION_DOMAIN:=argocd.my-org.io}"

kpt fn eval --exec azure-vault-secrets

kpt fn eval --exec refhash

kpt fn eval --exec annotate -- \
  "$ANNOTATION_DOMAIN/app=$ARGOCD_APP_NAME" \
  "$ANNOTATION_DOMAIN/rev=$ARGOCD_APP_REVISION" \
  "$ANNOTATION_DOMAIN/repo=$ARGOCD_APP_SOURCE_REPO_URL" \
  "$ANNOTATION_DOMAIN/branch=$ARGOCD_APP_SOURCE_TARGET_REVISION" \
  "$ANNOTATION_DOMAIN/path=$ARGOCD_APP_SOURCE_PATH"

kpt fn eval --exec remove-local-config-resources --output unwrap
