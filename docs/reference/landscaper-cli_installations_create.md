## landscaper-cli installations create

create an installation template for a component which is stored in an OCI registry

```
landscaper-cli installations create [baseURL] [componentName] [componentVersion] [flags]
```

### Examples

```
landscaper-cli installations create my-registry:5000 github.com/my-component v0.1.0
```

### Options

```
      --allow-plain-http                 allows the fallback to http if the oci registry does not support https
      --blueprint-resource-name string   name of the blueprint resource in the component descriptor (optional if only one blueprint resource is specified in the component descriptor)
      --cc-config string                 path to the local concourse config file
  -h, --help                             help for create
      --name string                      name of the installation (default "my-installation")
      --registry-config string           path to the dockerconfig.json with the oci registry authentication information
      --render-schema-info               render schema information of the component's imports and exports as comments into the installation (default true)
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

* [landscaper-cli installations](landscaper-cli_installations.md)	 - commands to interact with installations

