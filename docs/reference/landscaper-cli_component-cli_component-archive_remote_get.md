## landscaper-cli component-cli component-archive remote get

fetch the component descriptor from a oci registry

### Synopsis


get fetches the component descriptor from a baseurl with the given name and Version.


```
landscaper-cli component-cli component-archive remote get BASE_URL COMPONENT_NAME VERSION [flags]
```

### Options

```
      --allow-plain-http                allows the fallback to http if the oci registry does not support https
      --cc-config string                path to the local concourse config file
      --component-name-mapping string   [OPTIONAL] repository context name mapping (default "urlPath")
  -h, --help                            help for get
      --insecure-skip-tls-verify        If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --registry-config string          path to the dockerconfig.json with the oci registry authentication information
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

* [landscaper-cli component-cli component-archive remote](landscaper-cli_component-cli_component-archive_remote.md)	 - command to interact with component descriptors stored in an oci registry

