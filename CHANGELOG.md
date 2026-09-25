# Changelog

All notable changes to **SchoolOS** are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project follows [Semantic Versioning](https://semver.org/).

> **Current status:** Pre-release development (`0.x`)

---

## [Unreleased]

### Added

* Multi-tenant School ERP architecture
* Flutter mobile application scaffold
* Go modular-monolith backend
* PostgreSQL 16 database architecture
* Redis integration foundation
* JWT-based authentication
* Rotating refresh-token architecture
* Role-based access control (RBAC)
* Tenant-aware authorization
* Student management module
* Attendance module
* Fee management module
* Offline fee payment support
* Razorpay payment integration
* Razorpay webhook handling
* Idempotent payment processing
* Announcements module
* Notifications module
* Dashboard module
* Internal event bus foundation
* Token-driven Flutter theme system
* Light and dark themes
* Riverpod state management
* `go_router` navigation
* Docker development environment
* PostgreSQL migrations
* Development seed data
* Health/readiness endpoints
* Prometheus monitoring foundation
* GitHub Actions CI pipeline
* Architecture documentation covering phases 0–8
* Beginner-friendly `SETUP.md`
* `CONTRIBUTING.md`
* `CODE_OF_CONDUCT.md`
* `LICENSE`
* `TODO.md`

### Backend

* Added modular domain structure for:

  * Authentication
  * Tenants
  * Students
  * Attendance
  * Fees
  * Announcements
  * Notifications
  * Dashboard
* Added canonical PostgreSQL schema
* Added tenant-aware repository scoping
* Added production-oriented PostgreSQL RLS architecture
* Added append-only fee ledger design
* Added payment state machine architecture
* Added Redis integration with fallback support
* Added demo school seed data

### Flutter

* Added application shell
* Added authentication screens
* Added dashboard foundation
* Added centralized design tokens
* Added semantic theme roles
* Added derived `ThemeData`
* Added light/dark mode support
* Added theme unit tests

### Documentation

* Added requirements documentation
* Added architecture documentation
* Added database design documentation
* Added API documentation
* Added frontend architecture documentation
* Added backend architecture documentation
* Added deployment documentation
* Added project roadmap

### Planned

The following modules are currently designed and schema'd but remain unimplemented:

* Homework
* Examinations
* Payroll

---

## [0.1.0] — Initial Architecture Release

> Initial SchoolOS architecture and runnable scaffold.

### Added

#### Architecture

* Defined multi-tenant SaaS architecture
* Adopted modular-monolith backend architecture
* Defined Flutter mobile-first application architecture
* Established domain/module boundaries
* Added internal event bus seam
* Defined dependency-injection strategy

#### Backend

* Added Go backend using Gin
* Added GORM persistence layer
* Added PostgreSQL integration
* Added Redis integration
* Added authentication foundation
* Added tenant context
* Added RBAC foundation
* Added core ERP modules
* Added health checks
* Added database migrations
* Added development seed system

#### Payments

* Added offline fee payment support
* Added Razorpay order architecture
* Added Razorpay webhook architecture
* Added idempotent payment processing
* Added append-only financial ledger

#### Mobile

* Added Flutter application
* Added Riverpod
* Added `go_router`
* Added authentication flow foundation
* Added dashboard foundation
* Added centralized theme architecture
* Added light/dark mode

#### Infrastructure

* Added Docker Compose environment
* Added PostgreSQL container
* Added Redis container
* Added Prometheus foundation
* Added GitHub Actions CI

#### Documentation

* Added complete architecture documentation
* Added local development setup
* Added contribution guidelines
* Added code of conduct
* Added project license
* Added project TODO and roadmap

---

# Upcoming

The next planned releases focus on completing the core ERP workflows.

### `0.2.0` — Academic Modules

Planned:

* [ ] Homework
* [ ] Assignment management
* [ ] Examination management
* [ ] Marks entry
* [ ] Grade calculation
* [ ] Report cards

### `0.3.0` — Staff & Payroll

Planned:

* [ ] Teacher management
* [ ] Staff management
* [ ] Staff attendance
* [ ] Leave management
* [ ] Payroll
* [ ] Payslips

### `0.4.0` — Parent & Student Experience

Planned:

* [ ] Parent portal
* [ ] Student portal
* [ ] Push notifications
* [ ] Fee payment UI
* [ ] Homework submissions
* [ ] Examination results
* [ ] Attendance history

### `0.5.0` — Production Readiness

Planned:

* [ ] Security audit
* [ ] Tenant-isolation audit
* [ ] Expanded automated testing
* [ ] Production deployment
* [ ] Database backups
* [ ] Disaster recovery
* [ ] Production monitoring
* [ ] Error tracking
* [ ] Production Razorpay configuration

### `1.0.0` — Production Release

Target criteria:

* [ ] Core ERP modules complete
* [ ] Production security review complete
* [ ] Tenant isolation verified
* [ ] Automated test coverage established
* [ ] Production deployment validated
* [ ] Monitoring and alerting operational
* [ ] Backup and disaster recovery validated
* [ ] Production documentation complete

---

## Versioning

SchoolOS follows **Semantic Versioning**:

```text
MAJOR.MINOR.PATCH
```

For example:

```text
1.4.2
│ │ └── Patch: bug fixes and small corrections
│ └──── Minor: backwards-compatible features
└────── Major: breaking changes
```

While the project is below `1.0.0`, the API and architecture may evolve more rapidly and breaking changes may occur between minor versions.

---

## Release Categories

Changes are grouped using the following categories:

* **Added** — New functionality
* **Changed** — Changes to existing functionality
* **Deprecated** — Features that will be removed
* **Removed** — Removed functionality
* **Fixed** — Bug fixes
* **Security** — Security-related changes
* **Performance** — Performance improvements
* **Documentation** — Documentation-only changes

---

## Links

* [Repository](.)
* [Documentation](docs/)
* [Roadmap](docs/08-roadmap.md)
* [Contributing](CONTRIBUTING.md)
* [Code of Conduct](CODE_OF_CONDUCT.md)
* [License](LICENSE)

---

[unreleased]: ../../compare/v0.1.0...HEAD
[0.1.0]: ../../releases/tag/v0.1.0
