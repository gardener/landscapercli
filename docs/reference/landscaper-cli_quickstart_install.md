## landscaper-cli quickstart install

command to install Landscaper (including Container, Helm, and Manifest deployers) in a target cluster. An OCI registry for testing can be optionally installed

```
landscaper-cli quickstart install --kubeconfig [kubconfig.yaml] --landscaper-values [landscaper-values.yaml] --namespace landscaper --install-oci-registry --install-registry-ingress --registry-username testuser --registry-password some-pw [flags]
```

### Examples

```
landscaper-cli quickstart install --kubeconfig ./kubconfig.yaml --landscaper-values ./landscaper-values.yaml --namespace landscaper --install-oci-registry --install-registry-ingress --registry-username testuser --registry-password some-pw
```

### Options

```
  -h, --help                              help for install
      --install-oci-registry              install an OCI registry in the target cluster
      --install-registry-ingress          install an ingress for accessing the OCI registry. 
                                          the credentials must be provided via the flags "--registry-username" and "--registry-password".
                                          the Landscaper instance will then be automatically configured with these credentials.
                                          prerequisites (!):
                                           - the target cluster must be a Gardener Shoot (TLS is provided via the Gardener cert manager)
                                           - a nginx ingress controller must be deployed in the target cluster
                                           - the command "htpasswd" must be installed on your local machine
      --kubeconfig string                 path to the kubeconfig of the target cluster
      --landscaper-chart-version string   use a custom Landscaper chart version (corresponds to Landscaper Github release with the same version number) (default "v0.23.0")
      --landscaper-values string          path to values.yaml for the Landscaper Helm installation
      --namespace string                  namespace where Landscaper and the OCI registry will get installed (default "landscaper")
      --registry-password string          password for authenticating at the OCI registry
      --registry-username string          username for authenticating at the OCI registry
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

* [landscaper-cli quickstart](landscaper-cli_quickstart.md)	 - useful commands for getting quickly up and running with Landscaper

