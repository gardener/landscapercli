## landscaper-cli componentdescriptor push

command to interact with a component descriptor stored an oci registry

```
landscaper-cli componentdescriptor push [flags]
```

### Examples

```
landscapercli cd push [path to component descriptor]
- The cli will read all necessary parameters from the component descriptor.

landscapercli cd push [baseurl] [componentname] [version] [path to component descriptor]
- The cli will add the baseurl as repository context and validate the name and version.

```

### Options

```
      --allow-plain-http   allows the fallback to http if the oci registry does not support https
  -h, --help               help for push
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

* [landscaper-cli componentdescriptor](landscaper-cli_componentdescriptor.md)	 - command to interact with component descriptors stored in an oci registry

