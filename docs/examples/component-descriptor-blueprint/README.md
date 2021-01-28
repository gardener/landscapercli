# Echo Server Example Component Descriptor

This examples contains an installation with an inline blueprint.
## Steps
1. Download helm chart 
```
helm pull --untar https://storage.googleapis.com/sap-hub-test/echo-server-1.1.0.tgz
```
2. Create a Port Forwarding to the OCI registry
(not required if a publicly exposed OCI registry is used)
```
kubectl port-forward oci-registry-<pod-id> 5000:5000
```

3. Save and upload the helm chart to the OCI registry
```
export HELM_EXPERIMENTAL_OCI=1
helm chart save echo-server localhost:5000/echo-server-chart:v1.1.0
helm chart push localhost:5000/echo-server-chart:v1.1.0
```
4. Upload Blueprint to the OCI registry
```
landscaper-cli blueprints push localhost:5000/echo-server-blueprint:v0.1.0 ./component-descriptor-blueprint/blueprint
```

5. Upload Component Descriptor to the OCI registry
```
landscaper-cli components-cli remote push localhost:5000/components github.com/gardener/echo-server-cd v0.1.0 ./component-descriptor-blueprint
```

6. Apply the target to your cluster

Adapt the `target.yaml` to contain the kubeconfig of your target cluster.
The target cluster can be any kubernetes cluster (including the same cluster).
```
kubectl apply -f ../target.yaml
```
7. Create a target namespace on target cluster
You need a namespace, in which the echo-server will be deployt. If it does not exist, create it (make sure to switch the kubectl config to the target server). If you choose another namespace, modify the `installation.yaml`in the next step.

```
kubectl create namespace landscaper-example
```

9. Apply the installation for the echo server.

Exchange the `<base url oci registry>` placeholder in the installation.yaml file. If you use the OCI registry installed 
with the `quickstart install`, the url is in the console output and follows the 
schema `oci-registry.<namespace>.svc.cluster.local` (here namespace is the one where the oci registry runs).

If you have edited the namespace where the echo-server should be deployed, change the 
`spec.importDataMappings.appnamespace`.

Then apply the installation:
```
kubectl apply -f installation.yaml
```

10. Wait for the echo-server to run.

11. Port forward the echo server
```
kubectl port-forward echo-server-<pod-id> 8080:8080

```
12. Test the echo server with a POST request
```
curl -d "Hello" localhost:8080
```
