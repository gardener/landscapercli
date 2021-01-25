# Commands

## Create component-descriptor and blueprint

The following command creates a component descriptor and a blueprint. It also adds the blueprint to the componentReferences of the component descriptor.

Command:

```
landscaper-cli create
    [component directory path]
    [component name]
    [component version]
```

Example:

```
landscaper-cli blueprint create
    .
    github.com/gardener/landscapercli/nginx
    v0.1.0
```

As a result we get the following file structure inside the component directory:

```
├── blueprint
│   └── blueprint.yaml
└── component-descriptor.yaml
```

The repositoryContext in the component-descriptor is incomplete, because the component-descriptor has not yet been pushed into an oci registry, so that the baseUrl is unknown.

The component descriptor contains a reference to the blueprint. The reference is not yet complete, because the blueprint has not yet been pushed into an oci registry, so that the address is imknown. 

## Add deployitem of type "helm-templating"

The following command adds a deployitem of type "helm-templating" to a blueprint.

Command:

```
landscaper-cli blueprint add helm-templating deployitem
    [component directory path]
    [execution name]
    [deployitem name]
    [oci reference]
    [chart-version]
```

Example:

```
landscaper-cli blueprint add helm-templating deployitem
    .
    ingress-nginx
    ingress-nginx
    oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0
    v0.1.0
```

Result:

- The chart is added to the componentReferences of the component-descriptor.
- An execution *ingress-nginx* with a deployitem *ingress-nginx* is added to the blueprint. 
- Import parameters *cluster* and *targetnamespace* are added to the blueprint and used in the deployitem.

## Add import

The following command adds an import to a blueprint.

Command:

```
landscaper-cli blueprint add import
    [component directory path]
    [import parameter name]
    --type [string | int | ...]
```

The `--type` flag is optional. The default type is string.

Example:

```
landscaper-cli blueprint add import
    .
    failure-threshold
    --type int
```

As a result, an import parameter definition is added to the blueprint.

## Add export

The following command adds an export to a blueprint.

Command:

```
landscaper-cli blueprint add export
    [component directory path]
    [export parameter name]
    --type [string | int | ...]
```

## Add value to deployitem of type "helm-templating"

The following command adds a single key-value pair to the values section of a deployitem.

Command:

```
landscaper-cli blueprint add helm-templating value 
    .
    [execution name]
    [deployitem name]
    --key [path in the values section of the deployitem]
    --value [value]
```

The values section can have deep structure. Therefore, a key can be a path.

Example:

```
landscaper-cli blueprint add helm-templating value
    .
    ingress-nginx
    ingress-nginx
    --key controller.readinessProbe.failureThreshold
    --value 5
```

Values can also be taken from a file. This variant has the advantage that a deep structure can be added with one command. 

````
landscaper-cli blueprint add helm-templating value
    .
    [execution name]
    [deployitem name]
    --file [value]
````

A value can also come from an import parameter. If the import parameter is not yet defined, the import will be added to the blueprint.

```
landscaper-cli blueprint add helm-templating value
    .
    [execution name]
    [deployitem name]
    --key [path in the values section of the deployitem]
    --import [import parameter name]
    --type [import parameter type]
```

## Add image to deployitem of type "helm-templating"

Command:

```
landscaper-cli blueprint add helm-templating image
    .
    [execution name]
    [deployitem name]
    --key [path in the values section of the deployitem]
    --image-name [name of the image resource in the component-descriptor]
    --image-reference [oci reference]
    --image-version [version]
```

Example:

```
landscaper-cli blueprint add helm-templating image
    .
    ingress-nginx
    ingress-nginx
    --key controller.admissionWebhooks.patch.image
    --image-name certgen
    --image-reference docker.io/jettech/kube-webhook-certgen:v1.5.0
    --image-version v1.5.0
```

Result:

- The image is added to the componentReferences of the component-descriptor:

  ```
  componentReferences:
  - type: ociImage
    name: certgen
    version: v1.5.0
    relation: external
    access:
      type: ociRegistry
      imageReference: docker.io/jettech/kube-webhook-certgen:v1.5.0
  ```

- The image is added to the values section of the deployitem:

  ```
  controller:
    admissionWebhooks:
      patch:
        image: {{ with (getResource .cd "name" "certgen") }}{{ .access.imageReference }}{{ end }}
  ```

  