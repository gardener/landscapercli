## landscaper-cli component-cli oci pull

Pulls a oci artifact from a registry

### Synopsis


Pull downloads the specified oci artifact from a registry.
If no output directory is specified, the blob is written to stdout.

If a blob digest is given, the artifact will download the specific blob.
If no blob is given the whole artifact is downloaded and written to a directory.
If no output directory is specified, the artifact manifest is written to stdout.



```
landscaper-cli component-cli oci pull ARTIFACT_REFERENCE [config | blob digest] [flags]
```

### Options

```
      --allow-plain-http           allows the fallback to http if the oci registry does not support https
      --cc-config string           path to the local concourse config file
  -h, --help                       help for pull
      --insecure-skip-tls-verify   If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
  -O, --output-dir string          specifies the output where the artifact should be written.
      --registry-config string     path to the dockerconfig.json with the oci registry authentication information
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

* [landscaper-cli component-cli oci](landscaper-cli_component-cli_oci.md)	 - 

