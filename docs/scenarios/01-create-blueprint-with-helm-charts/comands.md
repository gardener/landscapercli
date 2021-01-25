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

In this example the helm chart is located as an OCI artifact in an OCI registry. 

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

### ### Step 2c: Add configurable values for helm templating 

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
  --value controller.livenessProbe.failureThreshold=5
  --import-param=failurethreshold=controller.readinessProbe.failureThreshold
  --import-type=failurethreshold=int
```

The result could be found [here](./02c-step)


## Todo

Missing for adding helm chart:

- Describe logically what is achieved and not only referencing some strange directories
- how to add image ref in blueprint and references
- export values

Next steps:
- Command to create target resource from kubeconfig
- Command to create input values
- Command to validate blueprint with input values (there exists already a blueprint validate/render command)
- Command to generate local installation with input values
- Command to upload all stuff to OCI registry




