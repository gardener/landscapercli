## landscaper-cli quickstart add-oci-endpoint

command to add an external https endpoint for the oci registry (experimental)

### Synopsis

This command add an external https endpoint to an oci registry installed with the Landscaper CLI quickstart command. This command is only supported for garden shoot clusters with activated nginx and a cert manager. Furthermore htpasswd must be installed on your local machine.

```
landscaper-cli quickstart add-oci-endpoint [flags]
```

### Examples

```
landscaper-cli quickstart add-oci-endpoint \
    --kubeconfig ./kubconfig.yaml \
    --namespace landscaper \
    --user testuser \
    --password sic7a5snk
```

### Options

```
  -h, --help                help for add-oci-endpoint
      --kubeconfig string   path to the kubeconfig of the target cluster
      --namespace string    namespace where the landscaper and the OCI registry are installed (default "landscaper")
      --password string     password for authentication at the oci endpoint
      --user string         user for authentication at the oci endpoint
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

* [landscaper-cli quickstart](landscaper-cli_quickstart.md)	 - useful commands for getting quickly up and running with the landscaper

