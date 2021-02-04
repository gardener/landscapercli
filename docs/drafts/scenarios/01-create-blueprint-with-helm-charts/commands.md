## Commands

This scenario describes how to create a reusable component to deploy helm charts. It is assumed that the helm charts are already
stored in an OCI registry as OCI helm artifacts.

### Create local blueprint and component descriptor skeletons

The following command creates a component descriptor, a blueprint and a resources.yaml. In the resources.yaml 
all resources are collected which will be added to the component descriptor later when the component is uploaded 
to some OCI registry. This command adds a local reference to the blueprint to the reference.yaml. This means that 
later the blueprint is stored as one layer in the OCI artifact uploaded to some OCI registry.

Command:

```
landscaper-cli component create
    [component directory path]
    [component name]
    [component version]
```

Example:

```
landscaper-cli component create
    .
    github.com/gardener/landscapercli/nginx
    v0.1.0
```

Result could be found [here](./01-step) and has the following file structure inside the component directory:

```
├── blueprint
│      └── blueprint.yaml
└── component-descriptor.yaml
└── resources.yaml
```

The repositoryContext in the component-descriptor is incomplete, because the component-descriptor has not yet been pushed 
into an oci registry, so that the baseUrl is unknown.

The resources.yaml file contains a reference to the blueprint which is later added automatically to the component descriptor
when the component is uploaded to some OCI registry.

### Step 2: Add helm deploy item 

In this example the helm chart is located as an OCI artifact in an OCI registry. The downloaded 
chart could be found [here](../02-create-blueprint-with-local-helm-charts/02-step/chart)

Command:

```
landscaper-cli component add helm-ls deployitem
    [deployitem name]
    --component-directory [directory path to component]
    --oci-reference [oci reference]
    --resource-version [chart-version]
    --cluster-param [cluster-parameter-name]
    --target-ns-param [target-namespace-parameter-name]
```

Example:

```
landscaper-cli component add helm-ls deployitem \
    ingress-nginx \
    --component-directory . \
    --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
    --resource-version v0.1.0 \
    --cluster-param target-cluster \
    --target-ns-param nginx-namespace
```

The reference to the chart is added to the resources.yaml. 

An execution *ingress-nginx* with a deployitem *ingress-nginx* is added to the blueprint.

Import parameters with the specified names are added to the blueprint definition. With these import 
parameter you could later provide the target cluster as well as the target naespace where the nginx helm chart
should be deployed.

Result could be found [here](./02-step).

### Step 3: Add fixed values for helm templating

Helm charts allow specifying values used for templating before the application is deployed. As 
a component developer you want to add some of them with a fixed value.

The following command adds a single key-value pair to the values section of a deploy item.

Command:

```
landscaper-cli blueprint add helm-ls value 
    [component directory path]
    [execution name]
    [deployitem name]
    --jsonpath [path in the values section of the deployitem]
    --value [value]
```

The values section can have a deep structure. Therefore, a key can be a path.

Example:

```
landscaper-cli blueprint add helm-ls value
    .
    ingress-nginx
    ingress-nginx
    --jsonpath .controller.readinessProbe.failureThreshold
    --value 5
```

Values can also be taken from a file. This variant has the advantage that a deep structure can be added with one command.

```
landscaper-cli blueprint add helm-ls value
    [component directory path]
    [execution name]
    [deployitem name]
    --file [path to values file]
```

Example:

Assume a values.yaml containing:

```
controller:
  readinessProbe:
    failureThreshold: 3
    initialDelaySeconds: 10
```

And the command: 

```
landscaper-cli blueprint add helm-ls value
    .
    ingress-nginx
    ingress-nginx
    --file ./values.yaml
```

The commands add the specified values to the values section of the deploy item. The result of the two example
commands could be found [here](./03-step). 

### Step 4: Add configurable values for helm templating 

To allow different users of a blueprint to specify its own values for helm templating you need import parameters
for the blueprint and connect them with entries in the helm values.yaml. You can achieve this with the following command:

```
landscaper-cli blueprint add helm-ls value
    [component directory path]
    [execution name]
    [deployitem name]
    --jsonpath [path in the values section of the deployitem]
    --import-param [import parameter name]
    --type [import parameter type]
```

By default, the type of a parameter is string. If you want to overwrite this you could do this with the flag --type.

Example:

```
landscaper-cli blueprint add helm-ls value
    .
    ingress-nginx
    ingress-nginx
    --jsonpath .controller.readinessProbe.failureThreshold
    --import failurethreshold
    --type int
```

The result of applying the command to our example [here](./01-step) adds an import parameter
and connects if in the values section of the corresponding deploy item.

The result could be found [here](./04-step)

## Step 5: Export parameter for helm deploy items

If you want to export values of the values.yaml or of the rendered manifest of a helm deploy item you could use
the following command:

```
landscaper-cli blueprint add helm-ls export
    [component directory path]
    [execution name]
    [deployitem name]
    --export-param [export parameter name]
    --jsonpath [path in the yaml structure]
    --type [export parameter type]
    --resource-apiVersion: [api version]
    --resource-kind: [kind]
    --resource-name: [name]
    --resource-namespace: [namespace]
    --resource-namespace-param: [name of import parameter]
```

By default, the type of an export parameter is string. You can overwrite this with the flag --type. 

If the flags --resource-* are not specified the export value is fetched from the values.yaml of the helm chart. 
Otherwise it is fetched from the specified rendered resource. Here you could either specify a fix namespace
with --resource-namespace of you specify an input parameter from which the value is fetched. If no flag for
a namespace exists, the target namespace of the deploy item is used.

Example: 

The following example creates an export parameter which gets its value from the specified path in the values.yaml:

```
landscaper-cli blueprint add helm-ls export
    .
    ingress-nginx
    ingress-nginx
    --export-param ingressClass
    --jsonpath .Values.controller.ingressClass
    --type int
```

The following example creates an export parameter which gets its value from a rendered resource. The 
flag --resource-namespace is optional. 

```
landscaper-cli blueprint add helm-ls export
    .
    ingress-nginx
    ingress-nginx
    --export-param replicas
    --jsonpath .spec.replicas
    --type int
    --resource-apiVersion: apps/v1
    --resource-kind: deployment
    --resource-name: controller
    --resource-namespace-param: nginx-namespace
```

The result of applying the example commands to our example [here](./01-step) adds an export parameter
and connects it with the deploy item.

The result could be found [here](./05-step).

## Step 6: Specify images referenced in the helm chart

The images used by a helm chart must be part of the component descriptor of the blueprint. Only then,
all used resourced are declared and could be collected during transport. Furthermore, the helm chart
must explicitly specify all used images in its values.yaml. Images specified somewhere else are hidden
for transport.

Used images could be specified with the following command:

```
landscaper-cli component add helm-ls image
    [component directory path]
    [execution name]
    [deployitem name]
    --jsonpath [path in the values section of the deployitem]
    --image-name [name of the image resource in the component-descriptor]
    --image-reference [oci reference]
    --image-version [version]
```

Example: 

For our example helm chart we add two images explicitly though there are more specified in its values yaml.

Add first image: 

```
landscaper-cli component add helm-ls image
    .
    ingress-nginx
    ingress-nginx
    --image-name controller
    --image-reference k8s.gcr.io/ingress-nginx/controller:v0.43.0
    --image-version v0.43.0
    --jsonpath .controller.admissionWebhooks.patch.image
```

Add second image:

```
landscaper-cli component add helm-ls image
    .
    ingress-nginx
    ingress-nginx
    --image-name certgen
    --image-reference docker.io/jettech/kube-webhook-certgen:v1.5.0
    --image-version v1.5.0
    --jsonpath .controller.admissionWebhooks.patch.image
```

The image references a added to the resourse.yaml and the values section of the deploy item is extended
by referencing these.

The result could be found [here](./06-step)

## Todo

Missing for adding helm chart:

- Describe logically what is achieved and not only referencing some strange directories

Next steps:
- Command to create target resource from kubeconfig
- Command to validate blueprint with input values (there exists already a blueprint validate/render command)
- Command to generate local installation with input values
- Command to upload all stuff to OCI registry




