# auth-service

A gRPC authentication / SSO service written in Go. It handles user
registration with email confirmation, password-based login with JWT
access/refresh tokens, password reset and change, profile updates,
soft-deletion and account restore.

## Overview

- **Transport:** gRPC (`Auth` service, default port `50051`)
- **Storage:** PostgreSQL (schema is created automatically via migrations on startup)
- **Token / temp-code cache:** Memcached (email-confirmation, password-reset and
  account-restore tokens)
- **Auth tokens:** RS256 JWT signed with an RSA keypair (separate access and
  refresh tokens, configurable TTL)
- **Email:** SMTP, for sending confirmation / reset / restore links
- **Config:** environment variables (see `.env.example`)
- **Logging:** structured logging via `zap`

### gRPC API (`pkg/proto/sso.proto`)

`Register`, `AcceptEmail`, `Login`, `RefreshToken`, `Auth`, `Update`,
`RequestPasswordReset`, `ConfirmPasswordReset`, `ChangePassword`,
`DeleteUser`, `ResendConfirmationEmail`, `GetMe`.

### Project layout

```
cmd/app            entrypoint (wires config, db, cache, smtp, grpc server)
internal/config    environment-based configuration
internal/controller/grpc   gRPC handlers
internal/service/production service/business logic
internal/repository/store  PostgreSQL access
internal/cache/memcached   token cache
internal/smtp      mailer
internal/pkg/provider      JWT issuing / validation
pkg/proto, pkg/gen proto definition and generated code
migration          SQL migrations (run automatically on startup)
infra/keys         RSA keypair (generated, git-ignored)
```

## Prerequisites

- Go 1.25+
- `make` and `openssl` (for the bootstrap target)
- Docker + Docker Compose (for the containerized setup)
- A reachable PostgreSQL, Memcached and SMTP server (provided by
  `docker-compose.yml`, except SMTP)

## Configuration

Configuration is read from environment variables. Copy the example file and
adjust as needed:

```sh
cp .env.example .env
```

Key variables (defaults in parentheses):

| Variable | Description |
|---|---|
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | PostgreSQL connection (`localhost:5432`, `auth`/`auth_password`/`auth`) |
| `CACHE_HOST` / `CACHE_PORT` | Memcached (`localhost:11211`) |
| `SMTP_HOST` / `SMTP_PORT` / `SMTP_USERNAME` / `SMTP_PASSWORD` / `SMTP_FROM` | SMTP server (`localhost:1025`) |
| `PROVIDER_PRIVATE_KEY_PATH` / `PROVIDER_PUBLIC_KEY_PATH` | RSA keypair paths (`infra/keys/*.pem`) |
| `PROVIDER_ACCESS_TOKEN_TTL` / `PROVIDER_REFRESH_TOKEN_TTL` | Token lifetimes in seconds (`86400` / `2592000`) |
| `GRPC_HOST` / `GRPC_PORT` | gRPC listen address (`0.0.0.0:50051`) |
| `HTTP_SCHEME` / `HTTP_HOST` / `HTTP_PORT` | Base URL used to build email links (`http://localhost:8080`) |

The RSA keypair is **not** committed. Generate it with `make keys` (or
`make init`).

## Run with Docker Compose (recommended)

This starts PostgreSQL, Memcached and the service. Note: an SMTP server is
**not** included in the compose file — point `SMTP_*` at a real server or a
local mail catcher (e.g. Mailpit on port `1025`).

```sh
make init     # generate RSA keys + create .env from .env.example
make up       # docker compose up --build -d
make logs     # follow application logs
make down     # stop everything
```

The service will then accept gRPC connections on `localhost:50051`
(`APP_GRPC_PORT` in `.env`).

## Run locally (without Docker)

1. Start PostgreSQL, Memcached and an SMTP server yourself, or run just the
   dependencies from compose:

   ```sh
   docker compose up -d postgres memcached
   ```

2. Bootstrap keys and config, then make sure `.env` points at your services
   (e.g. `DB_HOST=localhost`, `CACHE_HOST=localhost`):

   ```sh
   make init
   ```

3. Build and run:

   ```sh
   make build      # produces bin/auth-service
   make run        # or: go run ./cmd/app
   ```

Migrations in `migration/` are applied automatically on startup.

## Useful Make targets

```
make help     list all targets
make init     generate RSA keypair + create .env (bootstrap)
make keys     generate RSA keypair into infra/keys
make build    build the binary into bin/auth-service
make run      run the service locally
make tidy     go mod tidy
make fmt      go fmt ./...
make vet      go vet ./...
make test     go test ./...
make up       docker compose up --build -d
make down     docker compose down
make logs     follow app logs
make clean    remove generated artifacts (bin/, keys)
```

## Notes

- There is no separate migration step — `cmd/app` runs migrations against
  PostgreSQL on every startup.
- Email confirmation, password reset and account restore links are built from
  `HTTP_SCHEME://HTTP_HOST:HTTP_PORT`; point these at the frontend that
  handles those routes.
