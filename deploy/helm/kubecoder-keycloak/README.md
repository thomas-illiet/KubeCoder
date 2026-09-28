# KubeCoder local Keycloak

This chart installs a local-only Keycloak 26.7.4 instance backed by PostgreSQL. It targets the Docker Desktop Kubernetes cluster and exposes Keycloak at `http://localhost:30080`.

The credentials in `values.yaml` are deliberately fixed for local development. Never reuse them in a shared or production environment.

## Install or upgrade

```sh
helm upgrade --install kubecoder-auth ./deploy/helm/kubecoder-keycloak \
  --namespace kubecoder-auth \
  --create-namespace \
  --wait \
  --timeout 10m
```

The idempotent Helm hook provisions realm `kubecoder`, public client `kubecoder-web`, and the demo account:

Access tokens issued to `kubecoder-web` include the `kubecoder-api` audience required by the local backend.

- user: `admin`
- password: `admin`

The application realm and the administration console both use `admin` / `admin` for this local demo.

## Verify

```sh
kubectl get pods,pvc,svc,jobs -n kubecoder-auth
curl --fail http://localhost:30080/realms/kubecoder/.well-known/openid-configuration
```

## Uninstall

```sh
helm uninstall kubecoder-auth --namespace kubecoder-auth
```

The StatefulSet PVC is retained by Kubernetes after uninstall. Delete it explicitly only when a full local identity reset is intended.
