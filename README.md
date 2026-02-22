<a name="readme-top"></a>

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <a href="https://github.com/ikhsan892/mox">
    <img src="docs/logo.png" alt="Logo" width="150" height="150">
  </a>
  <h3 align="center">Mox</h3>
  <p align="center">
    Zero-Downtime Proxy — HAProxy wrapper with Master-Worker Unix OS architecture.
    <br />
    <em>MVP: Terminal User Interface (TUI)</em>
    <br />
    <br />
    <a href="https://github.com/ikhsan892/mox">View Demo</a>
    ·
    <a href="https://github.com/ikhsan892/mox/issues">Report Bug</a>
    ·
    <a href="https://github.com/ikhsan892/mox/issues">Request Feature</a>
  </p>
</div>

## About Mox

**Mox** is a zero-downtime reverse proxy built as a wrapper around [HAProxy](https://www.haproxy.org/), leveraging the **master-worker** process model on Unix-based operating systems. It provides seamless configuration reloads without dropping active connections.

### Key Highlights

- 🔄 **Zero-Downtime Reloads** — Hot-reload HAProxy configuration without dropping a single connection via master-worker mode.
- 🖥️ **TUI Interface (MVP)** — Manage backends, frontends, and observe real-time traffic from your terminal.
- 🔧 **HAProxy Wrapper** — Abstracts HAProxy's complex configuration into a simple, opinionated CLI/TUI experience.
- 🐧 **Unix Master-Worker** — Leverages Unix process management (fork, signals) for reliable, production-grade process supervision.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Commands](#commands)
- [Tech Stack](#tech-stack)
- [Contributing](#contributing)
- [License](#license)

---

## Quick Start

### Prerequisites

| Tool | Version | Required |
|------|---------|----------|
| [Go](https://go.dev/dl/) | 1.23+ | ✅ |
| [Make](https://www.gnu.org/software/make/) | any | ✅ |
| [Go-Migrate](https://github.com/golang-migrate/migrate) | latest | ✅ |
| [Protobuf](https://developers.google.com/protocol-buffers) | latest | ✅ |
| [NATS](https://docs.nats.io/) | latest | Optional |

### Run

```bash
# Clone
git clone https://github.com/ikhsan892/mox.git && cd mox

# Start all services (HTTP + Message Broker + Telemetry)
make run/dev

# Or start individual services
make run/http             # HTTP server only
make run/message-broker   # Message broker only
```

---

## Architecture

Mox implements **Hexagonal Architecture** (Ports & Adapters). Business logic is isolated from infrastructure — every external dependency (database, message broker, HTTP framework, observability) connects through an interface, making components fully swappable.

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

> **Pluggable by design** — swap NATS for **Kafka/RabbitMQ/Pulsar**, swap PostgreSQL for **MySQL/ClickHouse**, swap Echo for **Fiber/Chi** — all without touching business logic. See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

---

## Project Structure

```
mox/
├── adapters/          # Database adapter registry (multi-DB support)
├── cmd/               # CLI commands (Cobra) — all, http, message-broker, migration, version
├── drivers/           # Infrastructure adapters
│   ├── http/          #   └── Echo HTTP server, routes, error handler
│   ├── messaging/     #   └── NATS JetStream (swappable)
│   └── monitoring/    #   └── OpenTelemetry (traces, metrics, logs)
├── examples/          # Application entry point (main.go)
├── gorm/              # ORM model definitions
├── infrastructure/    # Low-level connections
│   ├── messaging/     #   └── NATS client
│   └── persistent/    #   └── PostgreSQL, MySQL, MeiliSearch
├── internal/          # Core kernel — App interface, bootstrap, lifecycle hooks
├── migrations/        # SQL migration files (go-migrate)
├── mocks/             # Auto-generated mocks (Mockery) — DO NOT EDIT
├── pkg/               # Reusable packages (config, datamanager, driver, hooks)
├── repositories/      # Repository implementations
├── seeders/           # Database seed scripts
├── tools/             # Utilities (logger, stack, helpers)
└── use_cases/         # Business logic — hexagonal port/adapter per domain
```

> 📖 For detailed file-by-file descriptions and improvement hints, see [CONTRIBUTING.md](CONTRIBUTING.md#folder-structure-comprehensive).

---

## Commands

| Command | Description |
|---------|-------------|
| `make run/dev` | Start all services |
| `make run/http` | HTTP server only |
| `make run/message-broker` | Message broker only |
| `make run/migration` | Run database migrations |
| `make run/seeders` | Run database seeders |
| `make run/live` | Dev with live reload |
| `make test` | Run all tests |
| `make test/cover` | Test coverage report |
| `make swagger/init` | Generate OpenAPI docs |

---

## Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.23+ |
| HTTP Framework | [Echo v4](https://echo.labstack.com/) |
| CLI Framework | [Cobra](https://github.com/spf13/cobra) |
| Message Broker | [NATS JetStream](https://nats.io/) (pluggable) |
| Database | PostgreSQL, MySQL (pluggable via adapters) |
| ORM | [GORM](https://gorm.io/) |
| Migrations | [go-migrate](https://github.com/golang-migrate/migrate) |
| Observability | [OpenTelemetry](https://opentelemetry.io/) (traces, metrics, logs) |
| Config | [Viper](https://github.com/spf13/viper) (TOML) |
| Search | [MeiliSearch](https://www.meilisearch.com/) |
| Testing | [Testify](https://github.com/stretchr/testify) + [Mockery](https://github.com/vektra/mockery) |

---

## Contributing

We welcome contributions! Please read **[CONTRIBUTING.md](CONTRIBUTING.md)** for:

- 📁 Detailed folder structure & file descriptions
- 💡 Areas open for contribution (prioritized)
- 📝 Coding style guide
- 🔀 Pull request workflow

---

## License

Distributed under the MIT License. See `LICENSE` for more information.

---

<p align="center">
  <a href="#readme-top">⬆ Back to top</a>
</p>
