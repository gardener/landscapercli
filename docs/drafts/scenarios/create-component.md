# Creating a Component with Landscaper CLI Commands

## Create Component Skeleton

The following command creates the skeleton for a component in the local file system.
You have to specify:

- an existing, empty directory, in which the component will be created,
- the name of the component,
- the version of the component.

```shell script
landscaper-cli component create \
    ./example-component \
    github.com/gardener/landscapercli/nginx \
    v0.1.0
```

## Add DeployItems

The following command adds a deployitem to the component. 
It is assumed that the helm chart is already stored as artifact in an oci registry.

```shell script
landscaper-cli component add helm-ls deployitem \ 
    ./example-component \
    nginx1 \
    nginx1 \
    --oci-reference eu.gcr.io/gardener-project/landscaper/tutorials/charts/ingress-nginx:v0.1.0 \
    --chart-version v0.1.0 \
    --cluster-param cluster \
    --target-ns-param nginx-ns1
```

The following command adds a second deployitem.
It is assumed that the helm chart is already stored in the local file system.

```shell script
landscaper-cli component add helm-ls deployitem \
    ./example-component \
    nginx2 \
    nginx2 \
    --chart-directory ./chart/ingress-nginx \
    --chart-version v0.1.0 \
    --cluster-param cluster \
    --target-ns-param nginx-ns2
```

## Transfer Resources

Up to now, three resource items have been collected in the resources.yaml file: the blueprint and the two charts.
The following command transfers them into the component-descriptor.yaml. It also collects some of the resources as blobs
in a `blobs` subdirectory of the component directory. This is done for the blueprint and helm chart in the local filesystem,
but not for the external helm chart. This is the preparation for the push in the next step. 

```shell script
landscaper-cli components-cli ca resources add \
    ./example-component \
    -r ./example-component/resources.yaml
```

## Push Component

Manually maintain the field `component.repositoryContext.baseUrl` in the component-descriptor.yaml.

Then use the following command to push the component to the oci registry.

```shell script
landscaper-cli components-cli ca remote push \
    ./example-component
```

After the component has been pushed, the oci registry contains an artifact in repository
`component-descriptors/github.com/gardener/landscapercli/nginx` (i.e. `component-descriptors/` + component name)
and with tag `v0.1.0` (the component version). The oci artifact has three layers, namely the component descriptor and
the blobs prepared in the previous step, i.e. the blueprint and the first of the two charts.

## Access the OCI Artifact

The OCI API allows us to acces the artifact that we have just pushed.

##### Get Repositories

```shell script
curl --location --request GET 'http://localhost:5000/v2/_catalog'
```

##### Get Tags

```shell script
curl --location --request GET 'http://localhost:5000/v2/component-descriptors/github.com/gardener/landscapercli/nginx/tags/list'
```

##### Get Manifest

```shell script
curl --location --request GET 'http://localhost:5000/v2/component-descriptors/github.com/gardener/landscapercli/nginx/manifests/v0.1.0' \
--header 'Accept: application/vnd.oci.image.manifest.v1+json'
```

##### Get Blob

```shell script
curl --location --request GET 'http://localhost:5000/v2/component-descriptors/github.com/gardener/landscapercli/nginx/blobs/sha256:445bc8e9074bc33cbaa15f127bda35ecfbb07bc20f87e5aa3f8d8829d5a9f0e4'
```
