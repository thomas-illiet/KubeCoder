# KubeCoder backend

The backend is a Go HTTP API backed by PostgreSQL. Its first business endpoint is `GET /api/v1/users/me`, authenticated with an OIDC Bearer access token.

## Local commands

Copy `config.example.yaml`, adjust PostgreSQL and OIDC values, then run:

```sh
go run ./cmd/kubecoder --config ./config.local.yaml migrate up
go run ./cmd/kubecoder --config ./config.local.yaml serve
```

Swagger UI is served at <http://localhost:8080/docs/> and the OpenAPI 3 document at <http://localhost:8080/openapi.yaml>.

## Package layout

The HTTP transport is split into subpackages so API growth does not create a flat directory:

- `internal/api/server`: route composition and embedded OpenAPI documentation;
- `internal/api/health`: liveness, readiness, and schema checks;
- `internal/api/users`: user HTTP handlers;
- `internal/api/httpx`: shared HTTP middleware and response helpers;
- `internal/users`: user model, service, and PostgreSQL repository.

New API categories must be added as their own `internal/api/<category>` package and registered explicitly by `internal/api/server`.

Available commands:

```text
kubecoder serve
kubecoder migrate up
kubecoder migrate down [steps]
kubecoder migrate version
kubecoder migrate wait --version 1
kubecoder version
```

Configuration accepts YAML, `KUBECODER_*` environment variables, and command flags. `oidc.issuer` is always the issuer that must be present in tokens. `oidc.discovery_url` may point to a trusted internal backchannel when the public issuer is not reachable from the API Pod.

## Database migrations

The GORM models are the desired database schema. Atlas compares them with the
existing versioned migrations and generates SQL files in the format consumed by
`golang-migrate`:

```sh
brew install ariga/tap/atlas
make atlas-schema
make atlas-diff NAME=add_organizations
make atlas-validate
```

Atlas uses an ephemeral PostgreSQL 17 development container when calculating a
schema diff, so Docker must be running. Always review both generated `up.sql`
and `down.sql` files. After generating a migration, update
`migration.LatestVersion` and the Helm value `migration.expectedVersion` to the
generated numeric version. Runtime migrations remain embedded in `kubecoder`
and are executed by `golang-migrate`; Atlas is a development-time generator and
validator only.

## Validation

```sh
go test -race ./...
go vet ./...
go build ./cmd/kubecoder
```

Set `KUBECODER_TEST_DATABASE_DSN` to a migrated, disposable PostgreSQL database to run the concurrent first-user provisioning integration test.
