# Login to an OCI Registry

Some commands of the Landscaper CLI access an OCI registry. If the OCI registry requires credentials, access for the Landscaper CLI can be 
provided by executing the [docker login command](https://docs.docker.com/engine/reference/commandline/login/):

```
docker login
```

The command stores required credentials on your local machine which are used by the Landscaper CLI afterwards.
