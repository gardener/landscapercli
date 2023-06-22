#!/usr/bin/env python3
import base64
import json
import subprocess
import tempfile
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
print(f"INFO current environment: {os.environ}")
source_path = os.environ['SOURCE_PATH']
root_path = os.getcwd()
target_cluster = os.environ['TARGET_CLUSTER']

try:
    # env var is implicitly set by the output dir in case of a release job
    integration_test_path = os.environ["INTEGRATION_TEST_PATH"]
except KeyError:
    print("Output dir env var not set. " +
          "The output of the integration test won't be saved in a file.")

golang_found = shutil.which("go")
if not golang_found:
    print(f"Go compiler not found")
    sys.exit(1)

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

print(f"INFO 'kubectl version' from PATH='{os.environ['PATH']}")
kubectl_version = run(["kubectl", "version", "--client"])

os.chdir(os.path.join(root_path, source_path, "integration-test"))

factory = ctx().cfg_factory()
print(f"Getting kubeconfig for {target_cluster}")
# landscape_kubeconfig = factory.kubernetes(target_cluster)
# see https://github.com/gardener/gardener/blob/master/docs/usage/shoot_access.md#shootsadminkubeconfig-subresource
expiration_seconds = 86400 # is 1 day
landscape_kubeconfig_json = None
with tempfile.TemporaryDirectory() as tmpdir:
    namespace = "garden-laas"
    service_account = factory.kubernetes("laas-integration-test-service-account")
    service_account_kubeconfig_path = os.path.join(tmpdir, 'service_account_kubeconfig')
    print(f"DEBUG garden-laas service_account_kubeconfig_path={service_account_kubeconfig_path}")
    with open(service_account_kubeconfig_path, "w") as file:
        file.write(yaml.safe_dump(service_account.kubeconfig()))

    debug_kubeconfig_string = yaml.safe_dump(service_account.kubeconfig())
    
    admin_kube_config_request = f'{{"apiVersion": "authentication.gardener.cloud/v1alpha1", "kind": "AdminKubeconfigRequest", "spec": {{"expirationSeconds": {expiration_seconds}}}}}'

    adminKubeconfigRequest = 'AdminKubeconfigRequest.json'
    with open(adminKubeconfigRequest, "w") as file:
        file.write(admin_kube_config_request)

    print(f"Getting shoots/adminkubeconfig subresource for '{target_cluster}' in namespace '{namespace}'")
    commandString = f"kubectl create --kubeconfig={service_account_kubeconfig_path} --raw /apis/core.gardener.cloud/v1beta1/namespaces/{namespace}/shoots/{target_cluster}/adminkubeconfig -f {adminKubeconfigRequest}"
    rc = subprocess.run(commandString, shell=True, check=True, capture_output=True, encoding='utf-8')

    if rc.returncode != 0:
        raise RuntimeError(f"Could not run command '{commandString}'")
    
    rc_json = json.loads(rc.stdout)
    kubeconfig_bytes = base64.b64decode(rc_json["status"]["kubeconfig"])
    landscape_kubeconfig_json = yaml.safe_load(kubeconfig_bytes.decode('utf-8'))

if landscape_kubeconfig_json == None:
    raise RuntimeError(f"Error getting kubeconfig for '{target_cluster}' in namespace '{namespace}'")
                       
with utils.TempFileAuto(prefix="landscape_kubeconfig_") as temp_file:
    json.dump(landscape_kubeconfig_json, temp_file)
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
    