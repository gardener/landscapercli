## Commands

### Create local blueprint and component descriptor skeletons

```
landscaper-cli blueprint create path-to-directory component-name component-version
```

Example:

```
landscaper-cli blueprint create ./01-step github.com/gardener/landscapercli/nginx v0.1.0
```

Result could be found [here](./01-step).

### Add helm deploy item from some helm chart repo

In this example the helm chart is located at some external helm chart repo. The helm chart is downloaded 
for later incorporation into the component descriptor as an OCI blob.

```
landscaper-cli blueprint add-deploy-item landscaper.gardener.cloud/helm --name=ingress-nginx  \
  --chart-repo=https://kubernetes.github.io/ingress-nginx \
  --chart=ingress-nginx:3.20.1 \
  --target-ns=testnamespace
```

Missing for adding helm chart:
- additional chart values 
- export values

Next steps:
- Command to create target resource from kubeconfig
- Command to create input values
- Command to validate blueprint with input values (there exists already a blueprint validate/render command)
- Command to generate local installation with input values
- Command to upload all stuff to OCI registry



