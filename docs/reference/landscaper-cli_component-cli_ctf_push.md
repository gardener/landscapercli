## landscaper-cli component-cli ctf push

Pushes all archives of a ctf to a remote repository

### Synopsis


Push pushes all component archives and oci artifacts to the defined oci repository.

The oci repository is automatically determined based on the component/artifact descriptor (repositoryContext, component name and version).

Note: Currently only component archives are supoprted. Generic OCI Artifacts will be supported in the future.


```
landscaper-cli component-cli ctf push CTF_PATH [flags]
```

### Options

```
      --allow-plain-http           allows the fallback to http if the oci registry does not support https
      --cc-config string           path to the local concourse config file
  -h, --help                       help for push
      --insecure-skip-tls-verify   If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --registry-config string     path to the dockerconfig.json with the oci registry authentication information
      --repo-ctx string            repository context url for component to upload. The repository url will be automatically added to the repository contexts.
  -t, --tag stringArray            set additional tags on the oci artifact
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

* [landscaper-cli component-cli ctf](landscaper-cli_component-cli_ctf.md)	 - 

