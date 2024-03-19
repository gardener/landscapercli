#!/bin/bash
#
# Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

PROJECT_ROOT="$(realpath $(dirname $0)/..)"

if [[ -z ${LOCALBIN:-} ]]; then
  LOCALBIN="$PROJECT_ROOT/bin"
fi
if [[ -z ${LINTER:-} ]]; then
  LINTER="$LOCALBIN/golangci-lint"
fi

GOLANGCI_LINT_CONFIG_FILE=""

lsc_module_path=()
for arg in "$@"; do
  case $arg in
    --golangci-lint-config=*)
      GOLANGCI_LINT_CONFIG_FILE="-c ${arg#*=}"
      shift
      ;;
    *)
      lsc_module_path+=("./$(realpath "--relative-base=$PROJECT_ROOT" "$arg")")
      ;;
  esac
done

echo "> Check"

(
  cd "$PROJECT_ROOT"
  echo "  Executing golangci-lint"
  "$LINTER" run $GOLANGCI_LINT_CONFIG_FILE --timeout 10m "${lsc_module_path[@]}"
  echo "  Executing go vet"
  go vet "${lsc_module_path[@]}"
)

if [[ ${SKIP_FORMATTING_CHECK:-"false"} == "false" ]]; then
  echo "Checking formatting"
  "$PROJECT_ROOT/hack/format.sh" --verify "$@"
fi

echo "All checks successful"