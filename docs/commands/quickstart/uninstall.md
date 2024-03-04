# Quickstart Install

The `quickstart uninstall` command allows to uninstall the landscaper and OCI registry in a specified kubernetes cluster, 
such as a minikube, kind or garden shoot cluster. It is intended to be used to uninstall the landscaper and OCI registry 
from the `quickstart install` command.

# Prerequisites
- K8s cluster
- Helm 3

# Usage
```
landscaper-cli quickstart uninstall --kubeconfig ./kubconfig.yaml --namespace landscaper
```

For more details on the cli usage, consult [landscaper-cli_quickstart_uninstall reference](../../reference/landscaper-cli_quickstart_uninstall.md).