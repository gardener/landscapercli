## Commands

This scenario describes how to create a component containing helm charts whereby the helm charts
are stored in an OCI registry as OCI artifacts.

### Create local blueprint and component descriptor skeletons

```
landscaper-cli blueprint create path-to-directory component-name component-version
```

Example:

```
landscaper-cli blueprint create ./01-step github.com/gardener/landscapercli/nginx v0.1.0
```

Result could be found [here](./01-step).

### Step 2a: Add helm deploy item from some helm chart repo

In this example the helm chart is located as an OCI artifact in an OCI registry. The downloaded 
chart could be found [here](../02-create-blueprint-with-local-helm-charts/02-step/chart)

If the resulting blueprint component should be transportable it is recommended to list

```
landscaper-cli blueprint add-deploy-item landscaper.gardener.cloud/helm path-to-directory \
  --name ingress-nginx \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --chart-version v0.1.0
```

Result could be found [here](./02a-step).

### Step 2b: Add fixed values for helm templating

Helm chart allow to specify values used for templating before the application is deployed. As 
a component developer you want to add some of them with a fixed value and others which could
be set by later user of the blueprint. Here we describe how to add fixed values. 

If you want to add a helm deploy item with some fixed values 

```
landscaper-cli blueprint add-deploy-item landscaper.gardener.cloud/helm path-to-directory \
  --name ingress-nginx \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --chart-version v0.1.0 \
  --value valueName=value \
  --from-file path-to-value-yaml1\
```

The flags --value and --from-file could be used several times.

Assume a values.yaml containing:

``` yaml
controller:
  readinessProbe:
    failureThreshold: 3
    initialDelaySeconds: 10
```

The result of applying the following command to our example [here](./01-step) adds some values to the 
Values section of the corresponding deploy item.

```
landscaper-cli blueprint landscaper.gardener.cloud/helm add-values path-to-directory \
  --name ingress-nginx \
  --value controller.livenessProbe.failureThreshold=5
  --from-file ./values.yaml
```

The result could be found [here](./02b-step)

### Step 2c: Add configurable values for helm templating 

To allow a user of blueprint to specify particular values for helm templating you need import parameters
for the blueprint. You could specify these as follows:

```
landscaper-cli blueprint add-deploy-item landscaper.gardener.cloud/helm path-to-directory \
  --name ingress-nginx \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --chart-version v0.1.0 \
  --import-param=parameterName=valuePath
  --import-type=parameterName=string|int|...
```

The flag --import-param could be used several times. By default, the type of a parameter is string.
If you want to overwrite this you could do this with the flag --import-type.

The result of applying the following command to our example [here](./01-step) adds an import parameter 
and connects if in the values section of the corresponding deploy item.

```
landscaper-cli blueprint landscaper.gardener.cloud/helm add-values path-to-directory \
  --name ingress-nginx \
  --import-param=failurethreshold=controller.readinessProbe.failureThreshold
  --import-type=failurethreshold=int
```

The result could be found [here](./02c-step)

## Step 2d: Export parameter for helm deploy items

If you want to export values of the values.yaml of a helm deploy item you could also specify 
when you add the deploy item:

```
landscaper-cli blueprint add-deploy-item landscaper.gardener.cloud/helm path-to-directory \
  --name ingress-nginx \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --chart-version v0.1.0 \
  --export-param=parameterName=valuePath
  --export-type=parameterName=string|int|...
  --export-from-resource=apiVersion,kind,name,namespace
```

The flag --export-param could be used several times. By default, the type of a parameter is string.
If you want to overwrite this you could do this with the flag --export-type. By default the 
export value is fetched from the values.yaml. If it should be fetched from a rendered resource this
has to be specified via --export-from-resource. 

The result of applying the following command to our example [here](./01-step) adds an export parameter
and connects if in the values section of the corresponding deploy item.

```
landscaper-cli blueprint landscaper.gardener.cloud/helm add-values path-to-directory \
  --name ingress-nginx \
  --export-param=ingressClass=controller.ingressClass
  --export-type=ingressClass=int
```

The result could be found [here](./02d-step)

## Step 2e: Specify images referenced in the helm chart

The images used by a helm chart must be part of the component descriptor of the blueprint. Only then,
all used resourced are declared and could be collected during transport. Furthermore, the helm chart
must explicitly specify all used images in its values.yaml. Images specified somewhere else are hidden
for transport.

The used images could be specified when adding a helm deploy item as follows:

```
landscaper-cli blueprint add-deploy-item landscaper.gardener.cloud/helm path-to-directory \
  --name ingress-nginx \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --chart-version v0.1.0 \
  --oci-image=name=oci-ref
  --oci-image-version=name=someversion
  --oci-value-path=name=somepath
```

For our example helm chart we add two images explicitly though there are more specified in its values yaml.

```
landscaper-cli blueprint add-deploy-item landscaper.gardener.cloud/helm path-to-directory \
  --name ingress-nginx \
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
  --chart-version v0.1.0 \
  --oci-image=controller=k8s.gcr.io/ingress-nginx/controller:v0.43.0
  --oci-image-version=controller=v0.43.0
  --oci-value-path=controller=.controller.image
  --oci-image=certgen=docker.io/jettech/kube-webhook-certgen:v1.5.0
  --oci-image-version=certgen=v1.5.0
  --oci-value-path=certgen=.controller.admissionWebhooks.patch.image
```

The result could be found [here](./02e-step)

## Todo

Missing for adding helm chart:

- Describe logically what is achieved and not only referencing some strange directories
- Discuss if staps like adding values, parameters etc. should be done after generation
  - this makes implementation much more complicated because the deploy items are not valid yaml
    but strings with template instructions

Next steps:
- Command to create target resource from kubeconfig
- Command to validate blueprint with input values (there exists already a blueprint validate/render command)
- Command to generate local installation with input values
- Command to upload all stuff to OCI registry




