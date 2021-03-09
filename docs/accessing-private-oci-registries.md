# Accessing Private OCI Registries

When Landscaper deploys a component, it accesses OCI registries to read components, blueprints, helm charts, images, etc. 
To simplify the setup, it is recommended to store all artifacts of a component in one registry.
If such an OCI registry is not public, you have to

- provide the access data in a secret, called *registry pull secret*,    
- add a reference to the registry pull secret to the installation.  

## Creating a Registry Pull Secret

Create the registry pull secret on the same kubernetes cluster, in the same namespace as the installation.

You can use the following command as described in 
[Create a Secret by providing credentials on the command line](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/#create-a-secret-by-providing-credentials-on-the-command-line)

```
kubectl create secret docker-registry [secret name] -n [secret namespace] \
--docker-server=[URL of the https endpoint of the OCI registry] \
--docker-username=[user] \
--docker-password=[password]
```

## Referencing a Registry Pull Secret

The component is deployed by Landscaper if you create an installation. In the installation, you add a reference to
the registry pull secret. 

```yaml
apiVersion: landscaper.gardener.cloud/v1alpha1
kind: Installation
...
spec:
  componentDescriptor:
    ref:
      repositoryContext:
        type: ociRegistry
        baseUrl: [base URL of the OCI registry]
      componentName: [component name]
      version: [component version]
  
  registryPullSecrets:
  - name: [secret name]
    namespace: [secret namespace]
...
```

The registry pull secret will be used to read the component descriptor and blueprint. In addition, deployers
can use the registry pull secret to read helm charts, images, etc. from the OCI registry.
