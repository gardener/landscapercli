# Render Blueprints

During the execution of a blueprint with an installation, Landscaper creates deploy items and subinstallations 
based on the imported values.  The creation of these resources can be locally tested by using the Landscaper CLI and 
its [render command](../../reference/landscaper-cli_blueprints_render.md).

The simplest form of the command is:

```shell
landscaper-cli blueprints render [path to blueprint directory]

# short form:
landscaper-cli bp render [path to blueprint directory]
```

Note that you must specify the path to the blueprint **directory**, rather than the path to the `blueprint.yaml` 
file in this directory.

### Import Values

Usually, a blueprint has import parameters that are used in the templating of deploy items and subinstallations.
Values for the import parameters can be provided with the flag `-f` together with a path to a values yaml file. 
The flag can be set multiple times to provide more than one values file.

The render command does not only template the deploy items and subinstallations. It also validates the values against 
the types of the corresponding import parameters in the blueprint.

```shell
landscaper-cli blueprints render [path to blueprint directory] -f [path to values yaml file]
```

The values yaml files must have the following structure:

```yaml
imports:
  [parameter name 1]: [parameter value 1]
  [parameter name 2]: [parameter value 2]
  ...
```

The parameter names in the values yaml file must match with the import parameter names of the blueprint.


### Examples 

You can find a first simple example [here](../../examples/render-blueprint/01-render-with-values).
It contains the following files:

- [blueprint/blueprint.yaml](../../examples/render-blueprint/01-render-with-values/blueprint/blueprint.yaml)
- [values.yaml](../../examples/render-blueprint/01-render-with-values/values.yaml)
- [render.sh](../../examples/render-blueprint/01-render-with-values/render.sh) &mdash; you can run this script to try out the render command

### More Examples

There are further examples in the [Landscaper Examples](https://github.com/gardener/landscaper-examples/tree/master/render-blueprint) repository.

### Output

By default, the command prints the rendered resources to stdout. The output has the following structure:

```shell script
--------------------------------------
-- deployitems state
--------------------------------------
state:
  <execution name>: ...

--------------------------------------
-- deployitems <deployitem name>
--------------------------------------
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: DeployItem
...

--------------------------------------
-- subinstallations state
--------------------------------------
state:
  ...
  
--------------------------------------
-- subinstllations <subinstallation name>
--------------------------------------
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: Installation
...
```

Alternatively, the rendered resources can be written to a directory by specifying `-w /path/to/output`.
The resources are written as files in the following directory structure to the given path.
```
/path/to/output
├── deployitems
│   ├── mydeployitem
│   └── state
└── subinstallations
    ├── mysubinstallation
    └── state
```

### Using Data from the Component Descriptor

The rendering of a blueprint might use data from the component descriptor. 
With the flag `-c` or `--component-descriptor` a path to a local component descriptor file can be specified in the 
render command. This allows to test the rendering before the component is uploaded into a component registry.

The component descriptor might have references to further component descriptors. If they are needed for the rendering, 
their paths can be specified with the flag `-a` or `--additional-component-descriptor`. This flag can be set multiple 
times.

##### Examples

- [How to access the component descriptor](https://github.com/gardener/landscaper-examples/tree/master/render-blueprint/access-repository-context)
- [How to access a referenced component descriptor](https://github.com/gardener/landscaper-examples/tree/master/render-blueprint/access-cd-reference)

### Using Local Resources

The rendering of a blueprint might use resources of the component. For example a JSON schema might be used for the 
validation of import values. Again one wants to provide such resources as local files, so that the rendering can be 
tested before the component is uploaded into a component registry. 

With the flag `-r` or `--resources` one can specify the path to a so-called resources file. 
An example of a resources file is shown below. It is a multi yaml file with a list of resources which will be added
to the component descriptor. Beside other information, it contains the paths where local resources are located in
the file system (field `input.path`). The render command can then access the local resources from there.

```yaml
---
type: blueprint
name: blueprint
version: v0.1.0
relation: local
input:
  type: dir
  path: ./blueprint
  mediaType: application/vnd.gardener.landscaper.blueprint.v1+tar+gzip
  compress: true
...
---
type: landscaper.gardener.cloud/jsonschema
name: vpa
version: v0.1.0
relation: local
input:
  type: file
  path: ./resources/schemas/vpa.json
  mediaType: application/vnd.gardener.landscaper.jsonschema.layer.v1.json
  compress: false
...
``` 

Remark: the resources file is primarily used by the 
[command that creates a transport (CTF) file](../../reference/landscaper-cli_component-cli_component-archive.md),
see also: [add a resource](https://github.com/gardener/component-cli/tree/main/docs#add-a-resource).

##### Examples

- [How to access a schema resource during the validation of import values](https://github.com/gardener/landscaper-examples/tree/master/render-blueprint/schema-in-cd)
- [How to access a chart resource during the templating of a deploy item](https://github.com/gardener/landscaper-examples/tree/master/render-blueprint/resource-in-cd)
