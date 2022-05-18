## landscaper-cli component-cli component-archive signature verify

command to verify the signature of a component descriptor

### Options

```
  -h, --help   help for verify
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

* [landscaper-cli component-cli component-archive signature](landscaper-cli_component-cli_component-archive_signature.md)	 - [EXPERIMENTAL] command to work with signatures and digests in component descriptors
* [landscaper-cli component-cli component-archive signature verify rsa](landscaper-cli_component-cli_component-archive_signature_verify_rsa.md)	 - fetch the component descriptor from an oci registry and verify its integrity based on a RSASSA-PKCS1-V1_5-SIGN signature
* [landscaper-cli component-cli component-archive signature verify x509](landscaper-cli_component-cli_component-archive_signature_verify_x509.md)	 - fetch the component descriptor from an oci registry and verify its integrity based on a x509 certificate chain and a RSASSA-PKCS1-V1_5 signature

