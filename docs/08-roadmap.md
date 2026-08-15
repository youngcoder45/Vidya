# Phase 8 — Development Roadmap

**Assumptions:** 2–3 backend engineers, 2 Flutter engineers, 1 QA, 1 designer (can be part-time); 2-week sprints.

---

## 1. Milestones Overview

| # | Milestone | Outcome | Timeline |
|---|---|---|---|
| M0 | Foundations (repo, CI, infra) | Monorepo, docker compose, CI green, staging up | Weeks 1–2 |
| M1 | **Core data + auth** | Tenants, sessions/classes, users, RBAC, login/OTP, devices | Weeks 3–5 |
| M2 | **Students + attendance** | Student CRUD, import, enroll/promote, attendance marking + analytics | Weeks 6–9 |
| M3 | **Fees v1 (offline)** | Heads/structures, ledger generation, offline payments, receipts, defaulter list | Weeks 10–13 |
| M4 | **Online payments (Razorpay)** | Orders, webhooks, ledger credit, receipts by email/in-app | Weeks 14–16 |
| M5 | **Homework + assignments** | Teacher create/attach, student submit, grading, analytics | Weeks 17–19 |
| M6 | **Exams & results** | Exam schedules, marks entry, grading systems, publish, report cards PDF | Weeks 20–23 |
| M7 | **Communication** | Announcements (targeting/schedule/read receipts), holiday calendar, notification service (push/email/in-app) | Weeks 24–26 |
| M8 | **Payroll + dashboard** | Salary structures/runs/slips, admin dashboard with cached metrics | Weeks 27–29 |
| M9 | **Parent app polish** | Child switcher, notifications UX, receipts download, fee payment UX | Weeks 30–31 |
| M10 | **Hardening & GA** | RLS, load testing, security audit, DPDP checklist, monitoring SLOs, GA release | Weeks 32–35 |

**Beta (invite 5 schools): end of week 26 · GA: week 35.**

---

## 2. Sprint Plan (2-week sprints)

| Sprint | Focus | Key deliverables |
|---|---|---|
| S1 | Foundations | Monorepo, Go scaffold, Flutter scaffold, theme tokens, docker compose, CI |
| S2 | Domain + DB | Canonical migration `001_init.sql`, models, repository interfaces, error/response envelope, health/metrics |
| S3 | Auth (backend) | JWT + refresh rotation, devices, RBAC middleware, rate limiting, audit middleware, OTP |
| S4 | Auth (app) | Login screens (staff password, parent OTP), secure token store, router guards, session refresh |
| S5 | Tenant & setup | School onboarding, sessions CRUD, classes/divisions/subjects, teacher records |
| S6 | Students | CRUD + profile, docs upload (S3), guardians, import CSV |
| S7 | Students + attendance (app) | Student screens, enrollment/promotion, attendance marking UI (bulk), history |
| S8 | Attendance analytics | Trends, class daily %, exports CSV/PDF, alert rule (<75%) |
| S9 | Fees data model | Heads, structures, ledger generation job, dues computation, concession |
| S10 | Fees offline | Offline payment capture, allocations, receipts PDF, defaulter list, collection reports |
| S11 | Fees (app) | Ledger screens (admin + parent), receipts download, defaulter screen |
| S12 | Razorpay integration | Order create, checkout, webhook (signed, idempotent), ledger credit, failure/refund paths |
| S13 | Payments UX | Parent pay flow, payment history, notification on capture, reconciliation report |
| S14 | Homework | Backend: CRUD, attachments, submissions, grading, analytics |
| S15 | Homework (app) | Teacher create/submissions, student list/submit, remarks UI, reminders |
| S16 | Assignments | Long-form tasks, rubric, resources, evaluation, resubmission |
| S17 | Exams data | Exam + schedules, grading systems/grade ranges, admit card PDF |
| S18 | Marks entry | Grid entry (teacher), validation, draft→publish subject results |
| S19 | Results | Full publish, aggregates/rank, report card PDF, analytics |
| S20 | Announcements | Targeting, scheduling, read receipts, holiday calendar CRUD |
| S21 | Notifications service | Templates, preferences, in-app feed, FCM push, email (SMTP), worker + retries |
| S22 | Notification events wiring | Hook all modules (fees, homework, exams, announcements) → worker fan-out |
| S23 | Payroll | Structures, monthly run, slips PDF, pay + notify, reports |
| S24 | Dashboard | Summary endpoints, cached metrics, admin dashboard app screens, revenue charts |
| S25 | Parent app polish | Child switcher, notification deep links, fee UX, offline caching of reads |
| S26 | Hardening I | RLS policies, load test (k6), p95 tuning, error budgets |
| S27 | Hardening II | Security audit (OWASP), DPDP checklist, backup/DR drill, monitoring dashboards |
| S28 | GA prep | Docs, runbooks, onboarding playbook, billing for plans, launch |

---

## 3. Team Loading & Risks

**Critical path:** fees (M3–M4) and exams (M6) are the longest — start fee data model early (S9) while auth/students land.

| Risk | Mitigation |
|---|---|
| Razorpay webhook edge cases (duplicate, delayed, refund) | Idempotent state machine + reconciliation job (S13) + webhook replay tooling |
| Attendance scale (bulk class × 40 students × daily) | Composite unique index, batch upsert, cached daily %, later partitioning |
| PDF generation (report cards/receipts) | Async worker + S3; template-locked in CI (golden-file tests) |
| RLS complexity vs dev velocity | RLS enabled only in prod migration; app-level scoping is primary, RLS is defense-in-depth |
| Flutter tablet UX for admin flows | Tablet-first design tokens from M1; admin screens tested on 10" iPad/Android |
| DPDP compliance drift | Consent + data-export + deletion workflows scoped in M10; legal review checkpoint |

---

## 4. Post-GA (v1.1+ backlog)

- **Web admin console** (Flutter Web) for school staff on desktop.
- Transport tracking, library module, hostel management, HR leave, timetables.
- SMS channel activation (MSG91/Twilio adapter) for fee reminders & OTPs.
- Offline-first sync (write-behind queue) for low-connectivity campuses.
- Multi-language beyond English/Hindi; report-card watermark/QR verification.
- Analytics warehouse (read replica → ClickHouse/BigQuery) + principal-level BI dashboards.
