# Installation

Prerequisites: 
  - Golang 1.15 or higher

The Landscaper CLI can be installed by cloning the landscaper-cli github repository and call the following
commands in the root folder of the repo: 

```shell script
export GO111MODULE=on

# install binary
make install-cli
```

The executable *landscaper-cli* is added to your *...go/bin* folder. Make sure that the *...go/bin* path 
is added to your `$PATH` env var: `export PATH=$PATH:$GOPATH/bin`

**Attention:** Currently the installation via `go get ...` does not work due to dependency problems. 
If this problem is resolved also the installation with the following commands is possible:

```shell script
export GO111MODULE=on

go get github.com/gardener/landscapercli/landscaper-cli

# or with a specific version
go get github.com/gardener/landscapercli/landscaper-cli@v0.0.1

```
