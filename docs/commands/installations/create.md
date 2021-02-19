# Creating installations
The general two steps for installing software components from an OCI registry via Landscaper are the following:

1. Creating one or more targets: These contain the access information for the system(s), into which the software components should be deployed. More information on this topic can be found [here](../targets/create.md).

1. Creating one or more installation objects: These contain the reference to the software component, the configuration of the component via its import and export parameters, and a reference to one or more target objects.

> More background information can be found in the [Landscaper docs](https://github.com/gardener/landscaper/blob/master/docs/README.md).

## Creating installations
> The example we will use in the following is based on the tutorial for creating components, which you can find [here](../create_component/create.md). In this tutorial you can also find more information on how to interact with OCI registries. For the following, we assume you are working with the cluster-internal OCI registry from the `landscaper-cli quickstart install` command and have `kubectl port-forward` to the OCI registry pod running in the background.

To recap, the blueprint of the example component is shown in the following snippet. The blueprint's interna are intentionally omitted, as they are not of interest in this step.

```yaml
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: Blueprint

imports:
- default:
    value: null
  name: target-cluster
  required: true
  targetType: landscaper.gardener.cloud/kubernetes-cluster
- default:
    value: null
  name: nginx-namespace
  required: true
  schema:
    type: string
- default:
    value: null
  name: echo-server-namespace
  required: true
  schema:
    type: string
- default:
    value: null
  name: password-1
  required: true
  schema:
    type: string
- default:
    value: null
  name: password-2
  required: true
  schema:
    type: string

exports: []

...
```

For installing a blueprint via Landscaper, an installation object must be created. The installation object references the blueprint and provides all the configuration information that the blueprint expects. The configuration information consists of all imports and exports that the blueprint defines. The actual values for these imports and exports can either be hardcoded in the installation, or loaded dynamically from data objects at runtime.

The command `landscaper-cli installations create [baseURL] [component name] [component version]` creates an installation template for a specified blueprint which is stored in an OCI registry.

> A component descriptor can contain multiple blueprints. If this is the case, the blueprint for which the installation should be created must be specified via the parameter `--blueprint-resource-name`.

E.g., executing the command

```
landscaper-cli installations create localhost:5000 github.com/gardener/landscapercli/nginx v0.1.0 --name my-installation
```

creates the following installation template:

```yaml
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: Installation
metadata:
  name: my-installation
spec:
  blueprint:
    ref:
      resourceName: blueprint
  componentDescriptor:
    ref:
      componentName: github.com/gardener/landscapercli/nginx
      repositoryContext:
        baseUrl: oci-registry.landscaper.svc.cluster.local:5000
        type: ociRegistry
      version: v0.1.0
  exports: {}
  imports:
    data:
      - # JSON schema
        # {
        #   "type": "string"
        # }
        name: nginx-namespace
      - # JSON schema
        # {
        #   "type": "string"
        # }
        name: echo-server-namespace
      - # JSON schema
        # {
        #   "type": "string"
        # }
        name: password-1
      - # JSON schema
        # {
        #   "type": "string"
        # }
        name: password-2
    targets:
      - # Target type: landscaper.gardener.cloud/kubernetes-cluster
        name: target-cluster
        target: ""
```

In the template, landscaper-cli has already filled out the correct references to the component descriptor and blueprint, see the properties `spec.componentDescriptor` and `spec.blueprint`. Additionally, a skeleton for all imports and exports of the blueprint was created.

> The JSON schema definitions of all imports and exports of the blueprint are by default rendered as comments into the installation. This behaviour can be disabled by setting  `--render-schema-info=false`.

The first step to make this installation work is to set the target references. For each target that a blueprint expects via its import parameters, an entry in the installation template is rendered in the list `spec.imports.targets`. Each entry has a comment that shows the target type that the blueprint expects. In this case, the blueprint expects 1 target of type `landscaper.gardener.cloud/kubernetes-cluster`. In the following, we set this target to the value `#my-target`.

> Target references in installations must start with a "#"!

```yaml
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: Installation
metadata:
  name: my-installation
spec:
  ...
  imports:
    data:
      ...
    targets:
      - # Target type: landscaper.gardener.cloud/kubernetes-cluster
        name: target-cluster
        target: "#my-target"
```

At runtime, Landscaper will look for a target object with the name `my-target` and the target type `landscaper.gardener.cloud/kubernetes-cluster` in the same K8s namespace where the installation is applied, and use it for the deploying into the target system.

Now, the data imports in the property `spec.imports.data` must still be defined. In general, data imports can be fulfilled by either

1. referencing data objects from which the values will be loaded at runtime. These data objects must be created by hand or are created by Landscaper as export values from other installations. If the referenced data objects don't exist at runtime, landscaper cannot execute the installation.
1. setting them to hardcoded values in the installation. In this case, no data objects are needed on the cluster at runtime.

Currently landscaper-cli only provides support for the second option, which is described in the following section.

### Setting import parameters to hardcoded values
For setting the imports to hardcoded values you can use the command `landscaper-cli installations set-import-parameters`. 

E.g., executing the command

```
landscaper-cli installations set-import-parameters ./my-installation.yaml nginx-namespace=nginx echo-server-namespace=echo-server password-1=pw1 password-2=pw2
```

will modify `my-installation.yaml` to look like the following:

```yaml
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: Installation
metadata:
  name: my-installation
spec:
  blueprint:
    ref:
      resourceName: blueprint
  componentDescriptor:
    ref:
      componentName: github.com/gardener/landscapercli/nginx
      repositoryContext:
        baseUrl: oci-registry.landscaper.svc.cluster.local:5000
        type: ociRegistry
      version: v0.1.0
  exports: {}
  importDataMappings:
    echo-server-namespace: echo-server
    nginx-namespace: nginx
    password-1: pw1
    password-2: pw2
  imports:
    targets:
    - name: target-cluster
      target: '#my-target'
```

The command has removed the imports from `spec.imports` and added them to `spec.importDataMappings` with the hardcoded values.

Before applying the installation, you should make sure that the namespaces you have used for the parameters `echo-server-namespace` and `nginx-namespace` actually exist on the target cluster. Then the installation can be applied using `kubectl apply -f ./my-installation.yaml`.

