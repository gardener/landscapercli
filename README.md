# Landscaper CLI

The Landscaper CLI supports users to develop, maintain, and test components processed by the 
[Landscaper](https://github.com/gardener/landscaper). This comprises the handling of objects like component descriptors, 
blueprints, installations, etc. 

The Landscaper CLI integrates the commands of the [Components CLI](https://github.com/gardener/component-cli).

The Landscaper CLI adds commands for the following use cases:

- Automatic setup of a landscaper and an OCI registry for development  
- Development of components including a component descriptor and blueprint  
- Local validation of such components  
- Support for testing components on a kubernetes cluster  
- Support for accessing and maintaining components with a blueprint in OCI registry

## Installation

Installation instructions can be found [here](docs/installation.md).

## Documentation 

Detailed descriptions for commands could be found [here](docs/commands).

The command reference is located [here](docs/reference/landscaper-cli.md). 

The Landscaper CLI support the installation of the [Docker OCI registry](https://hub.docker.com/_/registry/) 
with the [quickstart command](docs/commands/quickstart).

A description how to access an OCI registry which requires authentication with the Landscaper CLI can be found [here](docs/login-to-oci-registry.md). 

Other examples:
https://github.com/gardener/landscaper/blob/master/docs/tutorials/02-simple-import.md 
