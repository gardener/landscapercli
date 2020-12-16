# Landscaper CLI

The Landscaper CLI tool supports users to interact with [Landscaper](https://github.com/gardener/landscaper) 
installations and their custom resources like Installations, Executions, DeployItems, DataObjects, Targets etc. These 
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

## Work with an OCI Registry

In order for the `landscaper-cli` to work with the registry, it needs valid credentials. The easiest way to generate 
these, would be via `docker login`. 

```shell
docker login ...
```

An example how to work with the Landscaper Cli and an [Harbor OCI registry](https://github.com/goharbor/harbor-helm) 
could be found [here](https://github.com/gardener/landscaper/blob/master/docs/tutorials/00-local-setup.md).

Other examples:
https://github.com/gardener/landscaper/blob/master/docs/tutorials/02-simple-import.md 

