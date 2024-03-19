#!/bin/bash
#
# Copyright (c) 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

PROJECT_ROOT="$(realpath $(dirname $0)/..)"

GITTREESTATE=`([ -z "$(git status --porcelain 2>/dev/null)" ] && echo clean || echo dirty)`


go install \
  -ldflags "-X github.com/gardener/landscapercli/pkg/version.LandscaperCliVersion=$EFFECTIVE_VERSION \
            -X github.com/gardener/landscapercli/pkg/version.ComponentCliVersion=$COMPONENT_CLI_VERSION \
            -X github.com/gardener/landscapercli/pkg/version.gitTreeState=$GITTREESTATE \
            -X github.com/gardener/landscapercli/pkg/version.gitCommit=$(git rev-parse --verify HEAD)" \
  ${PROJECT_ROOT}/landscaper-cli
