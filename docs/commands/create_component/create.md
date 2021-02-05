# Create Component

This tutorial describes how to develop a [landscaper](https://github.com/gardener/landscaper) component in your 
local file system. A landscaper component consists of a component descriptor and a blueprint with deploy items. 
It defines a set of cloud artifacts (e.g. kubernetes applications, cloud infrastructure components etc.) to be
installed. Such a component could be easily uploaded into an OCI registry and subsequently used with different
configuration settings to install the specified actifacts into a landscaper controlled cloud environment. 

# 1 Create Component Skeleton

The *component create command* set up the skeleton of a landscaper component with a blueprint in the local file system 
with the following structure:

```
component-dir
├── blueprint
│      └── blueprint.yaml
└── component-descriptor.yaml
└── resources.yaml
```

The skeleton consists of the following files and directories:

- blueprint: Directory containing the blueprint.yaml. In the next steps the cloud artifacts are added to the blueprint.
- component-descriptor.yaml: Contains a component descriptor skeleton.
- resources.yaml: A file for collecting all resources of the component. Initially only the blueprint is added. 
  Finally, these resources are merged into the component-descriptor.yaml as explained later.

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

With our new component we want to deploy applications and other cloud artifacts into our cloud infrastructure. 
Therefore, we need to add such applications/artifacts to the component. This is done by adding the 
applications/artifacts as deploy items to it. 


### 2.1 Add an Application provided as a Helm Chart stored in an OCI Registry

In the next step we want to add an nginx application provided as a helm chart as a deploy item to the component. 
We assume the Helm Chart is stored in an OCI registry. 

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
  into which the added application should be deployed. Later, this allows using the blueprint to deploy the application
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

This chapter describes how to add a helm application, which is stored on your file system. You could easily 
download a helm chart from every Helm Chart Repository or OCI registry using 
*[helm pull command](https://helm.sh/docs/helm/helm_pull/)* or the new 
*[helm OCI support](https://helm.sh/docs/topics/registries/#enabling-oci-support)*.

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

In this example, we want to add an echo server application, which is provided as a helm chart stored in the folder
*[echo-server](resouces/charts/echo-server)*. 

```
landscaper-cli component add helm-ls deployitem echo \
  --component-directory ~/demo-component
  --chart-directory ../resouces/charts/echo-server
  --resource-version v0.3.0
  --cluster-param target-cluster
  --target-ns-param echo-server-namespace
```

In the example, we use the same parameter name for *cluster-param*. Therefore, both the nginx and the echo-server applications
will be deployed to the same cluster with the resulting blueprint/component. We use a different *target-ns-parameter*,
which allow the later deployment of the applications into different namespaces.

Applying the command on the component folder in
*[02-step-add-helm-chart-in-oci](resouces/02-step-add-helm-chart-in-oci)* results in the resources stored 
in the folder
*[03-step-add-local-helm-chart](resouces/03-step-add-local-helm-chart)*.

In the file *[blueprint.yaml](resouces/03-step-add-local-helm-chart/demo-component/blueprint/blueprint.yaml)* you
find a new entry under *deployExecutions*, referencing to the file
*[deploy-execution-echo.yaml](resouces/03-step-add-local-helm-chart/demo-component/blueprint/deploy-execution-echo.yaml)*
which contains the specification of the new deploy item adding the echo server application to the blueprint. 

In the imports section of the 
*[blueprint.yaml](resouces/03-step-add-local-helm-chart/demo-component/blueprint/blueprint.yaml)* you find the
new import parameter *echo-server-namespace*. This import parameter as well as the parameter *target-cluster* 
are referenced at dedicated positions in
*[deploy-execution-echo.yaml](resouces/03-step-add-local-helm-chart/demo-component/blueprint/deploy-execution-echo.yaml)*
to later provide their values to the adequate parts of the deploy item specification.

In the file *[resources.yaml](resouces/03-step-add-local-helm-chart/demo-component/resources.yaml)* you
find a new entry for the added echo server helm chart resource. You see here the input type *dir* which means 
that during the later upload of the component to the OCI registry the complete chart folder is added as a
separate layer of the OCI artifact. 

### 2.3 Add a Deploy Item providing particular Kubernetes Manifests

In the next step we want to a add deploy item, which deploys particular kubernetes resources.

The general syntax of the command is:

```
landscaper-cli component add manifest deployitem [deploy-item-name] \
    --component-directory [some-path]  \
    --manifest-file [path-to-yaml-file] \
    --import-param [param-name:param-type] \
    --cluster-param [target-cluster-param-name] 
```

The meaning of the arguments and flags is as follows:

- component-directory: Path to the component directory
  
- manifest-file: Path to a yaml file containing the manifest of one kubernetes resource. This flag could appear 
  multiple times.
  
- import-param: The value of this flag consists of two parts separated by a colon, e.g. *replicas:integer*. 
  The first part defines the name on an import parameter for the blueprint and the second part its type. 
  Furthermore, the import parameters are connected to matching field values in the manifests as shown in the example below.
  Currently, only integer, string and boolean types are supported. This flag could appear multiple times.

- cluster-param: Defines the name of the import parameter of the blueprint for the access data to the target cluster
  into which the manifests should be deployed. Later, this allows using the blueprint to deploy the manifests
  into different target-clusters by providing different input values to this parameter.

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
field values in the *[secret yaml files](resouces/manifests/set1)*.

Applying the command on the component folder in
*[03-step-add-local-helm-chart](resouces/03-step-add-local-helm-chart)* results in the resources stored
in the folder *[04-step-add-secret-manifests](resouces/04-step-add-secret-manifests)*.

In the file *[blueprint.yaml](resouces/04-step-add-secret-manifests/demo-component/blueprint/blueprint.yaml)* you
find a new entry under *deployExecutions* referencing to the file
*[deploy-execution-secrets.yaml](resouces/04-step-add-secret-manifests/demo-component/blueprint/deploy-execution-secrets.yaml)*
which contains the specification of the new deploy item added to the blueprint which deploys two secrets.

In the imports section of the
*[blueprint.yaml](resouces/04-step-add-secret-manifests/demo-component/blueprint/blueprint.yaml)* you find the
new import parameters *password-1* and *password-2*. These import parameters are referenced at dedicated positions in
*[deploy-execution-secrets.yaml](resouces/04-step-add-secret-manifests/demo-component/blueprint/deploy-execution-secrets.yaml)*
to later provide their values to the adequate parts of the manifest specification. The so called *dedicated positions*
are all those field values in the manifests with the same string as the import parameter name.

Nothing was added to the file *[resources.yaml](resouces/04-step-add-secret-manifests/demo-component/resources.yaml)* 
because all manifests are already part of the 
*[blueprint.yaml](resouces/04-step-add-secret-manifests/demo-component/blueprint/blueprint.yaml)*.

## 3 Upload Component into an OCI Registry

Now we describe how to upload the locally developed *[component](resouces/04-step-add-secret-manifests)* into an OCI registry.

### 3.1 Add Resources to Component Descriptor

In a first step, we add the resources in file *[resources.yaml](resouces/04-step-add-secret-manifests/demo-component/resources.yaml)* 
to the component descriptor with the following command:

```shell script
landscaper-cli components-cli component-archive resources add \
   .../landscapercli/docs/commands/create_component/resouces/05-step-prepare-push/demo-component \
   -r .../landscapercli/docs/commands/create_component/resouces/05-step-prepare-push/demo-component/resources.yaml
```

Applying the command on the component folder in
*[04-step-add-secret-manifests](resouces/04-step-add-secret-manifests)* results in the resources stored
in the folder *[05-step-prepare-push](resouces/05-step-prepare-push)*.

The command packs all resources with *input.type=dir* into a *blobs* directory. 
Moreover, it adds all resources to the
*[component-descriptor.yaml](resouces/05-step-prepare-push/demo-component/component-descriptor.yaml)*.

**Remark:**
The resources in the *blobs* directory will be stored together with the component descriptor in one OCI artifact.

### 3.2 Maintain Base URL of the OCI Registry

Set the field *component.repositoryContexts.baseUrl* of the 
*[component-descriptor.yaml](resouces/05-step-prepare-push/demo-component/component-descriptor.yaml)*
to the base URL of the OCI registry into which you want to upload the component, e.g.

```yaml
component:
  repositoryContexts:
  - baseUrl: eu.gcr.io/some-path
    type: ociRegistry
```

The component will be uploaded to the OCI registry to the namespace/repository

```
[baseUrl]/component-descriptors/[component-name]
```

### 3.3 Upload Component

Next, we upload the component to the OCI registry with the following command:

```shell script
landscaper-cli components-cli ca remote push \
    eu.gcr.io/some-path \
    github.com/gardener/landscapercli/nginx \
    v0.1.0 \
    .../landscapercli/docs/commands/create_component/resouces/05-step-prepare-push/demo-component
```

In this case, the push command has the following arguments:

* *eu.gcr.io/some-path*: the base URL of the OCI registry as defined in the component-descriptor.yaml
* *github.com/gardener/landscapercli/nginx*: the component name as defined in the component-descriptor.yaml
* *v0.1.0*: the component version as defined in the component-descriptor.yaml
* *.../landscapercli/docs/commands/create_component/resouces/05-step-prepare-push/demo-component*: the path to the component directory

After the push, the OCI registry contains the following artefact:

```
eu.gcr.io/some-path/component-descriptors/github.com/gardener/landscapercli/nginx:v0.1.0
```

It contains the component descriptor, the blueprint, and the helm chart of the echo server.  It does not contain 
the nginx helm chart, because this is only referenced and stored as a separate OCI artifact.


## Todo

- Create Installation

- Describe that the current helm deploy mechanism is not helm but only helm template plus apply




