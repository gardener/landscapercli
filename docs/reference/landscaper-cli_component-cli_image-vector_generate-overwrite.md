## landscaper-cli component-cli image-vector generate-overwrite

Get parses a component descriptor and returns the defined image vector

### Synopsis


generate-overwrite parses images defined in a component descriptor and returns them as image vector.

Images can be defined in a component descriptor in 3 different ways:
1. as 'ociImage' resource: The image is defined a default resource of type 'ociImage' with a access of type 'ociRegistry'.
   It is expected that the resource contains the following labels to be identified as image vector image.
   The resulting image overwrite will contain the repository and the tag/digest from the access method.
<pre>
resources:
- name: pause-container
  version: "3.1"
  type: ociImage
  relation: external
  extraIdentity:
    "imagevector-gardener-cloud+tag": "3.1"
  labels:
  - name: imagevector.gardener.cloud/name
    value: pause-container
  - name: imagevector.gardener.cloud/repository
    value: gcr.io/google_containers/pause-amd64
  - name: imagevector.gardener.cloud/source-repository
    value: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  - name: imagevector.gardener.cloud/target-version
    value: "< 1.16"
  access:
    type: ociRegistry
    imageReference: gcr.io/google_containers/pause-amd64:3.1
</pre>

2. as component reference: The images are defined in a label "imagevector.gardener.cloud/images".
   The resulting image overwrite will contain all images defined in the images label.
   Their repository and tag/digest will be matched from the resources defined in the actual component's resources.

   Note: The images from the label are matched to the resources using their name and version. The original image reference do not exit anymore.

<pre>
componentReferences:
- name: cluster-autoscaler-abc
  componentName: github.com/gardener/autoscaler
  version: v0.10.1
  labels:
  - name: imagevector.gardener.cloud/images
    value:
      images:
      - name: cluster-autoscaler
        repository: eu.gcr.io/gardener-project/gardener/autoscaler/cluster-autoscaler
        tag: "v0.10.1"
</pre>

3. as generic images from the component descriptor labels.
   Generic images are images that do not directly result in a resource.
   They will be matched with another component descriptor that actually defines the images.
   The other component descriptor MUST have the "imagevector.gardener.cloud/name" label in order to be matched.

<pre>
meta:
  schemaVersion: 'v2'
component:
  labels:
  - name: imagevector.gardener.cloud/images
    value:
      images:
      - name: hyperkube
        repository: k8s.gcr.io/hyperkube
        targetVersion: "< 1.19"
</pre>

<pre>
meta:
  schemaVersion: 'v2'
component:
  resources:
  - name: hyperkube
    version: "v1.19.4"
    type: ociImage
    extraIdentity:
      "imagevector-gardener-cloud+tag": "v1.19.4"
    labels:
    - name: imagevector.gardener.cloud/name
      value: hyperkube
    - name: imagevector.gardener.cloud/repository
      value: k8s.gcr.io/hyperkube
    access:
	  type: ociRegistry
	  imageReference: my-registry/hyperkube:v1.19.4
</pre>



```
landscaper-cli component-cli image-vector generate-overwrite --component="example.com/my/component/name:v0.0.1 | /path/to/local/component-descriptor" -o IV_OVERWRITE_OUTPUT_PATH [--add-comp=ADDITIONAL_COMPONENT]... [flags]
```

### Options

```
      --add-comp stringArray       list of name and version of an additional component or a path to the local component descriptor. The component ref is expected to be of the format '<component-name>:<component-version>'
      --allow-plain-http           allows the fallback to http if the oci registry does not support https
      --cc-config string           path to the local concourse config file
  -c, --component string           name and version of the main component or a path to the local component descriptor. The component ref is expected to be of the format '<component-name>:<component-version>'
  -h, --help                       help for generate-overwrite
      --insecure-skip-tls-verify   If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
  -o, --output string              The path to the image vector that will be written.
      --registry-config string     path to the dockerconfig.json with the oci registry authentication information
      --repo-ctx string            base url of the component repository
      --resolve-tags               enable that tags are automatically resolved to digests
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

* [landscaper-cli component-cli image-vector](landscaper-cli_component-cli_image-vector.md)	 - command to add resource from a image vector and retrieve from a component descriptor

