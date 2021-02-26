#!/usr/bin/env python3

import os
import subprocess
import sys

sys_args = sys.argv
root_path = os.getcwd()
source_path = os.environ['SOURCE_PATH']

# parse sys var(will be passed to subprocess calls)
# parse sys var(will be passed to subprocess calls)
for i in range(len(sys_args)):
    if "--target-clustername" in sys_args[i]:
        os.environ['TARGET_CLUSTER'] = sys_args[i].split(' ', 2)[1]

os.environ['ROOT_PATH'] = root_path
os.environ['LANDSCAPE'] = "dev"

int_test_command = os.path.join(
    root_path, source_path,
    ".ci",
    "integration_test.py"
)

command = [int_test_command]

result = subprocess.run(command)
result.check_returncode()
