#!/bin/bash -xe

#  Kubernetes JSON Schema

#  Copyright (C) 2017 Gareth Rushgrove
#  Copyright (C) 2017 Yanh Hamon

#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at

#      https://www.apache.org/licenses/LICENSE-2.0

#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

# NOTE:
# This script has been copied and modified from https://github.com/yannh/kubernetes-json-schema/blob/master/build.sh
# It was modified to only generate schemas for the master branch in strict expanded standalone mode.

: "${K8S_VERSION:=master}"

if test -d "./cmd/kubeconform/kodata/schemas/${K8S_VERSION}-standalone-strict"; then
  printf "data already exists\n"
  exit 0
fi

OPENAPI2JSONSCHEMABIN="docker run --rm -i -u "$(id -u):$(id -g)" -v ${PWD}:/out ghcr.io/yannh/openapi2jsonschema:latest"

SCHEMA=https://raw.githubusercontent.com/kubernetes/kubernetes/${K8S_VERSION}/api/openapi-spec/swagger.json

$OPENAPI2JSONSCHEMABIN -o "cmd/kubeconform/kodata/schemas/${K8S_VERSION}-standalone-strict" --expanded --kubernetes --stand-alone --strict "${SCHEMA}"
