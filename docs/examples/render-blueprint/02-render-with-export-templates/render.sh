#!/bin/bash
#
# Copyright (c) 2022 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

EXAMPLE_DIR="$(dirname $0)"
BLUEPRINT_DIR=${EXAMPLE_DIR}/blueprint/root

landscaper-cli blueprints render ${BLUEPRINT_DIR} \
    -c ${EXAMPLE_DIR}/component-descriptor/root/component-descriptor.yaml \
    -f ${EXAMPLE_DIR}/values.yaml \
    -r ${EXAMPLE_DIR}/resources.yaml \
    -e ${EXAMPLE_DIR}/export-templates.yaml \
    -a ${EXAMPLE_DIR}/component-descriptor/comp-a/component-descriptor.yaml \
    -a ${EXAMPLE_DIR}/component-descriptor/comp-b/component-descriptor.yaml \
    -a ${EXAMPLE_DIR}/component-descriptor/comp-c/component-descriptor.yaml
