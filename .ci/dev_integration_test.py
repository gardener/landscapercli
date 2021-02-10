#!/usr/bin/env python3

import os
import subprocess
import sys

sys_args = sys.argv
root_path = os.getcwd()
source_path = os.environ['SOURCE_PATH']

# parse sys var(will be passed to subprocess calls)
os.environ['ROOT_PATH'] = root_path
os.environ['LANDSCAPE'] = "dev"
os.environ['NAMESPACE'] = "landscapercli-release-test"
os.environ['HELM_V3_VERSION'] = "v3.2.4"


hub_kubeconfig = os.path.join(
    root_path, source_path,
    ".ci",
    "integration_test.py"
)

command = [hub_kubeconfig, "--namespace", "release-test"]

result = subprocess.run(command)
result.check_returncode()
