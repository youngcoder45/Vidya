# SchoolOS — Multi-Tenant School ERP SaaS

Production-ready architecture + runnable scaffold for a multi-tenant school ERP:
**Flutter** apps (Android/iOS/tablet, mobile-first), **Go** modular-monolith API,
**PostgreSQL** + **Redis**, Razorpay online fees, RBAC + JWT, tenant isolation,
full notification system, and a token-driven theme system with dark/light mode.
 
> **This repo = all 8 phases of architecture docs + a compiling monorepo scaffold.**
> The scaffold implements working auth, tenant setup, students, attendance, fees
> (offline + Razorpay order/webhook), announcements, notifications, and dashboard.
> Homework/Exams/Payroll are designed & schema'd, registered as 501 stubs.

---

> **New here?** If you're a beginner (or sending this to a non-programmer), follow
> **[`SETUP.md`](SETUP.md)** — a from-scratch, click-by-click guide that gets the whole
> stack running in a browser with zero programming knowledge.

## Monorepo layout

```
├── docs/                          # Phases 1–8 + assumptions (start here)
│   ├── 00-assumptions-and-decisions.md
│   ├── 01-requirements.md         # FR/NFR, user stories, use cases
│   ├── 02-architecture.md         # diagrams (Mermaid), data flows
│   ├── 03-database.md             # ER diagram, tables, indexing, multi-tenancy
│   ├── 04-api.md                  # REST endpoints, schemas, auth flow
│   ├── 05-frontend.md             # Flutter architecture + theme system
│   ├── 06-backend.md              # Go clean architecture, modules, DI
│   ├── 07-deployment.md           # Docker, CI/CD, monitoring, DR
│   └── 08-roadmap.md              # milestones, sprint plan, timeline
├── backend/                       # Go API (Gin + GORM, modular monolith)
│   ├── cmd/server/main.go         # composition root
│   ├── internal/
│   │   ├── config/  server/  pkg/  events/  deps/  db/  seeds/
│   │   └── modules/               # auth · tenant · students · attendance
│   │                              # fees · announcements · notifications · dashboard
│   ├── migrations/001_init.sql    # canonical production schema (Phase 3)
│   ├── .env.example  Dockerfile  Makefile
├── mobile/                        # Flutter apps
│   ├── lib/core/theme/            # ⭐ token → semantic → ThemeData theme system
│   ├── lib/features/              # auth · dashboard (Riverpod + go_router)
│   └── test/theme_test.dart
├── deploy/                        # docker-compose (+ monitoring stack), prometheus
└── .github/workflows/ci.yml       # backend + mobile CI, deploy gate
```

---

## Quickstart (local)

```bash
# 1. Infrastructure
docker compose -f deploy/docker-compose.yml up -d postgres redis

# 2. API (seeds a demo school on first run)
cd backend
cp .env.example .env
make run            # or: APP_SEED_ON_START=true go run ./cmd/server

# 3. Smoke test
curl -s http://localhost:8080/health/ready
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"principal@greenwood.edu","password":"admin12345","device":{"device_id":"dev-1"}}'

# 4. Flutter app
cd mobile && flutter pub get && flutter run
```

**Demo login:** `principal@greenwood.edu` / `admin12345` (school admin)
**Platform admin:** `admin@schoolos.app` / `admin12345`

---

## Key decisions (summary — details in `docs/00`)

| Area | Decision |
|---|---|
| Frontend | **Flutter** — one codebase, pixel-consistent tablets, data-dense screens (comparison in docs/05 §1) |
| Backend | **Go modular monolith** (Gin, GORM) with DDD module boundaries + event bus seam |
| Database | PostgreSQL 16, shared schema + `school_id` everywhere + RLS (prod), canonical SQL in `migrations/` |
| Cache | Redis (rate limits, dashboard cache, future queues) with in-memory fallbacks |
| Fees | Append-only ledger, integer INR, offline + Razorpay online (webhook, idempotent state machine) |
| Auth | JWT access (15 m) + rotating refresh, OTP for parents, device tracking, data-driven RBAC |
| Multi-tenancy | JWT claims → tenant context → repository scoping (constructor-injected) → RLS defense-in-depth |
| Theme | Design tokens → semantic roles → derived `ThemeData`; zero hardcoded colors outside `tokens/` |

---

## Status 

- **Docs:** Phases 1–8 complete.
- **Backend:** compiles clean (`go build ./...`, `go vet ./...` pass — verified with Go 1.23).
- **Mobile:** theme system, router, auth + dashboard screens; `flutter analyze`/`flutter test` wired in CI (no toolchain in this environment to run locally).
- **Next modules:** homework, exams, payroll (schema + stubs ready) — see `docs/08-roadmap.md`.
