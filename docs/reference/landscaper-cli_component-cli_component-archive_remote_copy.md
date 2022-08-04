## landscaper-cli component-cli component-archive remote copy

copies a component descriptor from a context repository to another

### Synopsis


copies a component descriptor and its blobs from the source repository to the target repository.

By default the component descriptor and all its component references are recursively copied.
This behavior can be overwritten by specifying "--recursive=false"



```
landscaper-cli component-cli component-archive remote copy COMPONENT_NAME VERSION --from SOURCE_REPOSITORY --to TARGET_REPOSITORY [flags]
```

### Options

```
      --allow-plain-http                    allows the fallback to http if the oci registry does not support https
      --backoff-factor duration             a backoff factor to apply between retry attempts: backoff = backoff-factor * 2^retries. e.g. if backoff-factor is 1s, then the timeouts will be [1s, 2s, 4s, …] (default 1s)
      --cc-config string                    path to the local concourse config file
      --copy-by-value                       [EXPERIMENTAL] copies all referenced oci images and artifacts by value and not by reference.
      --force                               Forces the tool to overwrite already existing component descriptors.
      --from string                         source repository base url.
  -h, --help                                help for copy
      --insecure-skip-tls-verify            If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --keep-source-repository              Keep the original source repository when copying resources.
      --max-retries uint                    maximum number of retries for copying a component descriptor
      --recursive                           Recursively copy the component descriptor and its references. (default true)
      --registry-config string              path to the dockerconfig.json with the oci registry authentication information
      --relative-urls                       converts all copied oci artifacts to relative urls
      --replace-oci-ref strings             list of replace expressions in the format left:right. For every resource with accessType == ociRegistry, all occurences of 'left' in the target ref are replaced with 'right' before the upload
      --source-artifact-repository string   source repository where realtiove oci artifacts are copied from. This is only relevant if artifacts are copied by value and it will be defaulted to the source component repository
      --target-artifact-repository string   target repository where the artifacts are copied to. This is only relevant if artifacts are copied by value and it will be defaulted to the target component repository
      --to string                           target repository where the components are copied to.
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

