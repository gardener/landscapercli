#!/usr/bin/env python3

import os
from subprocess import Popen, PIPE, STDOUT, run
import sys
import shutil
import utils
import yaml

from util import ctx

print("Starting integration test")
print(f"current environment: {os.environ}")
source_path = os.environ['SOURCE_PATH']
root_path = os.environ['ROOT_PATH']
landscape = os.environ['LANDSCAPE']
garden_namespace = os.environ['GARDEN_NAMESPACE']
target_cluster = os.environ['TARGET_CLUSTER']
test_namespace = os.environ['TEST_NAMESPACE']

try:
    # env var is implicitly set by the output dir in case of a release job
    integration_test_path = os.environ["INTEGRATION_TEST_PATH"]
except KeyError:
    print("Output dir env var not set. " +
          "The output of the integration test won't be saved in a file.")
factory = ctx().cfg_factory()
landscape_kubeconfig = factory.kubernetes("garden-" + landscape + "-virtual")
landscape_kubeconfig_name = "landscape_kubeconfig"
landscape_kubeconfig_path = os.path.join(root_path, source_path,
                                         "integration-test",
                                         landscape_kubeconfig_name)

utils.write_data(landscape_kubeconfig_path, yaml.dump(
                landscape_kubeconfig.kubeconfig()))

golang_found = shutil.which("go")
if golang_found:
    print(f"Found go compiler in {golang_found}")
else:
    print("No Go compiler found, installing Go")
    command = ['apk', 'add', 'go', '--no-progress']
    result = run(command)
    result.check_returncode()


os.chdir(os.path.join(root_path, source_path, "integration-test"))

command = ["go", "run", "main.go",
           "-garden-kubeconfig", landscape_kubeconfig_path,
           '-landscaperNamespace', garden_namespace,
           '-target-clustername', target_cluster,
           "-testNamespace", test_namespace,
           "-maxRetries", "10"]

print(f"Running integration test with command: {' '.join(command)}")
try:
    # check if path var is set
    integration_test_path
except NameError:
    run = run(command)
else:
    output_path = os.path.join(root_path, integration_test_path, "out")

    with Popen(command, stdout=PIPE, stderr=STDOUT, bufsize=1, universal_newlines=True) as run, open(output_path, 'w') as file:
        for line in run.stdout:
            sys.stdout.write(line)
            file.write(line)

if run.returncode != 0:
    raise EnvironmentError("Integration test exited with errors")