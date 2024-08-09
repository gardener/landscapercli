#!/bin/bash
#
# SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

PROJECT_ROOT="$(realpath $(dirname $0)/..)"

if [[ -z ${EFFECTIVE_VERSION:-} ]]; then
  EFFECTIVE_VERSION=$(cat $PROJECT_ROOT/VERSION)
fi

(
  cd "$PROJECT_ROOT"
  mkdir -p dist

  build_matrix=("linux,amd64" "darwin,amd64" "darwin,arm64", "windows,amd64")

  for i in "${build_matrix[@]}"; do
    IFS=',' read os arch <<< "${i}"

    echo "Build $os/$arch"
    bin_path="dist/landscapercli-$os-$arch"

    CGO_ENABLED=0 GOOS=$os GOARCH=$arch GO111MODULE=on \
    go build -o $bin_path \
    -ldflags "-s -w \
              -X github.com/gardener/landscapercli/pkg/version.LandscaperCliVersion=$EFFECTIVE_VERSION \
              -X github.com/gardener/landscapercli/pkg/version.gitTreeState=$([ -z git status --porcelain 2>/dev/null ] && echo clean || echo dirty) \
              -X github.com/gardener/landscapercli/pkg/version.gitCommit=$(git rev-parse --verify HEAD)" \
    ${PROJECT_ROOT}/landscaper-cli

    # create zipped file
    gzip -f -k "$bin_path"
  done
)
