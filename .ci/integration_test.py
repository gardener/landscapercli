#!/usr/bin/env python3
import helm
import kubectl
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
namespace = os.environ['NAMESPACE']

try:
    # env var is implicitly set by the output dir in case of a release job
    integration_test_path = os.environ["INTEGRATION_TEST_PATH"]
except KeyError:
    print("Output dir env var not set. " +
          "The output of the integration test won't be saved in a file.")

golang_found = shutil.which("go")
if golang_found:
    print(f"Found go compiler in {golang_found}")
else:
    print("No Go compiler found, installing Go")
    command = ['apk', 'add', 'go', '--no-progress']
    result = run(command)
    result.check_returncode()

if os.path.isfile("/bin/helm3"):
    os.environ['HELM_EXECUTABLE'] = "/bin/helm3"
else:
    # ensure latest helm version
    helm_client = helm.HelmClient()
    print(f"Helm was installed to path '{helm_client.bin_path}'")
    os.environ['HELM_EXECUTABLE'] = helm_client.bin_path
    os.environ['PATH'] = f"{helm_client.int_test_tools_dir}:{os.environ['PATH']}"
print(f"'helm version' PATH={os.environ['PATH']}")
helm_version = run([os.environ['HELM_EXECUTABLE'], "version"])

kubectl_client = kubectl.KubectlClient()
if kubectl_client.int_test_tools_dir:
    os.environ['PATH'] = f"{kubectl_client.int_test_tools_dir}:{os.environ['PATH']}"
print(f"-- List of directory {kubectl_client.int_test_tools_dir}")
from pathlib import Path
print(*Path(kubectl_client.int_test_tools_dir).iterdir(), sep="\n")
print(f"-- End List of directory {kubectl_client.int_test_tools_dir}")
print(f"'kubectl version' PATH={os.environ['PATH']}")
kubectl_version = run(["kubectl", "version"])
    
os.chdir(os.path.join(root_path, source_path, "integration-test"))

factory = ctx().cfg_factory()
landscape_name = "hub-" + landscape + "-test"
landscape_kubeconfig = factory.kubernetes(landscape_name)

with utils.TempFileAuto(prefix="landscape_kubeconfig_") as temp_file:
    temp_file.write(yaml.safe_dump(landscape_kubeconfig.kubeconfig()))
    landscape_kubeconfig_path = temp_file.switch()

    command = ["go", "run", "main.go",
            "--kubeconfig", landscape_kubeconfig_path,
            "--landscaper-namespace", "lndscpr-int-test",
            "--test-namespace", "ls-cli-inttest",
            "--max-retries", "10"]

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