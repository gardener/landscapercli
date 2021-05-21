## landscaper-cli components-cli oci tags

Lists all tags of artifact reference

### Synopsis


tags lists all tags for a specific artifact reference that is known by the registry.



```
landscaper-cli components-cli oci tags ARTIFACT_REFERENCE [flags]
```

### Options

```
      --allow-plain-http         allows the fallback to http if the oci registry does not support https
      --cc-config string         path to the local concourse config file
  -h, --help                     help for tags
      --registry-config string   path to the dockerconfig.json with the oci registry authentication information
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

* [landscaper-cli components-cli oci](landscaper-cli_components-cli_oci.md)	 - 

