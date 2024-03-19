#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(realpath $(dirname $0)/..)"

VERSION=$("$PROJECT_ROOT/hack/get-version.sh")
COMPONENT_REGISTRY="eu.gcr.io/sap-gcp-cp-k8s-stable-hub/landscaper"
COMPONENT_REGISTRY_DEV="eu.gcr.io/gardener-project/development"

if [[ -z ${LOCALBIN:-} ]]; then
  LOCALBIN="$PROJECT_ROOT/bin"
fi
if [[ -z ${OCM:-} ]]; then
  OCM="$LOCALBIN/ocm"
fi

overwrite=""
if [[ -n ${OVERWRITE_COMPONENTS:-} ]] && [[ ${OVERWRITE_COMPONENTS} != "false" ]]; then
  overwrite="--overwrite"
fi

echo "> Uploading Component Descriptors to $COMPONENT_REGISTRY ..."
$OCM transfer componentversions "$PROJECT_ROOT/components" "$COMPONENT_REGISTRY" $overwrite

echo "> Uploading Component Descriptors to $COMPONENT_REGISTRY_DEV ..."
$OCM transfer componentversions "$PROJECT_ROOT/components" "$COMPONENT_REGISTRY_DEV" $overwrite
