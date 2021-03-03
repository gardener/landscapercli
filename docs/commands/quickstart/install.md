# Quickstart Install

The quickstart install command allows to install the landscaper (and optionally an OCI registry) in a specified kubernetes cluster, such as a minikube, kind or garden shoot cluster. This is the quickest way to get the landscaper up and running.

## Prerequisites
- K8s cluster
- [Helm 3](https://helm.sh/docs/intro/install/)

## Usage
```
landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --landscaper-values ./landscaper-values.yaml --namespace landscaper --install-oci-registry
```
`landscaper-values.yaml`is the values file of the landscaper. See below for a minimal example.
To install a specific version of the landscaper chart, use the `landscaper-chart-version` argument.

For more details on the cli usage, consult [landscaper-cli_quickstart_install reference](../../reference/landscaper-cli_quickstart_install.md).
If the installation succeeds you can verify the created pods with:

```
kubectl get pods -n landscaper --kubeconfig=./kubconfig.yaml
NAME                            READY   STATUS    RESTARTS   AGE
landscaper-6674fd9c47-9jz2n     1/1     Running   0          78s
oci-registry-6654f55648-klp22   1/1     Running   0          83s
```

### Interact with OCI registry

If you have installed the OCI registry with the quickstart command there is not external https endpoint to access it.
You have two alternatives to continue:

#### Alternative 1: Generate an external endpoint
If you have installed the OCI registry on a Gardener managed k8s shoot cluster with nginx enabled you could execute
the following command to create an external https endpoint. The command requires *htpasswd* installed on your computer.

```
landscaper-cli quickstart add-oci-endpoint \
    --kubeconfig ./kubconfig.yaml \
    --namespace landscaper \
    --user testuser \
    --password some-pw
```


This command creates an external https endpoint with basic authentication. Be aware that this endpoint is accessible
from everywhere. Therefore, use some strong password. 

If the URL of the API server of the k8s cluster is *https://api.cluster-domain* then the endpoint is 
*https://o.ingress.cluster-domain*, e.g. 

```
API-Server: https://api.mycluster.myproject.shoot.live.k8s-hana.ondemand.com
OCI Endpoint: https://o.ingress.mycluster.myproject.shoot.live.k8s-hana.ondemand.com
```

You could check the endpoint with, e.g. the *curl* command:

```
curl --location --request GET https://o.ingress.mycluster.myproject.shoot.live.k8s-hana.ondemand.com/v2/_catalog -u "testuser:some-pw" 
```

#### Alternative 2: Continue with port forwarding

Without an external endpoint for the OCI registry, you have to use port-forwarding. 
You can forward the port 5000 of the registry pod to your localhost with the following command:
```
kubectl port-forward oci-registry-<pod-id> 5000:5000 -n landscaper
```
To check the endpoint then use:
```
curl --location --request GET http://localhost:5000/v2/_catalog
```
This should give the output:

```
{"repositories":[]}
```

indicating an empty registry.

Afterwards, you can use the tools of your choice to push artifacts against the localhost:5000 registry url, e.g. 

TODO: verify special /etc/hosts domain name for docker push

### Landscaper Values

The landscaper values are used during the internal helm install of the landscaper chart. Therefore, all values from the 
chart can be specified. For more options see also [here](https://github.com/gardener/landscaper/blob/master/charts/landscaper/values.yaml)

> ❗ If you use the `--install-oci-registry` flag and you have not configured an external https endpoint for the OCI registry, 
>  set `landscaper.registryConfig.allowPlainHttpRegistries = true`.

A minimum working example goes as follows:
```yaml
landscaper:

  registryConfig: # contains optional oci secrets
    allowPlainHttpRegistries: false
    secrets: {}
#     <name>: <docker config json>

  deployers:
  - container
  - helm
  - manifest
#  - mock

```


