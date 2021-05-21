## landscaper-cli components-cli image-vector add

Adds all resources of a image vector to the component descriptor

### Synopsis


add parses a image vector and generates or enhances the corresponding component descriptor resources.

There are 4 different scenarios how images are added to the component descriptor.
1. The image is defined with a tag and will be directly translated as oci image resource.

<pre>
images:
- name: pause-container
  sourceRepository: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  repository: gcr.io/google_containers/pause-amd64
  tag: "3.1"
</pre>

<pre>
meta:
  schemaVersion: 'v2'
...
resources:
- name: pause-container
  version: "3.1"
  type: ociImage
  extraIdentity:
    "imagevector-gardener-cloud+tag": "3.1"
  labels:
  - name: imagevector.gardener.cloud/name
    value: pause-container
  - name: imagevector.gardener.cloud/repository
    value: gcr.io/google_containers/pause-amd64
  - name: imagevector.gardener.cloud/source-repository
    value: github.com/kubernetes/kubernetes/blob/master/build/pause/Dockerfile
  access:
    type: ociRegistry
    imageReference: gcr.io/google_containers/pause-amd64:3.1
</pre>

2. The image is defined by another component so the image is added as label ("imagevector.gardener.cloud/images") to the "componentReference".

Images that are defined by other components can be specified 
1. when the image's repository matches the given "--component-prefixes"
2. the image is labeled with "imagevector.gardener.cloud/component-reference"

If the component reference is not yet defined it will be automatically added.
If multiple images are defined for the same component reference they are added to the images list in the label.

<pre>
images:
- name: cluster-autoscaler
  sourceRepository: github.com/gardener/autoscaler
  repository: eu.gcr.io/gardener-project/gardener/autoscaler/cluster-autoscaler
  targetVersion: "< 1.16"
  tag: "v0.10.0"
  labels: # recommended bbut only needed when "--component-prefixes" is not defined
  - name: imagevector.gardener.cloud/component-reference
    value:
      name: cla # defaults to image.name
      componentName: github.com/gardener/autoscaler # defaults to image.sourceRepository
      version: v0.10.0 # defaults to image.version
</pre>

<pre>
meta:
  schemaVersion: 'v2'
...
componentReferences:
- name: cla
  componentName: github.com/gardener/autoscaler
  version: v0.10.0
  extraIdentity:
    imagevector-gardener-cloud+tag: v0.10.0
  labels:
  - name: imagevector.gardener.cloud/images
    value:
	  images:
	  - name: cluster-autoscaler
	    repository: eu.gcr.io/gardener-project/gardener/autoscaler/cluster-autoscaler
	    sourceRepository: github.com/gardener/autoscaler
	    tag: v0.10.0
	    targetVersion: '< 1.16'
</pre>

3. The image is a generic dependency where the actual images are defined by the overwrite.
A generic dependency image is not part of a component descriptor's resource but will be added as label ("imagevector.gardener.cloud/images") to the component descriptor. 

Generic dependencies can be defined by
1. defined as "--generic-dependency=<image name>"
2. the label "imagevector.gardener.cloud/generic"

<pre>
images:
- name: hyperkube
  sourceRepository: github.com/kubernetes/kubernetes
  repository: k8s.gcr.io/hyperkube
  targetVersion: "< 1.19"
  labels: # only needed if "--generic-dependency" is not set
  - name: imagevector.gardener.cloud/generic
</pre>

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
	    sourceRepository: github.com/kubernetes/kubernetes
	    targetVersion: '< 1.19'
</pre>

4. The image has not tag and it's repository matches a already defined resource in the component descriptor.
This usually means that the image is build as part of the build pipeline and the version depends on the current component.
In this case only labels are added to the existing resource

<pre>
images:
- name: gardenlet
  sourceRepository: github.com/gardener/gardener
  repository: eu.gcr.io/gardener-project/gardener/gardenlet
</pre>

<pre>
meta:
  schemaVersion: 'v2'
...
resources:
- name: gardenlet
  version: "v0.0.0"
  type: ociImage
  relation: local
  labels:
  - name: imagevector.gardener.cloud/name
    value: gardenlet
  - name: imagevector.gardener.cloud/repository
    value: eu.gcr.io/gardener-project/gardener/gardenlet
  - name: imagevector.gardener.cloud/source-repository
    value: github.com/gardener/gardener
  access:
    type: ociRegistry
    imageReference: eu.gcr.io/gardener-project/gardener/gardenlet:v0.0.0
</pre>



```
landscaper-cli components-cli image-vector add --comp-desc component-descriptor-file --image-vector images.yaml [--component-prefixes "github.com/gardener/myproj"]... [--generic-dependency image-source-name]... [--generic-dependencies "image-name1,image-name2"] [flags]
```

### Options

```
      --comp-desc string                          path to the component descriptor directory
      --component-prefixes stringArray            Specify all prefixes that define a image  from another component
      --exclude-component-reference stringArray   Specify all image name that should not be added as component reference
      --generic-dependencies string               Specify all prefixes that define a image  from another component
      --generic-dependency stringArray            Specify all image source names that are a generic dependency.
  -h, --help                                      help for add
      --image-vector string                       The path to the resources defined as yaml or json
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

