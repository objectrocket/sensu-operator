#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -x

SCRIPT_ROOT=$(realpath $(dirname ${BASH_SOURCE})/../../..)

${SCRIPT_ROOT}/hack/k8s/codegen/generate-groups.sh all \
github.com/objectrocket/sensu-operator/pkg/generated \
  github.com/objectrocket/sensu-operator/pkg/apis \
  objectrocket:v1beta1 \
  --go-header-file ${SCRIPT_ROOT}/hack/k8s/codegen/boilerplate.go.txt \
"$@"

#docker run --rm \
#  -v "$PWD":"$DOCKER_REPO_ROOT" \
#  -w "$DOCKER_REPO_ROOT" \
#  "$IMAGE" \
#  "/go/src/k8s.io/code-generator/generate-groups.sh"  \
#  "all" \
#  "github.com/objectrocket/sensu-operator/pkg/generated" \
#  "github.com/objectrocket/sensu-operator/pkg/apis" \
#  "objectrocket:v1beta1" \
#  --go-header-file "./hack/k8s/codegen/boilerplate.go.txt" \
#  $@
