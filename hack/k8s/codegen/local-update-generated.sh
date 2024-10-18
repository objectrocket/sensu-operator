#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

#SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/../../..
#CODEGEN_PKG=${CODEGEN_PKG:-$(cd ${SCRIPT_ROOT}; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
#${CODEGEN_PKG}/kube_codegen.sh all \
#  pkg/generated \
#  pkg/apis \
#  objectrocket:v1beta1 \
#  --go-header-file ${SCRIPT_ROOT}/hack/k8s/codegen/boilerplate.go.txt \
#  "$@"

# Example for future version use
# ingress:v1alpha1,v1alpha2 \


#!/bin/bash

set -e
#set -o errexit
#set -o nounset
#set -o pipefail

# Define your project paths
SCRIPT_ROOT=$(dirname ${BASH_SOURCE})/..
echo $SCRIPT_ROOT
CODEGEN_PKG=../code-generator
echo $CODEGEN_PKG
output_base=$(dirname ${BASH_SOURCE})/../../../pkg/generated/
ls $SCRIPT_ROOT
echo "ouput: "$output_base
ls ${output_base}
ls $CODEGEN_PKG
# Generate the clientset, listers, and informers
source "${CODEGEN_PKG}/kube_codegen.sh"

kube::codegen::gen_helpers \
    --boilerplate "${SCRIPT_ROOT}/codegen/boilerplate.go.txt" \
    "pkg/apis" \

kube::codegen::gen_client \
    --boilerplate "${SCRIPT_ROOT}/codegen/boilerplate.go.txt" \
    --with-watch \
    --output-dir ${output_base} \
    --output-pkg "github.com/objectrocket/sensu-operator/pkg/generated" \
    "pkg/apis"

kube::codegen::gen_openapi \
    --boilerplate "${SCRIPT_ROOT}/codegen/boilerplate.go.txt" \
    --output-dir ${output_base} \
    --output-pkg "github.com/objectrocket/sensu-operator/pkg/generated" \
    "pkg/apis"

#echo "${CODEGEN_PKG}/kube_codegen.sh" \
#  "clientset,informers,listers" \
#  ${output_base} \
#  github.com/objectrocket/sensu-operator/pkgs/apis \
#  "objectrocket:v1beta1" \
#  --go-header-file ${SCRIPT_ROOT}/codegen/boilerplate.go.txt \
#  --v=2

echo "Done"

