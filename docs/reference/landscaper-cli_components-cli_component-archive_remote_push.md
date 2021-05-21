## landscaper-cli components-cli component-archive remote push

pushes a component archive to an oci repository

### Synopsis


pushes a component archive with the component descriptor and its local blobs to an oci repository.

The command can be called in 2 different ways:

push [path to component descriptor]
- The cli will read all necessary parameters from the component descriptor.

push [baseurl] [componentname] [Version] [path to component descriptor]
- The cli will add the baseurl as repository context and validate the name and Version.


```
landscaper-cli components-cli component-archive remote push [path to component descriptor] [flags]
```

### Options

```
      --allow-plain-http         allows the fallback to http if the oci registry does not support https
      --cc-config string         path to the local concourse config file
  -h, --help                     help for push
      --registry-config string   path to the dockerconfig.json with the oci registry authentication information
      --repo-ctx string          repository context url for component to upload. The repository url will be automatically added to the repository contexts.
  -t, --tag stringArray          set additional tags on the oci artifact
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

* [landscaper-cli components-cli component-archive remote](landscaper-cli_components-cli_component-archive_remote.md)	 - command to interact with component descriptors stored in an oci registry

