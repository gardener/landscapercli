# Readme

## Helper

**Update and upload component** 

```
landscaper-cli components-cli component-archive resources add \
$LS_COMPONENT_DIR/demo-component \
-r $LS_COMPONENT_DIR/demo-component/resources.yaml

landscaper-cli components-cli ca remote push \
eu.gcr.io/sap-gcp-cp-k8s-stable-hub/examples/landscaper/temp \
github.com/gardener/landscapercli/examplecont \
v0.1.0 \
$LS_COMPONENT_DIR/demo-component`
```

**Open shell in pod:**

```
kubectl -n default exec --stdin --tty demo-installation-container1-7dv9v-9kld9 -- /bin/sh
```

## Todo

- Test export: echo "testparam: ttt" > $EXPORTS_PATH

- Documentation
