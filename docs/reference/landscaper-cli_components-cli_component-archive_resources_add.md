## landscaper-cli components-cli component-archive resources add

Adds a resource to an component archive

### Synopsis


add generates resources from a resource template and adds it to the given component descriptor in the component archive.
If the resource is already defined (quality by identity) in the component-descriptor it will be overwritten.

The component archive can be specified by the first argument, the flag "--archive" or as env var "COMPONENT_ARCHIVE_PATH".
The component archive is expected to be a filesystem archive. If the archive is given as tar please use the export command.

The resource template can be defined by specifying a file with the template with "resource" or it can be given through stdin.

The resource template is a multidoc yaml file so multiple templates can be defined.

<pre>

---
name: 'myimage'
type: 'ociImage'
relation: 'external'
version: 0.2.0
access:
  type: ociRegistry
  imageReference: eu.gcr.io/gardener-project/component-cli:0.2.0
...
---
name: 'myconfig'
type: 'json'
relation: 'local'
input:
  type: "file"
  path: "some/path"
...
---
name: 'myconfig'
type: 'json'
relation: 'local'
input:
  type: "dir"
  path: /my/path
  compress: true # defaults to false
  exclude: "*.txt"
...

</pre>


```
landscaper-cli components-cli component-archive resources add [component archive path] [-r resource-path] [flags]
```

### Options

```
  -a, --archive string             path to the component archive directory
      --component-name string      name of the component
      --component-version string   version of the component
  -h, --help                       help for add
      --repo-ctx string            repository context url for component to upload. The repository url will be automatically added to the repository contexts.
  -r, --resource string            The path to the resources defined as yaml or json
```

### Options inherited from parent commands

```
      --cli                  logger runs as cli logger. enables cli logging
      --dev                  enable development logging which result in console encoding, enabled stacktrace and enabled caller
      --disable-caller       disable the caller of logs (default true)
      --disable-stacktrace   disable the stacktrace of error logs (default true)
      --disable-timestamp    disable timestamp output (default true)
  -v, --verbosity int        number for the log level verbosity (default 1)
```

### SEE ALSO

* [landscaper-cli components-cli component-archive resources](landscaper-cli_components-cli_component-archive_resources.md)	 - command to modify resources of a component descriptor

