# Contributing to Mox

Thank you for your interest in contributing to **Mox**! This guide will help you understand the codebase, find areas to improve, and get started quickly.

---

## Table of Contents

- [Getting Started](#getting-started)
- [Project Architecture](#project-architecture)
- [Folder Structure (Comprehensive)](#folder-structure-comprehensive)
- [Areas Open for Contribution](#areas-open-for-contribution)
- [Migration Guide](#migration-guide)
- [Protobuf](#protobuf)
- [NATS Cheat Sheet](#nats-cheat-sheet)
- [Coding Style](#coding-style)
- [Commit Convention](#commit-convention)

---

## Getting Started

### Prerequisites

| Tool | Required | Description |
|------|----------|-------------|
| [Go 1.23+](https://go.dev/dl/) | ✅ | Main language |
| [Go-Migrate](https://github.com/golang-migrate/migrate) | ✅ | Database migration tool |
| [Make](https://www.gnu.org/software/make/) | ✅ | Build automation ([Windows guide](https://medium.com/@samsorrahman/how-to-run-a-makefile-in-windows-b4d115d7c516)) |
| [Protobuf](https://developers.google.com/protocol-buffers) | ✅ | Schema serialization for DTOs & messaging |
| [Mockery](https://github.com/vektra/mockery) | ✅ | Mock generation for unit tests |
| [NATS Server & CLI](https://docs.nats.io/) | Optional | Default message broker (replaceable) |

### Available Make Commands

| Command | Description |
|---------|-------------|
| `make run/dev` | Start all services (HTTP + message broker) |
| `make run/http` | Start HTTP REST API server only |
| `make run/message-broker` | Start message broker driver only |
| `make run/migration` | Run database migrations |
| `make run/seeders` | Run database seeders |
| `make run/live` | Start with live reload (hot-reload on file change) |
| `make test` | Run all unit tests |
| `make test/tiultemplate` | Run tests with race condition detection |
| `make test/cover` | Run tests with coverage report |
| `make swagger/init` | Generate Swagger/OpenAPI documentation |
| `make generate-migration NAME=<name>` | Create a new migration file |
| `make compile-proto PROTO_FOLDER=<folder> PROTO_FILE=<file>` | Compile protobuf definitions |

---

## Project Architecture

Mox uses **Hexagonal Architecture** (Ports & Adapters). The core principle: **business logic never depends on infrastructure details**. All external systems (databases, message brokers, HTTP frameworks, observability) connect through interfaces (ports), making every component swappable.

```
                    ┌──────────────────────────┐
                    │      Business Logic       │
                    │       (use_cases/)        │
                    └────────┬────┬─────────────┘
                             │    │
                   ┌─────────┘    └──────────┐
                   ▼                         ▼
            ┌─────────────┐          ┌──────────────┐
            │  Input Port │          │ Output Port  │
            │ (service/)  │          │(repository/) │
            └──────┬──────┘          └──────┬───────┘
                   │                        │
        ┌──────────┼──────────┐    ┌────────┼────────┐
        ▼          ▼          ▼    ▼        ▼        ▼
    ┌───────┐ ┌────────┐ ┌─────┐ ┌────┐ ┌──────┐ ┌──────┐
    │ HTTP  │ │  NATS  │ │gRPC │ │ PG │ │MySQL │ │ ...  │
    │(Echo) │ │        │ │     │ │    │ │      │ │      │
    └───────┘ └────────┘ └─────┘ └────┘ └──────┘ └──────┘
               Adapters (drivers/ & infrastructure/)
```

---

## Folder Structure (Comprehensive)

### `adapters/` — Database Adapter Registry

Manages multi-database support. Contains the registry of SQL drivers and the factory that creates database connections based on configuration.

| File | Purpose |
|------|---------|
| `register.go` | Registry of all available SQL adapters (PostgreSQL, MySQL, etc.) |
| `sql_adapters.go` | Factory that creates `*sql.DB` instances based on adapter configuration |
| `sql_adapters_test.go` | Unit tests for the SQL adapter factory |

> 💡 **Contribute**: Add a new database adapter (e.g., SQLite, CockroachDB) by registering it in `register.go` and adding connection logic.

---

### `bin/` — Local Binary Tools

Stores locally-installed binary tools like `mockery` or `go-migrate` so the project doesn't depend on global installations.

---

### `cmd/` — CLI Commands (Cobra)

Each file defines one Cobra subcommand. These are registered in `commands.go` and attached to the root command in `service.go`.

| File | Command | Description |
|------|---------|-------------|
| `commands.go` | — | Registry: wires all subcommands into `[]*cobra.Command` |
| `all.go` | `all` | Starts **all** services (telemetry + NATS + HTTP) |
| `http.go` | `http` | Starts HTTP server only (Echo) |
| `message_broker.go` | `message-broker` | Starts NATS message broker only (+ telemetry if enabled) |
| `migration.go` | `migration <up\|down> [steps]` | Runs database migrations via go-migrate |
| `version.go` | `version` | Prints application version from config |
| `seeder.go` | `seeders` | Runs database seeders (**⚠️ currently not registered in `commands.go`**) |

> 💡 **Contribute**: Register `NewSeeder` in `commands.go`. Add new CLI commands for common DevOps tasks (health check, config validation, etc.).

---

### `docs/` — Generated Documentation

Contains auto-generated Swagger/OpenAPI documentation files. **Do not edit manually** — regenerate with `make swagger/init`.

---

### `drivers/` — Infrastructure Drivers (Adapters)

Drivers are the **primary adapters** in hexagonal architecture. They initialize and manage connections to external systems.

#### `drivers/http/`

| File | Purpose |
|------|---------|
| `echo.go` | Echo HTTP server setup — middleware, CORS, error handler, server lifecycle |
| `routes.go` | Route registration entry point |
| `httperrorhandler.go` | Centralized HTTP error response handler |
| `api/` | API handlers, grouped by domain |

> 💡 **Contribute**: Swap Echo for another HTTP framework (Fiber, Chi, Gin) by implementing the same driver interface.

#### `drivers/messaging/nats/`

NATS JetStream driver — handles connection, stream/consumer creation, publish/subscribe.

> 💡 **Contribute**: Add a **Kafka**, **RabbitMQ**, or **Pulsar** driver under `drivers/messaging/<provider>/`. Implement the same interface so the `cmd/` layer can swap providers without code changes.

#### `drivers/monitoring/`

| File | Purpose |
|------|---------|
| `otel.go` | OpenTelemetry SDK initialization (traces, metrics, logs) |
| `span.go` | Helper for creating and managing trace spans |
| `trace_context_request.go` | Extract/inject trace context from/to HTTP requests |
| `logger/` | OTel log exporter |
| `metric/` | OTel metric exporter |
| `trace/` | OTel trace exporter |

> 💡 **Contribute**: Add custom metric collectors, implement Prometheus exporter as alternative to OTLP.

---

### `examples/` — Application Entry Point

Contains `main.go` — the reference entry point showing how to bootstrap and start the application with lifecycle hooks.

---

### `gorm/` — ORM Models

GORM model definitions used by repositories for database operations.

> 💡 **Contribute**: Keep models in sync with migrations. Add model hooks, soft-delete support, or custom data types.

---

### `infrastructure/` — Low-Level Infrastructure Implementations

Provides the actual connection/client implementations that adapters use.

#### `infrastructure/persistent/`

| File | Purpose |
|------|---------|
| `postgres.go` | Raw `*sql.DB` connection to PostgreSQL |
| `postgres_gorm.go` | GORM-wrapped PostgreSQL connection |
| `mysql.go` | Raw `*sql.DB` connection to MySQL |
| `meilisearch.go` | MeiliSearch client setup |

> 💡 **Contribute**: Add Redis, MongoDB, ClickHouse, or Elasticsearch connections.

#### `infrastructure/messaging/nats/`

Low-level NATS connection management and JetStream configuration.

> 💡 **Contribute**: Add `infrastructure/messaging/kafka/`, `infrastructure/messaging/rabbitmq/`, etc.

---

### `internal/` — Core Application Kernel

The heart of the application. Defines the `App` interface and `BaseApp` implementation that orchestrates configuration, logging, database, drivers, and lifecycle hooks.

| File | Purpose |
|------|---------|
| `app.go` | `App` interface — the contract every application component depends on |
| `base_app.go` | `BaseApp` struct — concrete implementation (bootstrap, config, logger, datasource, shutdown) |
| `events.go` | Lifecycle event types (`BeforeApplicationBootstrapped`, `AfterApplicationBootstrapped`, `CloseEvent`) |
| `command.go` | Command interface definition |
| `query.go` | Query bus / CQRS query handling |
| `queue.go` | Queue interface definition |

> 💡 **Contribute**:
> - Implement `Cache()` — currently `panic("unimplemented")`
> - Implement `Storage()` — filesystem abstraction, currently `panic("unimplemented")`
> - Implement `Restart()` — graceful restart, currently `panic("unimplemented")`
> - Add CQRS command bus alongside the existing query bus

---

### `migrations/` — Database Migration Files

SQL migration files in `<timestamp>_<name>.up.sql` / `.down.sql` format, used by go-migrate.

---

### `mocks/` — Auto-Generated Mocks

Mock implementations generated by Mockery. **Do not edit manually** — regenerate with `make mock`.

---

### `pkg/` — Reusable Public Packages

Standalone packages that can be imported by other projects.

| Package | Purpose |
|---------|---------|
| `config/` | Configuration loading via Viper (TOML files), struct binding |
| `datamanager/` | Multi-datasource manager — register adapters, connect, retrieve by name |
| `driver/` | Driver lifecycle manager — register, run, close external service drivers |
| `hooks/` | Generic typed event hook system (subscribe, execute) |

> 💡 **Contribute**: Add `pkg/validator/`, `pkg/pagination/`, `pkg/response/` for commonly needed utilities.

---

### `repositories/` — Repository Implementations

Concrete database query implementations of the repository interfaces defined in `use_cases/<domain>/port/output/repository/`.

---

### `seeders/` — Database Seeders

Seed scripts to populate the database with initial or test data.

---

### `tools/` — Internal Helper Utilities

| Package | Purpose |
|---------|---------|
| `logs/` | Custom `slog.Handler` with structured JSON output and filtering |
| `stack/` | Generic stack data structure |
| `utils/` | General-purpose utility functions |

---

### `use_cases/` — Business Logic (Domain Layer)

The core domain logic, organized by bounded context. Each use case follows hexagonal architecture:

```
use_cases/<domain>/
├── dto/              # Data Transfer Objects (request/response structs)
├── port/
│   ├── input/
│   │   ├── service/     # Service interface (what the business logic exposes)
│   │   └── message/
│   │       └── listener/  # Message listener interface
│   └── output/
│       ├── repository/  # Repository interface (what the business logic needs)
│       └── messaging/   # Messaging output interface
├── <domain>_service.go  # Service implementation
└── <domain>_service_test.go
```

> 💡 **Contribute**: Add new use cases by following this folder structure. Always define port interfaces before writing implementations.

---

## Areas Open for Contribution

Here's a prioritized list of improvements the project needs:

### 🔴 High Priority

| Area | Description | Where |
|------|-------------|-------|
| **Register Seeder Command** | `NewSeeder` exists but is not registered in `cmd/commands.go` | `cmd/commands.go` |
| **Implement Cache** | `App.Cache()` is defined but panics | `internal/base_app.go` |
| **Implement Storage** | `App.Storage()` filesystem API is undefined | `internal/base_app.go` |
| **Add Unit Tests** | Many packages lack test coverage | Project-wide |

### 🟡 Medium Priority

| Area | Description | Where |
|------|-------------|-------|
| **Kafka Driver** | Add Apache Kafka as alternative message broker | `drivers/messaging/kafka/` |
| **RabbitMQ Driver** | Add RabbitMQ as alternative message broker | `drivers/messaging/rabbitmq/` |
| **Redis Integration** | Add Redis for caching/session | `infrastructure/persistent/redis.go` |
| **Graceful Restart** | Implement `App.Restart()` | `internal/base_app.go` |
| **CQRS Command Bus** | Add command bus alongside existing query bus | `internal/command.go` |

### 🟢 Nice to Have

| Area | Description | Where |
|------|-------------|-------|
| **Prometheus Exporter** | Alternative to OTLP gRPC exporter | `drivers/monitoring/` |
| **Rate Limiting Middleware** | Add to HTTP driver | `drivers/http/` |
| **Health Check Endpoint** | Standardized health/readiness probes | `drivers/http/api/` |
| **Config Validation** | Validate config on startup | `pkg/config/` |
| **CLI Autocomplete** | Add shell completion for Cobra commands | `cmd/` |
| **Swagger Auth** | Add auth to Swagger UI | `drivers/http/` |

---

## Migration Guide

This project uses [go-migrate](https://github.com/golang-migrate/migrate) for database migrations.

### Create a new migration

```shell
make generate-migration NAME=add_status_to_orders
```

### Run all pending migrations

```shell
make run-migration
```

### CLI usage

```shell
# Apply all up migrations
./mox migration up

# Apply N steps
./mox migration up 3

# Rollback all
./mox migration down
```

For more details, see the [go-migrate documentation](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate).

---

## Protobuf

Generate DTO files from protobuf definitions:

```shell
make compile-proto PROTO_FOLDER=orders PROTO_FILE=order_payload
```

---

## NATS Cheat Sheet

<details>
<summary>Click to expand NATS commands</summary>

### Server

```bash
nats-server                          # Start with defaults
nats-server -c /path/to/config.conf  # Custom config
nats-server -p 4222                  # Custom port
nats-server -DV                      # Debug/verbose logging
```

### Publish & Subscribe

```bash
nats pub subject_name "message"      # Publish
nats sub subject_name                # Subscribe
nats sub 'subject.*'                 # Wildcard subscribe
```

### JetStream

```bash
nats stream create --config stream.json   # Create stream
nats publish subject "message"            # Publish to stream
nats subscribe subject                    # Subscribe to stream
nats consumer next consumer_name          # Consume next message
nats stream info stream_name              # Stream info
nats consumer info stream consumer        # Consumer info
```

</details>

---

## Coding Style

This project follows the [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md). Key points:

- Run `goimports` on save
- Run `golint` and `go vet` before committing
- Verify interface compliance at compile time: `var _ Interface = (*Impl)(nil)`
- Use `defer` for cleanup
- Handle errors exactly once
- Don't panic — return errors
- Use field tags in marshaled structs
- Prefer `strconv` over `fmt` for conversions
- Don't fire-and-forget goroutines

For the full guide, see [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md).

---

## Commit Convention

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(messaging): add Kafka driver
fix(migration): handle empty args gracefully
docs(readme): update architecture diagram
refactor(internal): extract config loader
test(adapters): add MySQL adapter tests
```

---

## How to Submit a Pull Request

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/your-feature`
3. Make your changes following the coding style above
4. Add/update tests for your changes
5. Run `make test` to ensure all tests pass
6. Commit using the conventional commit format
7. Push and open a Pull Request against `main`

Thank you for contributing! 🙌
