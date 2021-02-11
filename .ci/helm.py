import os
import requests
import subprocess
import tarfile
import tempfile

def ensure_helm_binary(f: callable):
    def wrapper(helm_client: 'HelmClient', *args, **kwargs):
        helm_client._get_helm_binary_stream()
        return f(helm_client, *args, **kwargs)
    return wrapper

def test_helm_binary(execPath):
    try:
        command = [execPath, 'version']
        result = subprocess.run(command, capture_output=True, text=True)
        print(f"Test {command} with return code: {result.returncode}")
        return result.returncode == 0
    except OSError:
        return False

class HelmClient:
    def __init__(self):
        self.helm_route = 'https://get.helm.sh/helm-v3.3.4-linux-amd64.tar.gz'
        self.bin_path = 'helm'
        if not test_helm_binary(self.bin_path):
            tempdir = tempfile.gettempdir()
            self.int_test_tools_dir = f"{tempdir}/int-test-tools"
            if not os.path.exists(self.int_test_tools_dir):
                os.makedirs(self.int_test_tools_dir)
            print(f"helm not found in path, installing it to {self.int_test_tools_dir}")
            self.bin_path = f"{self.int_test_tools_dir}/helm"
        else:
            self.int_test_tools_dir = ""

    def _get_helm_binary_stream(self):
        if os.path.isabs(self.bin_path) and not os.path.isfile(self.bin_path):
            res = requests.get(self.helm_route, stream=True)
            with tarfile.open(fileobj=res.raw, mode='r|*') as tar:
                res.raw.seekable = False
                for member in tar:
                    if not member.name == 'linux-amd64/helm':
                        continue

                    fileobj = tar.extractfile(member)
                    with open(self.bin_path, "wb") as outfile:
                        outfile.write(fileobj.read())
                    os.chmod(self.bin_path, 744)
