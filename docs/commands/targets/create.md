# Creating targets
The general two steps for installing software components from an OCI registry via Landscaper are the following:

1. Creating one or more targets: These contain the access information for the system(s), into which the software components should be deployed.

1. Creating one or more installation objects: These contain the reference to the software component, the configuration of the component via its import and export parameters, and a reference to one or more target objects. More information on this topic can be found [here](../installations/create.md).

> More background information can be found in the [Landscaper docs](https://github.com/gardener/landscaper/blob/master/docs/README.md).

## Creating targets
The command `landscaper-cli targets create [target type]` creates a target object of the specified target type. Depending on the target type, different additional CLI parameters will be required. E.g. when creating a target of type `landscaper.gardener.cloud/kubernetes-cluster` which points to a K8s cluster, the kubeconfig for this cluster must be set via the CLI parameter `--kubeconfig`.

The following command generates a Landscaper target of type `landscaper.gardener.cloud/kubernetes-cluster`, which points to the K8s cluster defined by the `--kubeconfig` parameter:

```
landscaper-cli targets create kubernetes-cluster --name my-target --namespace my-namespace --kubeconfig [path to kubeconfig] > ./my-target.yaml
```

The command output is written to `./my-target.yaml`:

```yaml
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: Target
metadata:
  name: my-target
  namespace: my-namespace
spec:
  type: landscaper.gardener.cloud/kubernetes-cluster
  config:
    kubeconfig: |
      ...
```

The generated target can be applied onto the cluster via `kubectl apply -f ./my-target.yaml` and then be referenced in installation objects.