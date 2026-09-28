# KubeCoder API chart

This chart installs the KubeCoder API, a schema migration Job, and optionally a single-node PostgreSQL 17 StatefulSet. The bundled database is intended for local development and early environments, not high availability.

## Build the local image

Docker Desktop Kubernetes can use an image built by the local Docker daemon:

```sh
docker build -t kubecoder:dev ./backend
```

## Install or upgrade

Install Keycloak first, then KubeCoder:

```sh
helm upgrade --install kubecoder ./deploy/helm/kubecoder \
  --namespace kubecoder-system \
  --create-namespace \
  --wait \
  --wait-for-jobs \
  --timeout 10m
```

The API is available at <http://localhost:30081>, and Swagger UI at <http://localhost:30081/docs/>.

## Verify

```sh
kubectl get pods,jobs,pvc,svc -n kubecoder-system
kubectl logs -n kubecoder-system job/kubecoder-kubecoder-migrate-1
curl --fail http://localhost:30081/healthz
curl --fail http://localhost:30081/readyz
```

Use a Keycloak access token with the API:

```sh
curl --fail --header "Authorization: Bearer $ACCESS_TOKEN" http://localhost:30081/api/v1/users/me
```

## External PostgreSQL

Create a Secret whose `dsn` key contains the complete PostgreSQL URL, then install with:

```sh
helm upgrade --install kubecoder ./deploy/helm/kubecoder \
  --namespace kubecoder-system --create-namespace \
  --values ./deploy/helm/kubecoder/values-external-database.yaml
```

When `postgresql.auth.existingSecret` is used with the bundled StatefulSet, that Secret must provide `username`, `password`, `database`, and `dsn` keys, or the corresponding key names configured in values.

## Rollback and uninstall

```sh
helm rollback kubecoder REVISION --namespace kubecoder-system
helm uninstall kubecoder --namespace kubecoder-system
```

Schema migrations are not automatically reversed during a Helm rollback. The PostgreSQL PVC is retained by Kubernetes after uninstall; remove it explicitly only when a full local data reset is intended.
