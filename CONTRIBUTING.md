# Contributing to Vidya

Thank you for your interest in contributing to **Vidya**.

Vidya is a multi-tenant School ERP SaaS built with **Flutter, Go, PostgreSQL, and Redis**. Contributions are welcome, provided they follow the project's architecture, coding standards, and contribution guidelines.

> Please read this document before opening an issue or pull request.

---

## Table of Contents

* [Code of Conduct](#code-of-conduct)
* [Before You Start](#before-you-start)
* [Development Setup](#development-setup)
* [Project Structure](#project-structure)
* [Finding Something to Work On](#finding-something-to-work-on)
* [Branching Strategy](#branching-strategy)
* [Making Changes](#making-changes)
* [Commit Messages](#commit-messages)
* [Testing](#testing)
* [Pull Requests](#pull-requests)
* [Code Review](#code-review)
* [Architecture Guidelines](#architecture-guidelines)
* [Database Changes](#database-changes)
* [Security](#security)
* [Reporting Bugs](#reporting-bugs)
* [Feature Requests](#feature-requests)
* [Documentation](#documentation)
* [License](#license)

---

# Code of Conduct

Please keep all project interactions professional and respectful.

Contributors are expected to:

* Be respectful to other contributors.
* Provide constructive feedback.
* Focus discussions on the technical problem.
* Avoid personal attacks or harassment.
* Respect different levels of experience.
* Keep discussions relevant to the project.

Disagreements about implementation are normal. Resolve them through technical discussion and documented reasoning.

---

# Before You Start

Before making a contribution:

1. Read the [`README.md`](README.md).
2. Read the relevant documentation under [`docs/`](docs/).
3. Check existing issues and pull requests.
4. Search for existing implementations before adding new ones.
5. For larger changes, open an issue first to discuss the proposed approach.

For architectural changes, review:

```text
docs/
├── 00-assumptions-and-decisions.md
├── 02-architecture.md
├── 03-database.md
├── 04-api.md
├── 05-frontend.md
├── 06-backend.md
├── 07-deployment.md
└── 08-roadmap.md
```

---

# Development Setup

## Prerequisites

Install:

* Git
* Docker
* Go 1.23+
* Flutter
* PostgreSQL client tools (optional)
* Redis CLI (optional)

Clone the repository:

```bash
git clone <repository-url>
cd Vidya
```

Start the infrastructure:

```bash
docker compose -f deploy/docker-compose.yml up -d postgres redis
```

Set up the backend:

```bash
cd backend
cp .env.example .env
```

Run the API:

```bash
make run
```

Or:

```bash
APP_SEED_ON_START=true go run ./cmd/server
```

Set up Flutter:

```bash
cd mobile
flutter pub get
flutter run
```

For the complete setup process, see [`SETUP.md`](SETUP.md).

---

# Project Structure

```text
Vidya/
├── docs/                  # Architecture & engineering documentation
├── backend/               # Go API
│   ├── cmd/
│   ├── internal/
│   ├── migrations/
│   └── seeds/
├── mobile/                # Flutter application
├── deploy/                # Docker & infrastructure
├── .github/
│   └── workflows/         # CI/CD
├── README.md
├── SETUP.md
├── CONTRIBUTING.md
└── LICENSE
```

---

# Finding Something to Work On

Good starting points include:

* Issues labeled `good first issue`
* Issues labeled `help wanted`
* Documentation improvements
* Test coverage
* Bug fixes
* UI improvements
* Accessibility improvements
* Developer tooling
* Performance improvements

Before starting substantial work, comment on the relevant issue so multiple contributors do not work on the same problem independently.

---

# Branching Strategy

Create a branch from the default branch before making changes.

### Feature

```text
feature/<short-description>
```

Example:

```text
feature/homework-module
```

### Bug Fix

```text
fix/<short-description>
```

Example:

```text
fix/attendance-date-validation
```

### Documentation

```text
docs/<short-description>
```

Example:

```text
docs/api-authentication
```

### Refactoring

```text
refactor/<short-description>
```

Example:

```text
refactor/fees-module
```

### Chores

```text
chore/<short-description>
```

Example:

```text
chore/update-go-dependencies
```

Avoid working directly on the default branch.

---

# Making Changes

Keep changes focused.

A pull request should generally address **one problem or feature**.

Avoid combining unrelated changes such as:

```text
Feature + dependency upgrade + formatting entire project
```

Prefer:

```text
Feature implementation
```

and separate unrelated changes into their own pull requests.

---

# Backend Guidelines

Vidya uses a **Go modular-monolith architecture**.

New backend functionality should belong to an appropriate domain module.

For example:

```text
internal/modules/
├── auth/
├── tenant/
├── students/
├── attendance/
├── fees/
├── announcements/
├── notifications/
└── dashboard/
```

When adding a new domain:

```text
internal/modules/<domain>/
├── handler/
├── service/
├── repository/
├── model/
└── routes/
```

Follow the existing patterns before introducing a new architectural abstraction.

### Backend principles

* Keep domain boundaries explicit.
* Avoid unnecessary global state.
* Prefer dependency injection.
* Keep handlers thin.
* Keep business logic in services/domain logic.
* Keep database access inside repositories.
* Validate external input.
* Return appropriate HTTP status codes.
* Avoid leaking internal errors to API clients.
* Consider tenant isolation for every tenant-owned resource.

---

# Flutter Guidelines

The Flutter application should follow the existing feature-oriented structure.

```text
mobile/lib/
├── core/
│   ├── theme/
│   ├── routing/
│   └── ...
└── features/
    ├── auth/
    ├── dashboard/
    └── ...
```

Prefer:

* Riverpod for state management
* `go_router` for navigation
* Feature-based organization
* Reusable widgets
* Immutable state where appropriate
* Testable business logic

Avoid placing business logic directly inside large widget trees.

---

# Theme System

Vidya uses a token-driven theme system.

When adding UI:

**Do not introduce arbitrary hardcoded colors.**

Avoid:

```dart
Color(0xFF123456)
```

Prefer the project's existing semantic theme tokens.

The intended flow is:

```text
Design Token
     ↓
Semantic Role
     ↓
ThemeData
     ↓
Widget
```

New visual tokens should be added to the centralized theme system rather than scattered across individual widgets.

---

# Database Changes

Database changes require additional care because Vidya is multi-tenant.

Every schema change should consider:

* Tenant isolation
* Foreign keys
* Indexes
* Constraints
* Migration safety
* Existing production data
* Rollback implications

Tenant-owned tables should generally contain:

```sql
school_id
```

Do not bypass tenant scoping.

Migration files should be:

* Explicit
* Ordered
* Reproducible
* Safe to run in a clean environment

Avoid manually modifying the production database as part of a normal feature implementation.

---

# API Changes

When modifying or adding an API endpoint:

1. Update the implementation.
2. Update API documentation.
3. Define request/response schemas.
4. Add authentication/authorization requirements.
5. Consider tenant isolation.
6. Add appropriate tests.
7. Document breaking changes.

Avoid silently changing existing API behavior.

For breaking changes, clearly document:

```text
Before
  ↓
Existing API behavior

After
  ↓
New API behavior
```

---

# Testing

All contributions should include appropriate tests.

## Backend

Run:

```bash
go test ./...
```

Static checks:

```bash
go vet ./...
```

Build:

```bash
go build ./...
```

## Flutter

Run:

```bash
flutter analyze
flutter test
```

For UI changes, add widget tests where practical.

For business logic, prefer unit tests.

---

# Pull Requests

Before opening a pull request:

```bash
go build ./...
go vet ./...
go test ./...
```

And for Flutter changes:

```bash
flutter analyze
flutter test
```

A pull request should include:

* A clear title
* A concise description
* The problem being solved
* The implementation approach
* Testing performed
* Screenshots for relevant UI changes
* Migration information for database changes
* Any relevant breaking changes

### Example

```markdown
## What does this PR do?

Adds attendance filtering by class and date.

## Why?

Teachers currently need to manually search through all attendance records.

## Changes

- Added attendance filtering API
- Added repository query
- Added Flutter filter UI
- Added backend tests
- Added widget tests

## Testing

- `go test ./...`
- `go vet ./...`
- `flutter analyze`
- `flutter test`
```

---

# Pull Request Checklist

Before submitting:

* [ ] The change solves the intended problem.
* [ ] Existing functionality has not been unnecessarily changed.
* [ ] Code follows existing project conventions.
* [ ] Tests have been added or updated.
* [ ] Backend tests pass.
* [ ] Flutter analysis passes.
* [ ] Flutter tests pass.
* [ ] Documentation has been updated where necessary.
* [ ] Database migrations are included if required.
* [ ] Tenant isolation has been considered.
* [ ] No secrets or credentials have been committed.
* [ ] No unnecessary dependencies have been added.
* [ ] The PR has a clear description.
* [ ] Screenshots are included for relevant UI changes.

---

# Commit Messages

Use concise, descriptive commit messages.

Recommended format:

```text
<type>: <description>
```

Examples:

```text
feat: add homework module
fix: prevent cross-tenant attendance access
docs: update API authentication guide
refactor: simplify fee repository
test: add attendance service tests
chore: update Flutter dependencies
```

Common types:

| Type       | Purpose                 |
| ---------- | ----------------------- |
| `feat`     | New functionality       |
| `fix`      | Bug fix                 |
| `docs`     | Documentation           |
| `refactor` | Code restructuring      |
| `test`     | Tests                   |
| `chore`    | Maintenance             |
| `perf`     | Performance improvement |
| `ci`       | CI/CD changes           |

Avoid commits such as:

```text
update
changes
final
fixed stuff
asdf
```

---

# Reporting Bugs

When reporting a bug, include enough information to reproduce it.

Please provide:

### Environment

```text
OS:
Go:
Flutter:
Database:
Browser/Device:
```

### Description

Explain what happened.

### Expected Behavior

Explain what should have happened.

### Actual Behavior

Explain what actually happened.

### Reproduction Steps

```text
1. Start the application
2. Log in as ...
3. Navigate to ...
4. Perform ...
5. Observe ...
```

### Logs

Include relevant logs or stack traces.

Remove:

* Passwords
* API keys
* JWTs
* Database credentials
* Personal information
* Other secrets

---

# Security Issues

**Do not report security vulnerabilities through public GitHub issues.**

If you discover a vulnerability involving:

* Authentication
* Authorization
* Tenant isolation
* Payment processing
* Sensitive data exposure
* SQL injection
* Remote code execution
* Credential exposure
* Infrastructure security

please contact the project maintainer privately.

Do not publicly disclose an exploitable vulnerability before it has been investigated and addressed.

---

# Feature Requests

Feature requests are welcome.

A good feature request should explain:

1. What problem does it solve?
2. Who benefits from it?
3. What behavior do you expect?
4. Are there alternative approaches?
5. Does it affect the existing architecture?

Example:

```markdown
## Feature

Add parent homework notifications.

## Problem

Parents currently have no centralized way to know when new homework is assigned.

## Proposed Solution

Send a notification when homework is published for a student's class.

## Additional Considerations

- Notification preferences
- Tenant isolation
- Push notification provider
- Database event handling
```

---

# Documentation Contributions

Documentation is treated as a first-class contribution.

You can improve:

* Architecture documentation
* API documentation
* Setup instructions
* Deployment documentation
* Code comments
* Examples
* Troubleshooting guides
* Developer experience

If implementation and documentation disagree, update the documentation or raise an issue explaining the discrepancy.

---

# Dependency Changes

Avoid adding dependencies unless they provide a clear benefit.

Before adding a dependency, consider:

* Is the functionality already available?
* Is the dependency actively maintained?
* Is the license compatible?
* Does it significantly increase build size?
* Does it introduce security or operational concerns?
* Is the dependency necessary?

For significant dependencies, explain the reasoning in the pull request.

---

# Architecture Changes

Large architectural changes should be discussed **before implementation**.

Examples include:

* Changing the database architecture
* Introducing microservices
* Changing authentication
* Changing tenant isolation
* Replacing Riverpod
* Replacing Gin/GORM
* Introducing a new message broker
* Changing the payment architecture
* Changing deployment architecture

For architectural proposals, include:

```text
Problem
    ↓
Current Architecture
    ↓
Proposed Architecture
    ↓
Trade-offs
    ↓
Migration Strategy
```

The existing architecture documentation should be treated as the starting point for these discussions.

---

# Contributor Recognition

Contributors whose changes are accepted may be recognized in the project's contributor documentation or GitHub contributor graph.

Significant contributions may also be acknowledged in release notes.

---

# License

By contributing to Vidya, you agree that your contributions may be distributed under the project's applicable license.

See [`LICENSE`](LICENSE) for the complete licensing terms.

---

# Questions?

If something is unclear:

1. Check [`README.md`](README.md).
2. Check [`SETUP.md`](SETUP.md).
3. Search existing GitHub issues.
4. Check the relevant documentation under [`docs/`](docs/).
5. Open a discussion or issue if the answer cannot be found.

Thank you for contributing to Vidya.
