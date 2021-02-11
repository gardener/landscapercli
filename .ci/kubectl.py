import os
import requests
import subprocess
import tempfile

def ensure_kubectl_binary(f: callable):
    def wrapper(kubectl_client: 'KubectlClient', *args, **kwargs):
        kubectl_client._get_kubectl_binary()
        return f(kubectl_client, *args, **kwargs)
    return wrapper

def test_kubectl_binary(execPath):
    try:
        command = [execPath, 'version']
        result = subprocess.run(command, capture_output=True, text=True)
        print(f"Test {command} with return code: {result.returncode}")
        return result.returncode == 0
    except OSError:
        return False

class KubectlClient:
    def __init__(self):
        self.kubectl_route = 'https://storage.googleapis.com/kubernetes-release/release/v1.17.0/bin/linux/amd64/kubectl'
        self.bin_path = 'kubectl'
        if not test_kubectl_binary(self.bin_path):
            tempdir = tempfile.gettempdir()
            print(f"kubectl not found in path, installing it to {tempdir}")
            self.bin_path = f"{tempdir}/kubectl"

    def _get_kubectl_binary(self):
        if os.path.isabs(self.bin_path) and not os.path.isfile(self.bin_path):
            res = requests.get(self.kubectl_route, stream=True)
            with open(self.bin_path, mode='wb') as file:
                file.write(res.content)
            os.chmod(self.bin_path, 744)

    @ensure_kubectl_binary
    def apply(self, path_to_file: str, path_to_kubeconfig: str, uninstall = False, forceUninstall = False):
        if uninstall:
            command = [self.bin_path, 'delete', '-f', path_to_file, '--kubeconfig', path_to_kubeconfig, "--ignore-not-found=true"]
            if forceUninstall:
                command.extend(["--grace-period=0", "--force=true", "--wait=false"])
        else:
            command = [self.bin_path, 'apply', '-f', path_to_file, '--kubeconfig', path_to_kubeconfig]
        print(f"  Run {' '.join(command)}")
        result = subprocess.run(command, capture_output=True, text=True)
        print(result.stdout)
        print(result.stderr)
        result.check_returncode()

    @ensure_kubectl_binary
    def apply_string(self, yamlStr, path_to_kubeconfig: str, uninstall = False, forceUninstall = False):
        if uninstall:
            command = [self.bin_path, 'delete', '-f', '-', '--kubeconfig', path_to_kubeconfig, "--ignore-not-found=true"]
            if forceUninstall:
                command.extend(["--grace-period=0", "--force=true", "--wait=false"])
        else:
            command = [self.bin_path, 'apply', '-f', '-', '--kubeconfig', path_to_kubeconfig]
        print(f"  Run {' '.join(command)}")
        result = subprocess.run(command, input=yamlStr, capture_output=True, text=True)
        print(result.stdout)
        print(result.stderr)
        result.check_returncode()

    @ensure_kubectl_binary
    def execute_command(self, commandArg, path_to_kubeconfig: str):
        command = [self.bin_path]
        if isinstance(commandArg, str):
            command.append(commandArg)
        else:
            command.extend(commandArg)
        command.extend(['--kubeconfig', path_to_kubeconfig])
        print(f"  Run {' '.join(command)}")
        result = subprocess.run(command, capture_output=True, text=True)
        print(result.stdout)
        print(result.stderr)
        result.check_returncode()
