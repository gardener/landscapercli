## landscaper-cli component-cli component-archive signatures verify rsa

fetch the component descriptor from an oci registry and verify its integrity based on a RSASSA-PKCS1-V1_5-SIGN signature

```
landscaper-cli component-cli component-archive signatures verify rsa BASE_URL COMPONENT_NAME VERSION [flags]
```

### Options

```
      --allow-plain-http           allows the fallback to http if the oci registry does not support https
      --cc-config string           path to the local concourse config file
  -h, --help                       help for rsa
      --insecure-skip-tls-verify   If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --public-key string          path to public key file
      --registry-config string     path to the dockerconfig.json with the oci registry authentication information
      --signature-name string      name of the signature to verify
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

* [landscaper-cli component-cli component-archive signatures verify](landscaper-cli_component-cli_component-archive_signatures_verify.md)	 - command to verify the signature of a component descriptor

