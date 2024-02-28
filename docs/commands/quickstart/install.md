# Quickstart Install

The quickstart install command allows to install the landscaper (and optionally an OCI registry) in a specified 
kubernetes cluster, such as a minikube, kind or garden shoot cluster. This is the quickest way to get the landscaper up 
and running.

## Prerequisites
- K8s cluster
- [Helm >=3.7.0](https://helm.sh/docs/intro/install/)

## Usage

The simple use case would be to install Landscaper without the internal OCI registry, e.g. if you already have an 
external OCI registry (GCR, Harbor, ...) for usage. This can be achieved by executing the following command:

```
landscaper-cli quickstart install --kubeconfig ./kubeconfig.yaml 
```

This command installs the landscaper in the namespace `landscaper` together with a helm-, manifest- and container-deployer.

If the installation succeeds you can verify the created pods with:

```
kubectl get pods -n landscaper --kubeconfig=./kubconfig.yaml
NAME                                  READY   STATUS    RESTARTS   AGE
container-deployer-58c59cdcdd-thggs   1/1     Running   0          9s
helm-deployer-66fc44f66b-h7262        1/1     Running   0          14s
landscaper-6d4488d86f-fg5dg           1/1     Running   0          16s
landscaper-main-c7d8c75db-ms2d7       1/1     Running   0          16s
landscaper-webhooks-fbc6cd9bb-frhx5   1/1     Running   0          16s
landscaper-webhooks-fbc6cd9bb-lgdgl   1/1     Running   0          16s
manifest-deployer-db4466cc5-g99wm     1/1     Running   0          12s
```

You see the central pods of the landscaper and its webhook as well as one pod for each deployer.  

To install a specific version of the landscaper chart, use the `landscaper-chart-version` argument.

If you want to install the landscaper and the deployer in another namespace use the `namespace` argument.

The default settings for the installation of the Landscaper are usually sufficient. If you want to provide some 
other configuration values, you might call the command in the following form:

```
landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --landscaper-values ./landscaper-values.yaml 
```

`landscaper-values.yaml` is the values file of the landscaper. See [here](#landscaper-values) for a minimal example. 
In the `landscaper-values.yaml` it is not possible to configure any of the deployer, enable the Landscaper agent or
deployer registration. The Landscaper agent and deployer configuration will be removed in the near future and 
particular values files for the different deployer will be introduced later, if required.

For more details on the cli usage, consult [landscaper-cli_quickstart_install reference](../../reference/landscaper-cli_quickstart_install.md).

### Installing and configuring the OCI registry

If you don't want to use an external OCI registry, it is also possible to setup a new registry inside the target cluster 
via the `quickstart install` command. **This registry should only be used for dev/testing purposes.**. There are 2 
possible ways to setup this OCI registry. Depending on which you chose, the interaction with it will be slightly different. 
These two alternatives are described in the following.

#### Alternative 1: Expose registry via ingress (recommended)

**Prerequisites**
- The target cluster must be a Gardener Shoot (TLS is provided via the Gardener cert manager).
- A nginx ingress controller must be deployed in the target cluster
- The command "htpasswd" must be installed on your local machine.

You can setup a Landscaper instance together with the OCI registry that is exposed via an ingress by executing the following command:

```
landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --landscaper-values ./landscaper-values.yaml --namespace landscaper --install-oci-registry --install-registry-ingress --registry-username testuser --registry-password some-pw
```

The OCI registry will be exposed via an external https endpoint with basic authentication. The credentials are defined by the flags `--registry-username` and `--registry-password`. Be aware that this endpoint is accessible from everywhere. Therefore, use strong credentials.

The installed Landscaper instance will be automatically configured with the OCI registry credentials. Therefore, you must not explicitely set the credentials in the Landscaper values or subsequently in the Landscaper installations.

If the command runs through successfully, the ingress URL gets printed out at the end. You can use this URL together with the provided credentials to access the OCI registry. You can also use this URL as the `baseURL` in your Landscaper artifacts.

Just for completeness: If the URL of the API server of the k8s cluster is *https://api.cluster-domain* then the endpoint is 
*https://o.ingress.cluster-domain*, e.g. 

```
API-Server: https://api.mycluster.myproject.shoot.live.k8s-hana.ondemand.com
OCI Endpoint: https://o.ingress.mycluster.myproject.shoot.live.k8s-hana.ondemand.com
```

Once the OCI registry is running, you can check the endpoint with, e.g. the *curl* command:

```
curl --location --request GET https://o.ingress.mycluster.myproject.shoot.live.k8s-hana.ondemand.com/v2/_catalog -u "testuser:some-pw" 
```

#### Alternative 2: Access registry with port forwarding

You can setup a Landscaper instance together with the OCI registry **without** an ingress by executing the following command:

```
landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --landscaper-values ./landscaper-values.yaml --namespace landscaper --install-oci-registry
```

> For this setup to work, `landscaper.registryConfig.allowPlainHttpRegistries` must be set to `true` in the Landscaper values.

Without an external endpoint for the OCI registry, it will only be reachable from within the cluster. Therefore, you have to use port-forwarding for access. You can forward the port 5000 of the registry pod to your localhost with the following command:

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

Afterward, you can use the tools of your choice to push artifacts against the localhost:5000 registry url, e.g. 

TODO: verify special /etc/hosts domain name for docker push

### Landscaper Values

The landscaper values are used during the internal helm install of the landscaper chart. Therefore, all values from the 
chart can be specified. For more options see also [here](https://github.com/gardener/landscaper/blob/master/charts/landscaper/values.yaml)

A minimum working example goes as follows:
```yaml
landscaper:
  landscaper: 
    registryConfig: # contains optional oci secrets
      allowPlainHttpRegistries: false
```


