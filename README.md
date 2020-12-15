# Landscaper CLI

**Under Construction**

The landscaper uses custom resources like Installations, Executions, DeployItems, DataObjects, Targets. These 
resources can simply be accessed using `kubectl`.

The landscaper also interacts with resources that are not stored in a cluster.
Some of these resources include Blueprints, ComponentDescriptors or jsonschemas that are stored remotely in an OCI 
registry.

The Landscaper CLI tool is mainly build to support human users interacting with these remote resources.
We may also think to improve the kubectl experience but this will then be rather a kubectl plugin than its own cli tool.
(ref https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)

## Installation

The landscaper cli can be installed via go:

```shell script
go get github.com/gardener/landscapercli/landscaper-cli

# or with a specific version
go get github.com/gardener/landscapercli/landscaper-cli@v0.1.0
```
Make sure that the go bin path is set in your `$PATH` env var: `export PATH=$PATH:$GOPATH/bin`
