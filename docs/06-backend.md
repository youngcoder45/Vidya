# Phase 6 — Backend Architecture (Go)

**Language:** Go 1.23 · **HTTP:** Gin · **ORM:** GORM (Postgres driver) · **Cache:** go-redis/v9 · **JWT:** golang-jwt/v5 · **Hashing:** x/crypto/bcrypt · **Validation:** go-playground/validator · **Logging:** stdlib `log/slog` · **Config:** env (`APP_*`) + optional `.env`

---

## 1. Principles

- **Clean Architecture (inward dependency rule):** `handlers → services → repositories → domain`; domain knows nothing of Gin, GORM, or Redis.
- **DDD modules:** each module is a bounded context with its own `domain` (entities), `repository` (interface), `service` (use cases), `handler` (HTTP), and `dtos`.
- **Dependency Injection:** a small composition root (`cmd/server`) wires `DB → repositories → services → handlers → router`; no global singletons, no service locator.
- **Modular monolith:** one binary, one process, one transaction boundary. Modules interact through Go interfaces + a **domain event bus** (in-process publisher; worker subscribes) so any module can later become a service.
- **Tenant safety:** every repository is created with a `TenantContext`; repositories *cannot* be constructed without a school scope in request paths. RLS added in production.

---

## 2. Folder Structure

```
backend/
├── cmd/server/main.go              # composition root, config load, DI, migrate, serve
├── internal/
│   ├── config/config.go            # env → Config (typed, validated)
│   ├── server/
│   │   ├── router.go               # Gin engine, route registration, versioning
│   │   ├── server.go               # http.Server lifecycle, graceful shutdown
│   │   └── middleware/
│   │       ├── requestid.go        # X-Request-ID (echo/inject)
│   │       ├── logger.go           # slog request logging (structured)
│   │       ├── cors.go             # allowed origins (per env)
│   │       ├── ratelimit.go        # Redis fixed-window; in-memory fallback
│   │       ├── auth.go             # JWT verify → claims in context
│   │       ├── tenant.go           # TenantContext from claims (+ header pre-auth)
│   │       ├── rbac.go             # permission check middleware
│   │       └── audit.go            # sensitive-op audit writer
│   ├── pkg/
│   │   ├── jwtutil/jwt.go          # issue/parse access+refresh
│   │   ├── passwd/passwd.go        # bcrypt hash/verify
│   │   ├── httpx/response.go       # envelope, error mapping, pagination
│   │   ├── httpx/errors.go         # AppError codes
│   │   └── tenant/tenant.go        # TenantContext type + context helpers
│   ├── events/bus.go               # domain event bus (sync dispatch + worker queue)
│   └── modules/
│       ├── auth/     domain · repository · service · handler · dtos
│       ├── tenant/   schools, sessions, classes, subjects, teachers
│       ├── students/ students, guardians, enrollments, documents
│       ├── attendance/
│       ├── homework/
│       ├── exams/
│       ├── fees/     heads, structures, ledger, orders, payments, receipts
│       ├── payroll/
│       ├── announcements/ (+ calendar events)
│       ├── notifications/
│       └── dashboard/
├── migrations/001_init.sql         # canonical schema (Phase 3)
├── seeds/seed.go                   # platform admin + demo school (dev)
├── .env.example
├── Dockerfile
├── Makefile
└── go.mod
```

### Module internal layout (example: `fees`)
```
internal/modules/fees/
├── domain/fee_ledger.go        # entities, value objects, invariants (money as int paise)
├── repository/gorm_repo.go     # GORM implementation of repository interface
├── repository/repository.go    # interface (port)
├── service/fee_service.go      # use cases: generate ledgers, offline payment, order creation, webhook
├── handler/http.go             # Gin handlers (thin): bind → validate → service → respond
└── dtos/requests.go            # request/response DTOs + validator tags
```

---

## 3. Cross-Cutting Concerns

| Concern | Implementation |
|---|---|
| AuthN | JWT access (15 m, RS256 in prod / HS256 dev) + opaque rotating refresh (hashed) |
| AuthZ | RBAC: `roles`/`permissions` from DB, cached in Redis 15 m; `RequirePermission("fees.payment.create")` middleware |
| Tenant | `tenant.MustFrom(ctx)` in every repository; scope always applied |
| Idempotency | money endpoints: unique `idempotency_key` in `fee_payment_orders`; webhook state machine |
| Audit | `audit.go` middleware + explicit service calls → `audit_logs` (async buffered writer) |
| Rate limit | Redis fixed-window per `(route-group, user/IP)`; `x/time/rate` fallback in-memory |
| Errors | `httpx.AppError{Code, Message, Details, Status}`; single error→response mapper |
| Logging | slog JSON, `request_id` + `school_id` + `user_id` fields on every line |
| Metrics | Prometheus: HTTP latency/status histograms, DB pool stats, Redis ops, queue depth |
| Validation | validator tags on DTOs; 400 with field details |
| Events | `events.Bus{Publish(ev)}` → sync side-effects + async worker via Redis stream (fallback in-process channel) |

---

## 4. Service Layer Highlights (money paths)

### Fee payment offline (accountant)
```
1. Load ledger items for student (tenant-scoped), verify outstanding.
2. Validate allocations sum == payment amount && each ≤ outstanding.
3. tx: insert payments (append-only) + payment_allocations + update fee_ledger.paid_amount/status + receipt row + audit.
4. Publish FeePaid event → notifications (receipt push/email) + dashboard cache invalidation.
```

### Razorpay webhook (`payment.captured`)
```
1. Verify signature (secret from env) → decode payload.
2. Load order by gateway_order_id; reject if already CAPTURED (idempotent).
3. tx: order CAPTURED → insert payment + allocations (map order.ledger_ids) + receipt + audit.
4. Publish FeePaid; 200 fast-ack; any failure → 500 (Razorpay retries).
```

### Salary run (monthly)
```
1. Create salary_run (month, year).
2. For each active teacher: apply structure → gross, deductions, net → salary_slips (draft).
3. Finalize (lock amounts) → Pay (mark paid, generate payslip PDF, publish SalaryPaid → notify).
```

### Result publish
```
1. Validate all subject marks published & graded.
2. Compute aggregates (per subject, class rank, overall grade) in tx.
3. Write result_publishes (versioned) + generate report card PDFs (async).
4. Publish ResultsPublished → notifications to parents/students.
```

---

## 5. Module → Service mapping (microservice-ready)

| Future service | Modules extracted | Communication |
|---|---|---|
| `api-gateway` | auth, rbac, tenant, rate-limit | stays in monolith first |
| `billing-service` | fees, receipts, Razorpay adapter | events: FeePaid, RefundIssued |
| `notifications-service` | notifications, templates, FCM/email/SMS adapters | consumes all events |
| `academics-service` | attendance, homework, exams, results | events: AttendanceMarked, ResultsPublished |
| `hr-payroll-service` | teachers, salary | events: SalaryPaid |
| `reporting-service` | dashboard analytics, exports | read-replica queries + events |

Each module already owns its schema subset + repository interface, so extraction = new process + wire event bus → message broker + move route group. No domain rewrite.

---

## 6. Dev & Prod Parity

- Local: `docker compose up` (Postgres 16 + Redis 7) + `make dev` (air hot reload) + `make migrate`.
- CI: `make lint` (golangci-lint), `make test` (unit + integration with testcontainers), `make build`.
- Prod image: multi-stage `golang:1.23-alpine` build → `distroless` runtime, non-root, `GOMAXPROCS` aware, healthchecks wired to `/health/ready`.
