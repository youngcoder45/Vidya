# Phase 7 — Deployment Architecture

---

## 1. Recommended Hosting Path (decision for "not sure — recommend one")

**Recommendation: staged AWS adoption, starting cost-effective.**

| Stage | Setup | When | Est. monthly infra |
|---|---|---|---|
| **Stage 0 (dev/MVP)** | Single VPS (₹1.5k–3k/mo, e.g., Hetzner/AWS Lightsail) running `docker compose`: api + postgres + redis + worker | 0–25 schools, proving product-market fit | ~$20–40 |
| **Stage 1 (production)** | **AWS `ap-south-1`** (Mumbai — latency & DPDP data-residency for Indian schools): ECS Fargate (api + worker), RDS Postgres (Multi-AZ optional), ElastiCache Redis, S3 (docs/receipts), ALB + ACM TLS, Route53 | 25–200 schools | ~$300–600 |
| **Stage 2 (scale)** | RDS read replicas (analytics), ElastiCache cluster mode, S3 lifecycle tiers, CloudFront CDN, sharding path | 200+ schools | scales with tenants |

**Why AWS over GCP/Azure for this product:** data residency in India (`ap-south-1`) aligns with DPDP Act 2023; RDS/ElastiCache/Fargate are the least-ops managed path to the exact stack (Postgres, Redis, containers); ecosystem (Secrets Manager, KMS, CloudWatch) covers the security requirements out of the box. GCP is an equally valid alternative if the team is already GCP-proficient — the Docker artifacts are cloud-agnostic.

**Security baseline on AWS:** private subnets for DB/Redis; security groups allow only ALB→API, API→DB/Redis; RDS encryption at rest + TLS; Secrets Manager for `APP_ENC_KEY`, Razorpay/Razorpay-webhook secrets, JWT signing keys; KMS CMK for RDS/S3; CloudTrail + GuardDuty enabled.

---

## 2. Container Topology (Stage 0 = the committed compose file)

```
deploy/docker-compose.yml
├── postgres:16-alpine      (volume pgdata; healthcheck pg_isready)
├── redis:7-alpine          (volume redisdata; healthcheck redis-cli ping)
├── api                     (backend image; depends_on healthy pg+redis; migrate-then-serve)
└── worker                  (same image, different command: consume notification events)
deploy/docker-compose.monitoring.yml  (optional profile)
├── prometheus              (scrapes /metrics)
└── grafana                 (dashboards: API latency, error rate, queue depth, DB)
```

**Networking:** one bridge network `schoolos`; api exposed on `:8080`; postgres/redis NOT published to host.

---

## 3. CI/CD Pipeline (GitHub Actions — `.github/workflows/ci.yml`)

```
┌─────────────┐   ┌──────────────┐   ┌──────────────┐   ┌───────────────┐
│  PR / push  │ → │   lint+test  │ → │  build image │ → │  deploy       │
│  main       │   │ go vet/lint, │   │ docker build  │   │  ssh compose  │
│             │   │ go test,     │   │ tag git-sha   │   │  up -d (stage0)│
│             │   │ flutter analyze│ │ push to ECR   │   │  or ecs deploy │
└─────────────┘   └──────────────┘   └──────────────┘   └───────────────┘
```

- **PR:** `make lint`, `make test` (unit), `flutter analyze` + `flutter test`.
- **main:** + integration tests (testcontainers Postgres/Redis), build & push image (ECR or GHCR).
- **release tag (`v*`):** deploy to staging → smoke tests (`/health/ready`, login, fee order) → deploy production (blue/green on ECS, `docker compose up` + health-gated swap on VPS).
- **Migrations:** `migrate up` runs as a separate one-shot job before app rollout (never inside app start in prod); rollback = `migrate down 1` with a documented manual runbook.

---

## 4. Monitoring & Alerting

| Layer | Tool | Key signals |
|---|---|---|
| Metrics | Prometheus (+ Grafana) | HTTP p95/p99 by route, 5xx rate, error budget, DB pool wait, Redis latency/evictions, queue depth & lag, worker processing rate |
| Logs | `log/slog` JSON → stdout → CloudWatch Logs / Loki + Grafana | request_id, school_id, user_id, latency, status; audit trail separate stream |
| Errors | Sentry (optional) | stack traces for 5xx + panic recovery, grouped by signature |
| Uptime | CloudWatch Synthetics / UptimeRobot | `/health` external checks, TLS expiry |
| Alerts | Grafana Alerting / CloudWatch Alarm | p95 > 500 ms (10 m), 5xx > 1% (5 m), queue lag > 15 m, disk > 80%, DB connections > 80% |

**SLO (v1):** availability 99.5%, API p95 < 500 ms, notification delivery ≤ 5 min from event.

---

## 5. Logging & Traceability

- Structured JSON everywhere; correlation via `X-Request-ID` propagated to DB (via `app.current_request_id` setting for RLS/audit correlation) and to outbound calls.
- **Audit trail** (compliance): append-only `audit_logs` rows for logins, fee/payment ops, result publish, payroll, student data edits, permission changes — with before/after JSON and device/IP.
- Retention: app logs 30 d (hot) → 12 mo (cold); audit logs ≥ 3 years (legal requirement).

---

## 6. Backup & Disaster Recovery

| Asset | Strategy | RPO/RTO |
|---|---|---|
| Postgres | RDS automated snapshots (daily) + WAL streaming (PITR); VPS: pgBackRest S3, nightly + WAL archiving | RPO 5 min / RTO < 1 h |
| Redis | Cache only (rebuildable) — no backup needed; sessions survive via DB | — |
| S3 objects | Versioning + cross-region copy (prod) | RPO 15 min |
| Secrets | Secrets Manager / KMS (never in repo) | — |

**DR runbook:** restore latest snapshot → apply WAL to point-in-time → point ALB at restored stack → verify `/health/ready` + smoke test fee order. Documented in `docs/runbooks/`.

---

## 7. Environments

| Env | Purpose | Config source |
|---|---|---|
| `local` | docker compose, AutoMigrate for scaffold modules, seeded demo school | `.env` |
| `staging` | full CI deploy, real Razorpay test keys, FCM test app | Secrets Manager (AWS) / env file |
| `production` | ECS or compose, real keys, RLS enabled, read replica for analytics | Secrets Manager |

**Feature flags:** `schools.plans.feature_flags` (jsonb) — e.g., `sms`, `transport`, `report_cards_pdf` — read at request time (cached), so plan gating needs no deploy.
