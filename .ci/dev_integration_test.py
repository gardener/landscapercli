#!/usr/bin/env python3

import os
import subprocess
import sys

sys_args = sys.argv
root_path = os.getcwd()
source_path = os.environ['SOURCE_PATH']

# parse sys var(will be passed to subprocess calls)
for i in range(len(sys_args)):
    if '--garden-namespace ' in sys_args[i]:
        os.environ['GARDEN_NAMESPACE'] = sys_args[i].split(' ', 2)[1]
    if "--target-clustername" in sys_args[i]:
        os.environ['TARGET_CLUSTER'] = sys_args[i].split(' ', 2)[1]
    if "--test-namespace " in sys_args[i]:
        os.environ['TEST_NAMESPACE'] = sys_args[i].split(' ', 2)[1]

os.environ['ROOT_PATH'] = root_path
os.environ['LANDSCAPE'] = "dev"
os.environ['NAMESPACE'] = "landscapercli-release-test"


hub_kubeconfig = os.path.join(
    root_path, source_path,
    ".ci",
    "integration_test.py"
)

command = [hub_kubeconfig, "--namespace", "release-test"]

result = subprocess.run(command)
result.check_returncode()
