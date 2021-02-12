# Ingress for OCI registry

- in command quickstart:
  - new flags
    - add-ingress: only for garden shoot cluster with nginx and cert manager, htpasswd must be installed,
    provide endpoint for 
    - user/passwd
  - return endpoint for registry, e.g. https://oci.ingress.rg.hubtest.shoot.dev.k8s-hana.ondemand.com/

- Create ingress:

```yaml
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  annotations:
    cert.gardener.cloud/purpose: managed
    dns.gardener.cloud/class: garden
    dns.gardener.cloud/dnsnames: # oci.ingress.rg.hubtest.shoot.dev.k8s-hana.ondemand.com
    nginx.ingress.kubernetes.io/auth-type: basic
    nginx.ingress.kubernetes.io/auth-secret: oci-secret
    nginx.ingress.kubernetes.io/auth-realm: 'Authentication Required'
  name: oci-ingress
  namespace: # landscaper
spec:
  rules:
  - host: # oci.ingress.rg.hubtest.shoot.dev.k8s-hana.ondemand.com
    http:
      paths:
      - backend:
          serviceName: oci-registry
          servicePort: 5000
        path: /
  tls:
  - hosts:
    - # oci.ingress.rg.hubtest.shoot.dev.k8s-hana.ondemand.com
    secretName: oci.tls-secret
```

- create secret for ingress

```shell
$ htpasswd -c auth testuser
New password: <bar>
New password:
Re-type new password:
Adding password for user testuser

$ kubectl create secret oci-secret basic-auth --from-file=auth
secret "basic-auth" created
```

- Remove all stuff with quickstart uninstall

- Test with curl --location --request GET https://oci.ingress.rg.hubtest.shoot.dev.k8s-hana.ondemand.com/v2/_catalog -u "admin:test"
