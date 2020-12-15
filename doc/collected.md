# Docu Collection

This is just a collection of documentation which need to be structured

## Sample

#### Working with the landscaper-cli
In order for the `landscaper-cli` to work with the registry, it needs valid credentials. The easiest way to generate these, would be via `docker login`. Here is an example for a registry accessible via port-forwarding:
```shell
docker login -u my-user localhost:5000 # use the user name and pwd as specified in the harbor chart values.yaml
```

Later, when dealing with artifacts like the component descriptor, be aware that the URLs used to push and access the artifacts differ due to the port-forwarding. Make sure the base URL points to the cluster-internal representation of the registry:

```yaml
  repositoryContexts:
  - type: ociRegistry
    baseUrl: harbor-harbor-registry.harbor:5000/comp
```
But push explicitly to `localhost` instead implicitly using the baseUrl:

```shell
landscaper-cli  componentdescriptor push localhost:5000/comp/ github.com/gardener/landscaper/ingress-nginx v0.1.0 component-descriptor.yaml
```

## Other samples

https://github.com/gardener/landscaper/blob/master/docs/tutorials/01-create-simple-blueprint.md
https://github.com/gardener/landscaper/blob/master/docs/tutorials/02-simple-import.md