# SchoolOS — Multi-Tenant School ERP SaaS

**SchoolOS** is a production-oriented, multi-tenant **School ERP SaaS platform** built with **Flutter, Go, PostgreSQL, and Redis**.

It is designed as a mobile-first platform for managing schools, students, attendance, fees, announcements, notifications, dashboards, and role-based access — with an architecture designed to scale as additional modules are introduced.

![Architecture](https://img.shields.io/badge/architecture-modular--monolith-blue)
![Backend](https://img.shields.io/badge/backend-Go-00ADD8?logo=go\&logoColor=white)
![Frontend](https://img.shields.io/badge/frontend-Flutter-02569B?logo=flutter\&logoColor=white)
![Database](https://img.shields.io/badge/database-PostgreSQL-4169E1?logo=postgresql\&logoColor=white)
![Cache](https://img.shields.io/badge/cache-Redis-DC382D?logo=redis\&logoColor=white)
![Payments](https://img.shields.io/badge/payments-Razorpay-3395FF)

> **Repository status:** Architecture documentation for all 8 phases + a compiling, runnable monorepo scaffold.

---

## ✨ What Is Included?

The repository currently contains:

* Authentication & authorization
* Multi-tenant school setup
* Student management
* Attendance
* Fee management
* Offline fee payments
* Razorpay orders & webhooks
* Announcements
* Notifications
* Dashboard
* RBAC
* Tenant isolation
* Light/dark theme system
* Docker-based infrastructure
* CI pipeline

The following modules are already **designed and schema'd**, but currently registered as `501 Not Implemented` stubs:

* Homework
* Examinations
* Payroll

---

## 🧱 Architecture

```mermaid
flowchart TB
    A[Flutter Apps<br/>Android · iOS · Tablet]

    B[Go REST API<br/>Gin · GORM · JWT]

    C[(PostgreSQL 16<br/>Shared Schema + RLS)]
    D[(Redis<br/>Cache + Rate Limiting)]
    E[Razorpay<br/>Online Payments]

    A --> B

    B --> C
    B --> D
    B --> E
```

### Backend Architecture

SchoolOS uses a **modular monolith** rather than a microservices-first architecture.

Each domain is isolated behind clear module boundaries while sharing a single deployable backend.

```text
backend/
└── internal/
    ├── auth/
    ├── tenant/
    ├── students/
    ├── attendance/
    ├── fees/
    ├── announcements/
    ├── notifications/
    └── dashboard/
```

This keeps local development and deployment relatively simple while leaving room for future service extraction if individual modules eventually need independent scaling.

---

## 🔐 Multi-Tenancy

Tenant isolation is implemented using multiple layers of protection:

```text
JWT Claims
    │
    ▼
Tenant Context
    │
    ▼
Repository Scoping
    │
    ▼
PostgreSQL Row-Level Security
```

Every tenant-owned record is associated with a `school_id`.

Production deployments additionally use PostgreSQL **Row-Level Security (RLS)** as defense-in-depth against accidental cross-tenant access.

---

## 💳 Payments

The fee system uses an **append-only ledger** designed for traceability and idempotent payment processing.

Supported flows:

* Offline fee payments
* Razorpay order creation
* Razorpay payment verification
* Razorpay webhooks
* Idempotent webhook processing
* Explicit payment state transitions
* Integer-based INR amounts

---

## 🔑 Authentication & RBAC

Authentication and authorization currently support:

* JWT access tokens
* Rotating refresh tokens
* OTP authentication for parents
* Device tracking
* Role-based access control
* Tenant-aware authorization
* Data-driven permissions

Authentication flow:

```text
Client
  │
  ▼
Login / OTP
  │
  ▼
JWT Access Token
  │
  ├── Tenant Context
  ├── User Identity
  └── Role / Permissions
  │
  ▼
Protected API
```

---

## 🎨 Theme System

The Flutter application uses a token-driven design system instead of scattering hardcoded colors throughout the UI.

```text
Design Tokens
      │
      ▼
Semantic Roles
      │
      ▼
Flutter ThemeData
      │
      ▼
Application UI
```

The system supports:

* Light mode
* Dark mode
* Centralized design tokens
* Semantic colors
* Typography tokens
* Spacing tokens
* Derived `ThemeData`
* Theme unit tests

---

# 📁 Repository Structure

```text
.
├── docs/
│   ├── 00-assumptions-and-decisions.md
│   ├── 01-requirements.md
│   ├── 02-architecture.md
│   ├── 03-database.md
│   ├── 04-api.md
│   ├── 05-frontend.md
│   ├── 06-backend.md
│   ├── 07-deployment.md
│   └── 08-roadmap.md
│
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── server/
│   │   ├── pkg/
│   │   ├── events/
│   │   ├── deps/
│   │   ├── db/
│   │   ├── seeds/
│   │   └── modules/
│   │       ├── auth/
│   │       ├── tenant/
│   │       ├── students/
│   │       ├── attendance/
│   │       ├── fees/
│   │       ├── announcements/
│   │       ├── notifications/
│   │       └── dashboard/
│   ├── migrations/
│   │   └── 001_init.sql
│   ├── .env.example
│   ├── Dockerfile
│   └── Makefile
│
├── mobile/
│   ├── lib/
│   │   ├── core/
│   │   │   └── theme/
│   │   └── features/
│   │       ├── auth/
│   │       └── dashboard/
│   └── test/
│       └── theme_test.dart
│
├── deploy/
│   ├── docker-compose.yml
│   └── prometheus/
│
├── .github/
│   └── workflows/
│       └── ci.yml
│
└── SETUP.md
```

---

# 🚀 Getting Started

> **New to the project?**
> Start with [`SETUP.md`](SETUP.md) for a complete, beginner-friendly setup guide.

## Prerequisites

Make sure you have:

* [Docker](https://docs.docker.com/get-docker/)
* [Go 1.23+](https://go.dev/dl/)
* [Flutter](https://docs.flutter.dev/get-started/install)
* Git

---

## 1. Clone the Repository

```bash
git clone <your-repository-url>
cd SchoolOS
```

---

## 2. Start Infrastructure

```bash
docker compose -f deploy/docker-compose.yml up -d postgres redis
```

---

## 3. Configure the Backend

```bash
cd backend
cp .env.example .env
```

---

## 4. Start the API

```bash
make run
```

Or:

```bash
APP_SEED_ON_START=true go run ./cmd/server
```

The API will be available at:

```text
http://localhost:8080
```

---

## 5. Verify the API

```bash
curl -s http://localhost:8080/health/ready
```

---

## 6. Test Authentication

```bash
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "identifier": "principal@greenwood.edu",
    "password": "admin12345",
    "device": {
      "device_id": "dev-1"
    }
  }'
```

---

## 7. Run the Flutter App

```bash
cd mobile

flutter pub get
flutter run
```

---

# 🧪 Demo Accounts

> These credentials are for the seeded development environment only.
> **Do not use them in production.**

### School Administrator

```text
Email:    principal@greenwood.edu
Password: admin12345
```

### Platform Administrator

```text
Email:    admin@schoolos.app
Password: admin12345
```

---

# ⚙️ Technology Stack

| Layer            | Technology           |
| ---------------- | -------------------- |
| Mobile           | Flutter              |
| State Management | Riverpod             |
| Routing          | go_router            |
| Backend          | Go                   |
| HTTP Framework   | Gin                  |
| ORM              | GORM                 |
| Database         | PostgreSQL 16        |
| Cache            | Redis                |
| Authentication   | JWT + Refresh Tokens |
| Authorization    | RBAC                 |
| Payments         | Razorpay             |
| Infrastructure   | Docker               |
| Monitoring       | Prometheus           |
| CI/CD            | GitHub Actions       |

---

# 🧠 Key Architecture Decisions

| Area               | Decision                                     |
| ------------------ | -------------------------------------------- |
| **Frontend**       | Flutter for a single cross-platform codebase |
| **Backend**        | Go modular monolith using Gin + GORM         |
| **Architecture**   | Domain-oriented module boundaries            |
| **Database**       | PostgreSQL 16                                |
| **Multi-tenancy**  | Shared schema + `school_id` + PostgreSQL RLS |
| **Caching**        | Redis                                        |
| **Payments**       | Append-only ledger + Razorpay                |
| **Authentication** | JWT access tokens + rotating refresh tokens  |
| **Authorization**  | Data-driven RBAC                             |
| **Events**         | Internal event bus seam                      |
| **Theming**        | Token → semantic role → `ThemeData`          |
| **Deployment**     | Docker + CI/CD                               |
| **Observability**  | Prometheus                                   |

---

# 📚 Documentation

The architecture is documented across eight phases:

| Phase | Document                                                          | Description                            |
| ----- | ----------------------------------------------------------------- | -------------------------------------- |
| 0     | [`Assumptions & Decisions`](docs/00-assumptions-and-decisions.md) | Architecture decisions and assumptions |
| 1     | [`Requirements`](docs/01-requirements.md)                         | Functional/non-functional requirements |
| 2     | [`Architecture`](docs/02-architecture.md)                         | System architecture and data flows     |
| 3     | [`Database`](docs/03-database.md)                                 | Schema, ERD, indexes, multi-tenancy    |
| 4     | [`API`](docs/04-api.md)                                           | REST API, schemas, authentication      |
| 5     | [`Frontend`](docs/05-frontend.md)                                 | Flutter architecture and theme system  |
| 6     | [`Backend`](docs/06-backend.md)                                   | Go architecture, modules and DI        |
| 7     | [`Deployment`](docs/07-deployment.md)                             | Docker, CI/CD, monitoring and DR       |
| 8     | [`Roadmap`](docs/08-roadmap.md)                                   | Milestones, sprints and future work    |

---

# 📊 Current Status

### Documentation

* [x] Phase 0 — Assumptions & Decisions
* [x] Phase 1 — Requirements
* [x] Phase 2 — Architecture
* [x] Phase 3 — Database
* [x] Phase 4 — API
* [x] Phase 5 — Frontend
* [x] Phase 6 — Backend
* [x] Phase 7 — Deployment
* [x] Phase 8 — Roadmap

### Backend

* [x] Authentication
* [x] Tenant management
* [x] Students
* [x] Attendance
* [x] Fees
* [x] Razorpay integration
* [x] Announcements
* [x] Notifications
* [x] Dashboard
* [x] Database migrations
* [x] Seed data
* [x] Health checks
* [x] `go build ./...`
* [x] `go vet ./...`

### Flutter

* [x] Project structure
* [x] Theme system
* [x] Light/dark mode
* [x] Routing
* [x] Authentication UI
* [x] Dashboard
* [x] Riverpod integration
* [x] Theme tests
* [x] CI integration

### Upcoming Modules

* [ ] Homework
* [ ] Examinations
* [ ] Payroll
* [ ] Expanded parent portal
* [ ] Expanded teacher workflows
* [ ] Background job processing
* [ ] Production deployment
* [ ] Extended observability

---

# 🗺️ Roadmap

```mermaid
flowchart LR
    A[Architecture] --> B[Core ERP]
    B --> C[Homework]
    C --> D[Examinations]
    D --> E[Payroll]
    E --> F[Production]
    F --> G[Scale & Optimize]
```

See [`docs/08-roadmap.md`](docs/08-roadmap.md) for the detailed roadmap and sprint plan.

---

# 🏗️ Development Philosophy

SchoolOS intentionally follows a **modular-monolith architecture** instead of starting with microservices.

The goal is to provide:

* Clear domain boundaries
* Strong tenant isolation
* Simple local development
* Straightforward deployment
* Testable modules
* Explicit dependency injection
* Event-driven extension points
* A potential path toward service extraction when justified

> **The architecture should solve today's problems without creating tomorrow's distributed-system problems prematurely.**

---

# 🔒 Production Considerations

The repository provides a production-oriented architecture, but the current scaffold should **not be treated as a fully production-hardened deployment out of the box**.

Before production use, review and configure:

* Secrets management
* JWT signing keys
* Database credentials
* Razorpay production credentials
* CORS policy
* Rate limiting
* TLS termination
* Database backups
* Disaster recovery
* Monitoring and alerting
* Log aggregation
* Audit logging
* Data retention
* Privacy/compliance requirements
* Production Docker configuration

---

# 🤝 Contributing

Contributions, issues, architecture discussions, and improvements are welcome.

Before opening a pull request:

```bash
go build ./...
go vet ./...           
```

For Flutter:

```bash
flutter analyze
flutter test
```

Please keep new functionality aligned with the existing module boundaries and architectural decisions documented in [`docs/`](docs/).

---

# 📄 License

SchoolOS is distributed under the **SchoolOS Proprietary License**. See
[`LICENSE`](LICENSE) for the full terms. Redistribution and external
contributions require the terms described there.
