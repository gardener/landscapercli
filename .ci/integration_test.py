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
print(f"current environment: {os.environ}")
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

kubectl_client = kubectl.KubectlClient()
if kubectl_client.int_test_tools_dir:
    os.environ['PATH'] = f"{kubectl_client.int_test_tools_dir}:{os.environ['PATH']}"
print(f"'kubectl version' from python")
kubectl_client.version()
print(f"'kubectl version' from PATH={os.environ['PATH']}")
kubectl_version = run(["kubectl", "version", "--client"])

os.chdir(os.path.join(root_path, source_path, "integration-test"))

factory = ctx().cfg_factory()
print(f"Getting kubeconfig for {target_cluster}")
# landscape_kubeconfig = factory.kubernetes(target_cluster)
# see https://github.com/gardener/gardener/blob/master/docs/usage/shoot_access.md#shootsadminkubeconfig-subresource
expiration_seconds = 86400 # is 1 day
landscape_kubeconfig = None
with tempfile.TemporaryDirectory() as tmpdir:
    namespace = "garden-laas"
    service_account = factory.kubernetes("laas-integration-test-service-account")
    service_account_kubeconfig_path = os.path.join(tmpdir, 'service_account_kubeconfig')
    print(f'DEBUG garden-laas service_account_kubeconfig_path={service_account_kubeconfig_path}')
    with open(service_account_kubeconfig_path, "w") as file:
        file.write(yaml.safe_dump(service_account.kubeconfig()))

    debug_kubeconfig_string = yaml.safe_dump(service_account.kubeconfig())
    print(f'DEBUG garden-laas service_account.kubeconfig()="{debug_kubeconfig_string[0:20]}..."')
    
    admin_kube_config_request = f'{{"apiVersion": "authentication.gardener.cloud/v1alpha1", "kind": "AdminKubeconfigRequest", "spec": {{"expirationSeconds": {expiration_seconds}}}}}'

    adminKubeconfigRequest = 'AdminKubeconfigRequest.json'
    with open(adminKubeconfigRequest, "w") as file:
        file.write(admin_kube_config_request)

    print(f'Getting shoots/adminkubeconfig subresource for "{target_cluster}" in namespace "{namespace}"')
    # kubectl --kubeconfig=/tmp/tmpsmkavxfu/service_account_kubeconfig create --raw /apis/core.gardener.cloud/v1beta1/namespaces/garden-laas/shoots/landscapercli-pr/adminkubeconfig -f AdminKubeconfigRequest.json
    command = ["kubectl", "create", f"--kubeconfig={service_account_kubeconfig_path}", "--raw",
               f"/apis/core.gardener.cloud/v1beta1/namespaces/{namespace}/shoots/{target_cluster}/adminkubeconfig",
               "-f", "AdminKubeconfigRequest.json"]
    # CompletedProcess(args=['kubectl', '--kubeconfig=/tmp/tmpl9cwh_p8/service_account_kubeconfig', 'create', '--raw', '/apis/core.gardener.cloud/v1beta1/namespaces/garden-laas/shoots/landscapercli-pr/adminkubeconfig', '-f', 'AdminKubeconfigRequest.json'], returncode=0)

    print(f'  DEBUG command="{command}"')
    commandString = f'kubectl create --kubeconfig={service_account_kubeconfig_path} --raw /apis/core.gardener.cloud/v1beta1/namespaces/{namespace}/shoots/{target_cluster}/adminkubeconfig -f {adminKubeconfigRequest}'
    command =  commandString.split(' ')
    print(f'  DEBUG command="{command}"')

    rc = run(command, shell=True, check=True)
    print(f'DEBUG rc=\n{rc}')
    if rc.returncode != 0:
        raise RuntimeError(f'Could not run command "{command}"')
    
    print(f'DEBUG rc.stdout=\n{rc.stdout}')


    print(f'Using kubeclient.execute_command...')

    kubectl_client.execute_command(f'kubectl create --kubeconfig={service_account_kubeconfig_path} --raw /apis/core.gardener.cloud/v1beta1/namespaces/{namespace}/shoots/{target_cluster}/adminkubeconfig -f {adminKubeconfigRequest}', service_account_kubeconfig_path)

    rc_json = json.loads(rc.stdout)
    kubeconfig_bytes = base64.b64decode(rc_json["status"]["kubeconfig"])
    landscape_kubeconfig = kubeconfig_bytes.decode('utf-8')

if landscape_kubeconfig == None:
    raise RuntimeError(f'Error getting kubeconfig for "{target_cluster}" in namespace "{namespace}"')
                       
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