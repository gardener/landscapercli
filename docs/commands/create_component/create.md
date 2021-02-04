# Create Component

In this tutorial describes how to create a [landscaper](https://github.com/gardener/landscaper) component in the 
local file system. The landscaper component consists of a component descriptor and a blueprint with deploy items. 
Such a component could be easily uploaded into an OCI registry for subsequent usage in landscaper installations.

# 1 Create Component Skeleton

The component create command creates the skeleton of a landscaper component in the local file system with the 
following structure:

```
component-dir
├── blueprint
│      └── blueprint.yaml
└── component-descriptor.yaml
└── resources.yaml
```

The skeleton consists of the following files and directories:

- blueprint: Directory containing the blueprint.yaml 
- component-descriptor.yaml: Contains a component descriptor skeleton
- resources.yaml: A file for collecting all resources of the component. Initially only the blueprint is added. Later
  these resources are merged into the component-descriptor.yaml as explained later.

The component create command looks as follows: 

```
landscaper-cli component create [component-name] [component-semver-version] --component-directory [some-path]
```

The flag *component-directory* is optional with the current folder as default.

Example:

```
`landscaper-cli component create github.com/gardener/landscapercli/nginx v0.1.0 --component-directory ~/demo-component`
```

The result of this example could be found in the folder
*[01-step-component-created](resouces/01-step-component-created)*. You see that the
resources.yaml file already contains a reference to the blueprint which will be included into the component 
descriptor when the component is complete and uploaded to an OCI registry later. 

The repositoryContext in the component-descriptor is incomplete, because the component has not yet been pushed
into an oci registry, so that the baseUrl is unknown.

## 2 Add Applications as Deploy Items to a blueprint

With our new component we want to deploy applications into kubernetes cluster. Therefore, we need to add such apps to 
the component. This is done by adding the applications as deploy items to it. 


### 2.1 Add an Application provided as a Helm Chart stored in an OCI Registry

In the next step we want to add an nginx application provided as a helm chart as a deploy item to the component. 
We assume the Helm Chart is stored in a OCI registry. 

The general syntax of the command is:

```
landscaper-cli component add helm-ls deployitem [deploy-item-name] \
    --component-directory [some-path]  \
    --oci-reference [oci-helm-chart-reference] \
    --resource-version [resource-version] \
    --cluster-param [target-cluster-param-name] \
    --target-ns-param [target-name-space]
```

The meaning of the arguments and flags is as follows:

- component-directory: Path to the component directory
  
- oci-reference: OCI reference to the helm chart OCI artefact which should be added as a deploy item
  
- resource-version: The version number of the added resource. This might differ from the helm chart version.
  
- cluster-param: Defines the name of the import parameter of the blueprint for the access data to the target cluster 
  into which the added application should be deployed. This allows using the blueprint to deploy the application
  into different target-clusters by providing different input values to this parameter.
  
- target-ns-param: Defines the name of the import parameter of the blueprint for the namespace in the target cluster
  into which the added application should be deployed. This allows using the blueprint to deploy the application
  into different namespaces by providing different input values to this parameter.
  
Example:

```
landscaper-cli component add helm-ls deployitem nginx \
  --component-directory ~/demo-component
  --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0
  --resource-version v0.2.0
  --cluster-param target-cluster
  --target-ns-param nginx-namespace
```

Applying the command on the component skeleton in the folder 
*[01-step-component-created](resouces/01-step-component-created)* results in the resources stored in the folder
*[02-step-add-helm-chart-in-oci](resouces/02-step-add-helm-chart-in-oci)*.

In the file *[blueprint.yaml](resouces/02-step-add-helm-chart-in-oci/demo-component/blueprint/blueprint.yaml)* you
find a new entry under *deployExecutions* referencing to the file 
*[deploy-execution-nginx.yaml](resouces/02-step-add-helm-chart-in-oci/demo-component/blueprint/deploy-execution-nginx.yaml)*
which contains the specification of the new deploy item which adds the ngnix application to the blueprint as a new
deploy item. Furthermore in the imports section of the
*[blueprint.yaml](resouces/02-step-add-helm-chart-in-oci/demo-component/blueprint/blueprint.yaml)* you find the
import parameters *target-cluster* and *nginx-namespace*. These import parameters are referenced at dedicated positions in
*[deploy-execution-nginx.yaml](resouces/02-step-add-helm-chart-in-oci/demo-component/blueprint/deploy-execution-nginx.yaml)*
to later provide their values to the adequate parts of the deploy item specification.

In the file *[resources.yaml](resouces/02-step-add-helm-chart-in-oci/demo-component/resources.yaml)* you
find a new entry for the added nginx helm chart resource.

### 2.2 Add an Application provided as local Helm Chart

In the next step we want to add an echo server application, which provided as a helm chart stored in the folder
*[echo-server](resouces/charts/echo-server)*.

The general syntax of the command is:

```
landscaper-cli component add helm-ls deployitem [deploy-item-name] \
    --component-directory [some-path]  \
    --chart-directory [chart-directory-path] \
    --cluster-param [target-cluster-param-name] \
    --target-ns-param [target-name-space]
```

The meaning of the arguments and flags is the same as in the chapter before. The new flag *chart-directory* 
is for specifying the path to the directory where the chart is stored.

Example:

```
landscaper-cli component add helm-ls deployitem echo \
  --component-directory ~/demo-component
  --chart-directory ../resouces/charts/echo-server
  --resource-version v0.3.0
  --cluster-param target-cluster
  --target-ns-param echo-server-namespace
```

We use the same parameter name for *cluster-param*. Therefore, both the nginx and the echo-server applications
will be deployed to the same cluster with the resulting blueprint/component. We use a different *target-ns-parameter*,
which allows the later deployment of the applications into different namespaces.

Applying the command on the component folder in
*[02-step-add-helm-chart-in-oci](resouces/02-step-add-helm-chart-in-oci)* results in the resources stored 
in the folder
*[03-step-add-local-helm-chart](resouces/03-step-add-local-helm-chart)*.

In the file *[blueprint.yaml](resouces/03-step-add-local-helm-chart/demo-component/blueprint/blueprint.yaml)* you
find a new entry under *deployExecutions* referencing to the file
*[deploy-execution-nginx.yaml](resouces/03-step-add-local-helm-chart/demo-component/blueprint/deploy-execution-echo.yaml)*
which contains the specification of the new deploy item which adds the echo server application to the blueprint as a new
deploy item. 

Furthermore in the imports section of the 
*[blueprint.yaml](resouces/03-step-add-local-helm-chart/demo-component/blueprint/blueprint.yaml)* you find the
new import parameter *echo-server-namespace*. This import parameter as well as the parameter *target-cluster* 
are referenced at dedicated positions in
*[deploy-execution-nginx.yaml](resouces/03-step-add-local-helm-chart/demo-component/blueprint/deploy-execution-echo.yaml)*
to later provide their values to the adequate parts of the deploy item specification.

In the file *[resources.yaml](resouces/03-step-add-local-helm-chart/demo-component/resources.yaml)* you
find a new entry for the added echo server helm chart resource. You see here the input type *dir* which means 
that during the later upload of the component to the OCI registry the complete chart folder is added as a
separate layer of the OCI artifact. 

### 2.3 Add a Deploy Item providing particular Kubernetes Manifests

In the next step we want to add deploy item, which deploys particular kubernetes resources.

The general syntax of the command is:

```
landscaper-cli component add manifest deployitem [deploy-item-name] \
    --component-directory [some-path]  \
    --manifest-file [path-to-yaml-file] \
    --import-param [param-name:param-type] \
    --target-ns-param [target-name-space]
```

The meaning of the arguments and flags is as follows:

- component-directory: Path to the component directory
  
- manifest-file: Path to a yaml file containing the manifest of one kubernetes resource. This flag could appear 
  multiple times.
  
- import-param: The value this flag consists of two parts separated by a colon, e.g. *replicas:integer*. 
  The first part defines the name on an import parameter for the blueprint and the second part its type. 
  Furthermore, the import parameters are connected to matching field in the manifests as shown in the example below.
  Currently, only integer, string and boolean types are supported. This flag could appear multiple times.

- target-ns-param: Defines the name of the import parameter of the blueprint for the namespace in the target cluster
  into which the added application should be deployed. This allows using the blueprint to deploy the application
  into different namespaces by providing different input values to this parameter.

Example:

Assume we want to add a deploy item which provisions the two secrets stored in the folder 
*[set1](resouces/manifests/set1)*. It should be possible that the passwords in the secrets could be set when the
component is deployed. The corresponding command looks as the following:

```
landscaper-cli component add manifest deployitem secrets \
  --component-directory ~/demo-component
  --manifest-file ../resouces/manifests/set1/demo-secret-1.yaml
  --manifest-file ../resouces/manifests/set1/demo-secret-2.yaml
  --import-param password-1:string
  --import-param password-2:string
  --cluster-param target-cluster
```

Again, we use the same parameter name for *cluster-param*. We define two import parameters which match the corresponding
fields in the secrets.

Applying the command on the component folder in
*[03-step-add-local-helm-chart](resouces/03-step-add-local-helm-chart)* results in the resources stored
in the folder *[04-step-add-secret-manifests](resouces/04-step-add-secret-manifests)*.

In the file *[blueprint.yaml](resouces/04-step-add-secret-manifests/demo-component/blueprint/blueprint.yaml)* you
find a new entry under *deployExecutions* referencing to the file
*[deploy-execution-secrets.yaml](resouces/04-step-add-secret-manifests/demo-component/blueprint/deploy-execution-secrets.yaml)*
which contains the specification of the new deploy item added to the blueprint which deploys two secrets.

Furthermore in the imports section of the
*[blueprint.yaml](resouces/04-step-add-secret-manifests/demo-component/blueprint/blueprint.yaml)* you find the
new import parameters *password-1* and *password-2*. These import parameters are referenced at dedicated positions in
*[deploy-execution-secrets.yaml](resouces/04-step-add-secret-manifests/demo-component/blueprint/deploy-execution-secrets.yaml)*
to later provide their values to the adequate parts of the manifest specification. The so called *dedicated positions*
are all those field values with in the manifests with the same string as the import parameter name.

In the file *[resources.yaml](resouces/03-step-add-local-helm-chart/demo-component/resources.yaml)* you
find a new entry for the added echo server helm chart resource. You see here the input type *dir* which means
that during the later upload of the component to the OCI registry the complete chart folder is added as a
separate layer of the OCI artifact.

## Todo

- Add Resource
- Upload to OCI Factor with URL adaption
- Create Installation

- Describe that the current helm deploy mechanism is not helm but only helm template plus apply




