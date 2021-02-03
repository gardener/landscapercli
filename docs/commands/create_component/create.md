# Create Component

In this tutorial describes how to create a [landscaper](https://github.com/gardener/landscaper) component in the 
local file system. The landscaper component consists of a component descriptor and a blueprint with deploy items. 
Such a component could be easily uploaded into an OCI registry for subsequent usage in landscaper installations.

# Create Component Skeleton

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
landscaper-cli component create [component-name] [component-semver-version] --component-path [some-path]
```

The flag *component-path* is optional with the current folder as default.

Example:

```
landscaper-cli component create github.com/gardener/landscapercli/nginx v0.1.0 --component-path ~/demo-component
```






