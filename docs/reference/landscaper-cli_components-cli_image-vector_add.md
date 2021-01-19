## landscaper-cli components-cli image-vector add

Adds all resources of a image vector to the component descriptor

### Synopsis


add parses a image vector and generates the corresponding component descriptor resources.

<pre>

images:
- name: pause-container
  sourceRepository: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  repository: gcr.io/google_containers/pause-amd64
  tag: "3.1"

</pre>


```
landscaper-cli components-cli image-vector add --comp-desc component-descriptor-file --image-vector images.yaml [--component-prefixes "github.com/gardener/myproj"]... [--generic-dependency image-source-name]... [--generic-dependencies "image-name1,image-name2"] [flags]
```

### Options

```
      --comp-desc string                 path to the component descriptor directory
      --component-prefixes stringArray   Specify all prefixes that define a image  from another component
      --generic-dependencies string      Specify all prefixes that define a image  from another component
      --generic-dependency stringArray   Specify all image source names that are a generic dependency.
  -h, --help                             help for add
      --image-vector string              The path to the resources defined as yaml or json
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

* [landscaper-cli components-cli image-vector](landscaper-cli_components-cli_image-vector.md)	 - command to add resource from a image vector and retrieve from a component descriptor

