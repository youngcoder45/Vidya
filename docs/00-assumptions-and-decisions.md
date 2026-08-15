# Assumptions & Key Decisions

All choices below were made from the spec + your answers to the clarifying questions. Flag anything that needs to change.

## Decisions from your answers
| Topic | Decision |
|---|---|
| Frontend | **Flutter** (comparison in `docs/05-frontend.md` §1) |
| Payments | **Razorpay** first, behind a gateway-agnostic adapter |
| Deployment | Staged AWS: VPS/Docker Compose (MVP) → ECS Fargate + RDS + ElastiCache (`docs/07-deployment.md` §1) |
| Region | **India**: INR (integer paise), CBSE/State boards, April–March academic sessions, UPI |
| Web | Deferred; mobile-first (Android/iOS/tablet) |

## Architecture assumptions
1. **Single-repo monorepo** (`backend/`, `mobile/`, `deploy/`, `docs/`) — one team, one CI.
2. **Shared-schema multi-tenancy** with `school_id` on all tenant tables + app-level scoping + Postgres RLS in prod. Per-school databases rejected: too costly to operate at SaaS scale, complicates cross-school features (none needed in v1).
3. **Gin over Fiber** — larger middleware ecosystem, stable API, easier hiring; Fiber's speed advantage is negligible at this I/O-bound workload.
4. **GORM for the scaffold** (velocity, matches the schema), with repository interfaces so SQLC can replace it later without touching services. Canonical schema lives in SQL migrations.
5. **Money as integer paise** — no float; amounts serialized as integers.
6. **Fee ledger is append-only**; corrections via reversing entries + audit. This is the single most important integrity decision for a school ERP.
7. **In-app + push + email now, SMS later** — SMS adapter interface is defined; MSG91/Twilio plugged in v1.1.
8. **Notifications use at-least-once delivery** with idempotency keys; parent gets duplicate-safe receipts.
9. **Academic sessions are first-class** — all class membership, fees, attendance, and exams are session-anchored; session activation gates writes.
10. **Scaffold scope:** backend implements working auth, tenant/setup, students, attendance, fees (offline + Razorpay order/webhook), announcements, notifications, dashboard, health — with homework/exams/payroll modeled in SQL and stubbed as modules to implement next. Flutter implements the theme system, router, auth + dashboard screens end-to-end.

## Requirements clarifications (spec gaps we filled)
- **Divisions** = sections ("A", "B") under a class, per academic session (Indian convention).
- **Homework vs assignments** = short daily tasks vs long-form graded tasks; separate modules sharing a shape.
- **Attendance model** = daily row per student per class-division (subject-attendance optional in v1).
- **Roles** = platform_admin, school_admin, accountant, teacher, parent, student (expandable).
- **Parents** = one account, many children; OTP login; child switcher in app.
- **Receipts/report cards/payslips** = generated PDFs on S3/CDN, never re-rendered ad hoc.
- **Fee frequency** = monthly/quarterly/term/yearly heads; ledgers auto-generated per session with due dates; partial + installment supported.
- **GST** = fee receipts carry invoice fields (schools usually exempt, but format is compliant).

## Open questions (non-blocking, answer when convenient)
1. Email provider preference for transactional mail (e.g., AWS SES, Resend, Postmark)? Default: **AWS SES** (cheapest in-region).
2. FCM app project ownership — one shared project across schools or per-school? Default: one project, tenant-tagged topics.
3. Report-card template ownership — do schools need custom layout editing, or is one branded template enough for v1? Default: one branded template.
4. Do you want the exam module to support **grading on raw marks + grade points** (CBSE-style) only, or also percentage-only schools?
