## landscaper-cli component-cli component-archive signatures sign signing-server

fetch the component descriptor from an oci registry, sign it with a signature provided from a signing server, and re-upload

```
landscaper-cli component-cli component-archive signatures sign signing-server BASE_URL COMPONENT_NAME VERSION [flags]
```

### Options

```
      --allow-plain-http            allows the fallback to http if the oci registry does not support https
      --cc-config string            path to the local concourse config file
      --client-cert string          [OPTIONAL] path to a file containing the client certificate in PEM format for authenticating to the server
      --force                       [OPTIONAL] force overwrite of already existing component descriptors
  -h, --help                        help for signing-server
      --insecure-skip-tls-verify    If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --private-key string          [OPTIONAL] path to a file containing the private key for the provided client certificate in PEM format
      --recursive                   [OPTIONAL] recursively sign and upload all referenced component descriptors
      --registry-config string      path to the dockerconfig.json with the oci registry authentication information
      --root-ca-certs string        [OPTIONAL] path to a file containing additional root ca certificates in PEM format. if empty, the system root ca certificate pool is used
      --server-url string           url where the signing server is running, e.g. https://localhost:8080
      --signature-name string       name of the signature
      --skip-access-types strings   [OPTIONAL] comma separated list of access types that will not be digested and signed
      --upload-base-url string      target repository context to upload the signed cd
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

* [landscaper-cli component-cli component-archive signatures sign](landscaper-cli_component-cli_component-archive_signatures_sign.md)	 - command to sign component descriptors

