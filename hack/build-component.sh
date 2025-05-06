#!/bin/bash
#
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors.
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

PROJECT_ROOT="$(realpath $(dirname $0)/..)"
HACK_DIR="$PROJECT_ROOT/hack"

VERSION=$("$HACK_DIR/get-version.sh")

if [[ -z ${LOCALBIN:-} ]]; then
  LOCALBIN="$PROJECT_ROOT/bin"
fi
if [[ -z ${OCM:-} ]]; then
  OCM="$LOCALBIN/ocm"
fi

if [[ -z ${CDVERSION:-} ]]; then
  CDVERSION=$VERSION
fi

echo "> Building component in version $CDVERSION"
$OCM add componentversions --file "$PROJECT_ROOT/components" --version "$CDVERSION" --create --force "$PROJECT_ROOT/.landscaper/components.yaml" -- \
  CDVERSION="$CDVERSION" \
  VERSION="$VERSION"
