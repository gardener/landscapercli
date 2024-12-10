#!/bin/bash
set -euo pipefail

PROJECT_ROOT="$(realpath $(dirname $0)/..)"

VERSION=$("$PROJECT_ROOT/hack/get-version.sh")
COMPONENT_REGISTRY="europe-docker.pkg.dev/sap-gcp-cp-k8s-stable-hub/landscaper"
COMPONENT_REGISTRY_DEV="europe-docker.pkg.dev/sap-gcp-cp-k8s-stable-hub/development"

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

for reg in "$COMPONENT_REGISTRY" "$COMPONENT_REGISTRY_DEV"; do
  if [[ ${REGISTRY_PREFIX:-} ]] && [[ $reg != ${REGISTRY_PREFIX}* ]]; then
    continue
  fi
  echo "> Uploading Component Descriptors to $reg ..."
  $OCM transfer componentversions "$PROJECT_ROOT/components" "$reg" $overwrite
done
