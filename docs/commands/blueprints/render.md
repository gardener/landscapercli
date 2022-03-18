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

## Render Modes

The blueprint render command supports two modes of rendering.
The default mode renders all deploy items and subinstallations that are defined within the blueprint, defined in the render arguments.

When specifying export templates via the `-e, --export-templates` flag, the blueprint renderer also generates the export values for deploy items and/or subinstallations.
With this generated export values subinstallations (both locally defined in the blueprint and defined in external components) will be executed recursively.
The blueprint renderer will output all deploy items, imports and exports for the subinstallations it executes.


### Rendering without Export Templates (default mode)

In the default mode, no additional flags have to be specified.

#### Output

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

### Rendering with Export Templates

For rendering with export templates, the `-e, --export-templates` flag has to be specified.
The flag has to point to a yaml file, that containing the export templates.

The export templates yaml file must have the following structure:

```yaml
# installations is an array of installation templates
installations:
  - name: "name of the installation template"
    selector: "a golang regular expression string that matches the path of an installation"
    template: |
      {{/* a go template that must result in a valid yaml structure containing a "dataExports" map and a "targetExports" map */}}

      dataExports:
        key: value

      targetExports:
        key: value

# deployItems is an array of deploy item templates
deployItems:
  - name: "name of the deploy item template"
    selector: "a golang regular expression string that matches the path of a deploy item"
    template: |
      {{/* a go template that must result in a valid yaml structure containing a "exports" map */}}

      exports:
        key: value
```

Example:

```yaml
installations:
  - name: subinst-c
    selector: ".*/subinst-c"
    template: |
      dataExports:
        subinst-c-export: {{ .installation.metadata.name }}
      targetExports: {}

deployItems:
  - name: subinst-a-deploy
    selector: ".*/subinst-a-deploy"
    template: |
      exports:
        subinst-a-export-a: {{ .deployItem.metadata.name }}
        subinst-a-export-b: {{ .cd.component.name }}

  - name: subinst-b-deploy
    selector: ".*/subinst-b-deploy"
    template: |
      exports:
        subinst-b-export-a: {{ .deployItem.metadata.name }}
        subinst-b-export-b: {{ .cd.component.name }}

```

The export templates file contains two keys, _installations_ and _deployItems_.

#### Installations 

The renderer tries to match the selector to the path of the installation. Each installation starts with the "root" installation.
If the user specified blueprint contains subinstallations, the path of the subinstallation will be appended to root.
For example the installations "root/subinst-a", "root/subinst-b", "root/subinst-a/subinst-c" exist.
A selector value of ".*subinst-a" would match installation "root/subinst-a", a selector value of ".*subinst-.*" would match installations "root/subinst-a", "root/subinst-b" and "root/subinst-a/subinst-c".

The template field must contain a go template that results in a valid yaml structure, containing a _dataExports_ and a _targetExports_ map.
_dataExports_ contains data objects and _targetExports_ contains Landscaper Targets.
All exports defined by the template will be available for all sibling installations and its parent installation.
For example, if the matched template for installation "root/subinst-a" is exporting key "export-a", this export definition will be available for the "root" installation as well as all subinstallations of "root".

During template execution the following parameters will be available:

* _imports_: the installation imports
* _installationPath_: the complete path to the installation, which is also used for selecting the template
* _templateName_: the user specified name of this template
* _installation_: the complete installation structure
* _cd_: the component descriptor
* _components_: the component descriptor list
* _state_: contains the calculated installation state

#### DeployItems

The renderer tries to match the selector to the path of the deploy item. The path contains the name of the installation, which blueprint contains the deploy item as well as all parent installations.
For example, "root/subinst-a/deploy-a", "root/subinst-b/deploy-a".
A selector value of ".*deploy-a" would match deploy items "root/subinst-a/deploy-a" and "root/subinst-b/deploy-a". A selector value of ".*/subinst-a/deploy-a" would match  deploy item "root/subinst-a/deploy-a".

The template field must contain a go template that results in a valid yaml structure, containing an exports map.
All exports defined by the template will be available for the blueprint of the respective installation.

During template execution, the following parameters will be available:

* _imports_: the installation imports
* _installationPath_: the complete installation path that contains the deploy item, which is also used for selecting the template
* _templateName_: the user specified name of this template
* _deployItem_: the complete deploy item structure
* _cd_: the component descriptor
* _components_: the component descriptor list

#### Output

By default, the command prints the rendered resources to stdout. The output has the following structure:

```shell script
executing installation root
--------------------------------------
-- root installation
--------------------------------------
metadata:
  name: root
spec:
  ...

--------------------------------------
-- root imports
--------------------------------------
import-a: vala
import-b: valb
...

executing installation root/subinst-a
--------------------------------------
-- root/subinst-a installation
--------------------------------------
metadata:
  name: subinst-a
spec:
  ...

--------------------------------------
-- root/subinst-a imports
--------------------------------------
import-a: vala
...

executing deploy item root/subinst-a/subinst-a-deploy
--------------------------------------
-- root/subinst-a deployitems subinst-a-deploy
--------------------------------------
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: DeployItem
metadata:
  name: subinst-a-deploy
spec:
  config:
    ...

--------------------------------------
-- root/subinst-a exports
--------------------------------------
export-subinst-a: vala
...
```

However, since the output can be quite large, it is recommended to specify an output directory with flag `-w /path/to/output`.
All resources are then written to the specified output directory, stored hierarchically in relation to their parent installation, starting with the "root" installation.

```
/path/to/output
└── root
    ├── exports.yaml
    ├── imports.yaml
    ├── installation.yaml
    ├── subinst-a
    │   ├── deployitems
    │   │   ├── deploy-a.yaml
    │   │   └── state.yaml
    │   ├── exports.yaml
    │   ├── imports.yaml
    │   └── installation.yaml
    └── subinst-b
        ├── deployitems
        │   └── deplpoy-a.yaml
        ├── exports.yaml
        ├── imports.yaml
        └── installation.yaml
    
```

#### Example

An example using export templates can be found [here](../../examples/render-blueprint/02-render-with-export-templates).

### More Examples

There are further examples in the [Landscaper Examples](https://github.com/gardener/landscaper-examples/tree/master/render-blueprint) repository.

