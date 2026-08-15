# Phase 1 — Requirements, User Stories & Use Cases

**Project:** Multi-tenant School ERP SaaS ("SchoolOS")
**Market:** India (INR, CBSE/State boards, April–March academic sessions)
**Platforms:** Android, iOS (mobile-first), tablet; web deferred to a later milestone
**Status:** v1.0 target scope

---

## 1. Context & Goals

SchoolOS is a SaaS platform that lets **multiple schools** manage operations through one
product: student management, attendance, homework/assignments, examinations & results,
fees & payments, teacher payroll, announcements, and parent/student engagement.

**Product goals**
1. Replace spreadsheets + paper registers + fragmented apps with one mobile-first ERP.
2. Reduce fee-default and collection friction (due reminders, online payments, receipts).
3. Give school leadership real-time visibility (attendance, revenue, results).
4. Let teachers do daily work (attendance, homework, marks) from a phone.
5. Keep parents informed (attendance, homework, results, fees, notices) without staff effort.

**Business goals (SaaS)**
- One codebase, many tenants; zero per-school custom deployments.
- Tenant isolation guarantees; per-school branding, sessions, and fee structures.
- Low marginal cost per school; revenue from per-school/monthly subscription tiers.

---

## 2. Actors

| Actor | Description |
|---|---|
| **Platform Super Admin** | Manages tenant (school) onboarding, platform config, plans, global support. Not a school user. |
| **School Admin** | Runs the school: students, staff, fees, exams, payroll, announcements, reports. |
| **Accountant / Bursar** | Fee structures, fee collection, receipts, defaulter lists, financial reports (subset of admin rights). |
| **Teacher** | Attendance, homework, assignments, marks, student progress, parent communication (scoped to own classes/subjects). |
| **Parent** | Linked to 1+ children; views progress, fees, notices; pays fees; receives notifications. |
| **Student** | Views own homework, results, timetable, notices (read-only, same portal as parent). |
| **System (cron/worker)** | Sends scheduled notifications, fee-due reminders, payroll runs, exam reminders. |

---

## 3. Functional Requirements

### FR-1 Platform & Tenant Management (Super Admin)
- FR-1.1 Onboard a new school (name, address, board, academic session config, branding: logo, primary color, domain/subdomain).
- FR-1.2 Suspend/reactivate a school; view per-tenant usage (users, storage, API volume).
- FR-1.3 Configure subscription plans and feature flags per tenant.
- FR-1.4 Platform audit log access.

### FR-2 School Setup (School Admin)
- FR-2.1 Configure academic sessions (April–March default, customizable), class structure (classes → divisions/sections), subjects per class, working days.
- FR-2.2 Manage staff records: teachers (and non-teaching staff), subjects taught, class-teacher assignment.
- FR-2.3 Configure branding for parent/student apps (logo, color, notices letterhead).

### FR-3 Student Management
- FR-3.1 CRUD students with profile: name, DOB, gender, admission no, enrollment date, class/division, roll no, guardian(s), contact, address, photo, blood group, medical notes.
- FR-3.2 Bulk import students (CSV/Excel) with validation and error report.
- FR-3.3 Enroll/promote/drop students across academic sessions (session-wise class assignment).
- FR-3.4 Upload documents (birth certificate, Aadhaar, transfer certificate, photos) with type & expiry.
- FR-3.5 Track academic history per session: class, division, roll, attendance %, aggregate, promotions.
- FR-3.6 Search/filter (class, division, name, admission no, status); view complete profile.

### FR-4 Attendance
- FR-4.1 Teacher marks daily attendance per class/division for own subjects/classes (present/absent/late/half-day/on-leave).
- FR-4.2 Bulk marking (mark all present then adjust exceptions), date range edit with permission & audit.
- FR-4.3 Attendance analytics: daily % per class, per-student trend, month-wise summary; export CSV/PDF.
- FR-4.4 Admin monitors attendance across all classes; alert rules (e.g., <75% triggers parent notification).
- FR-4.5 Lock/approve attendance per day after a cutoff (configurable) to prevent retro-edits.

### FR-5 Homework
- FR-5.1 Teacher creates homework: title, description, class/division, subject, due date, attachments, optional marks.
- FR-5.2 Students/parents view pending & past homework with due dates.
- FR-5.3 Submission workflow: student submits text/attachment; teacher reviews, remarks, grades, returns.
- FR-5.4 Admin monitors submission rates per class/subject; pending-work lists.
- FR-5.5 Automatic reminders before deadlines (T-24h), overdue alerts to students/parents.

### FR-6 Assignments & Resources
- FR-6.1 Teacher creates assignments (longer tasks): rubric, due date, resources (files/links), evaluation scheme.
- FR-6.2 Student submits; teacher evaluates with marks, feedback, resubmission toggle.
- FR-6.3 Upload shared class resources (notes, worksheets) in a per-class library.

### FR-7 Examinations & Results
- FR-7.1 Create exams: name (e.g., "Term 1"), type (unit test/term/final), class scope, schedule (date, time, subject-wise), max marks per subject.
- FR-7.2 Configure grading systems per school (marks → grade + grade point, e.g., CBSE A1–E2) and pass marks.
- FR-7.3 Teachers enter marks per subject (per student), validate (≤ max marks, numeric), publish subject-wise.
- FR-7.4 Admin publishes exam results (aggregate per subject, class rank, overall grade), generates report cards (PDF, branded).
- FR-7.5 Generate admit cards (student photo, exam schedule, instructions) as PDF.
- FR-7.6 Academic analytics: subject/class performance, pass %, grade distribution, year-over-year.

### FR-8 Fees & Payments
- FR-8.1 Fee structure setup: fee heads (tuition, transport, hostel, lab, exam, misc), amounts, frequency (monthly/quarterly/term/yearly), class/division applicability, due dates.
- FR-8.2 Student-wise fee ledger auto-generated from structure per session; supports concessions/scholarships and adjustments.
- FR-8.3 Partial payments and installments (schedule: amount + due date per installment).
- FR-8.4 Payment capture: cash/UPI/card/cheque recorded by office; **online payments via Razorpay** (order → gateway → webhook → ledger credit).
- FR-8.5 Receipt generation (branded, receipt no, ledger breakdown, payment mode) — PDF + in-app + email; receipts downloadable by parents.
- FR-8.6 Due tracking: auto-computed outstanding; late-fee rules (configurable); **defaulter list** generation (by class, date range, amount).
- FR-8.7 Payment history per student; refunds/adjustments with audit.
- FR-8.8 Financial reports: fee collection summary, head-wise, class-wise, month-wise, outstanding ageing, revenue dashboard.

### FR-9 Teacher Salary & Payroll
- FR-9.1 Teacher profile + bank details (encrypted).
- FR-9.2 Salary structure: components (basic, HRA, allowances, deductions), monthly net; per-teacher or by role template.
- FR-9.3 Monthly salary run: generate salary slips, mark paid (bank transfer/cheque/cash), track history, arrears & deductions.
- FR-9.4 Payslip PDF per teacher; payroll reports (monthly total, head-wise, YTD).

### FR-10 Announcements & Notices
- FR-10.1 Create announcements with audience targeting: whole school, class, division, individual students, staff group.
- FR-10.2 Schedule announcements (publish at a future date/time); draft → publish → archive.
- FR-10.3 Delivery: in-app feed + push + email; read-receipt tracking (seen by X of Y).
- FR-10.4 Holiday calendar: holidays, events, activities with dates, descriptions, audience; school calendar view.

### FR-11 Notifications
- FR-11.1 Central notification service; channels: in-app, push (FCM), email; SMS adapter interface (future).
- FR-11.2 Event types: homework assigned, assignment deadline, exam schedule, results published, fee due, announcement, holiday, attendance alerts, salary paid.
- FR-11.3 Per-user preference: channel opt-in per event category; quiet hours.
- FR-11.4 Templates (per event + channel) with variables; tenant-level branding.

### FR-12 Parent/Student Portal
- FR-12.1 Link parent ↔ 1..n children (invite by phone/OTP; one account, child switcher).
- FR-12.2 Dashboard: today's attendance, pending homework, upcoming exams, due fees, recent notices, results.
- FR-12.3 Fee tracking: pending, history, receipts download, pay online (Razorpay).
- FR-12.4 Progress: subject-wise performance, attendance trends, report cards, teacher remarks.

### FR-13 Dashboard & Analytics (Admin)
- Student count (active, per class), teacher count, today's attendance %, pending fees (amount + count), upcoming exams, recent announcements, revenue overview (collected vs due, month trend), defaulter snapshot.

### FR-14 Authentication & Security
- FR-14.1 Login (email/phone + password) for staff; OTP login for parents (phone), optional email.
- FR-14.2 JWT access + refresh tokens; device tracking (device id, name, platform, last seen, revoke device).
- FR-14.3 RBAC: roles → permissions; admin grants roles (e.g., accountant, exam cell); permission checks server-side on every request.
- FR-14.4 Password policy & hashing (bcrypt), forgot-password flow (email/OTP), rate limiting on auth endpoints.
- FR-14.5 Audit logs: who did what, when, from which device/IP, for all sensitive operations (fees, results, payroll, student data edits, permissions).

---

## 4. Non-Functional Requirements

| ID | Category | Requirement |
|---|---|---|
| NFR-1 | Performance | API p95 < 300 ms for read endpoints (non-report); dashboard < 2 s; attendance marking < 1 s/class-day save. |
| NFR-2 | Scale | 100+ schools, 100k students, 1M attendance rows/month on initial architecture; horizontal scale path documented. |
| NFR-3 | Availability | 99.5% monthly uptime target; graceful degradation when Redis/email downstream is down; no data loss (durable Postgres). |
| NFR-4 | Security | OWASP Top 10; bcrypt hashes; JWT with short-lived access tokens + rotating refresh; TLS everywhere; encrypted PII at rest (AES-256 via column-level or app-level encryption); tenant isolation enforced at query layer **and** Postgres RLS (production hardening); audit for sensitive ops. |
| NFR-5 | Privacy | PII minimization; role-scoped data access; parent sees only own children; staff see only assigned classes/subjects; data export per tenant; deletion workflows (GDPR-style, DPDP Act 2023 alignment). |
| NFR-6 | Multi-tenancy | Shared-schema, tenant-scoped rows; every table carries `school_id`; repository layer always scopes; RLS as defense-in-depth. |
| NFR-7 | Usability | Mobile-first; offline-tolerant reads (caching); minimum target Android 7+/iOS 13+; screen reader support; dark & light themes; min touch target 44 px. |
| NFR-8 | Maintainability | Modular monolith with module boundaries; generated docs; versioned API; coverage ≥ 70% for core money/result paths; lint in CI. |
| NFR-9 | Observability | Structured JSON logs with trace/request IDs; Prometheus metrics (HTTP, DB, Redis, queue); alerting on error rate, p95, queue lag; centralized error tracking. |
| NFR-10 | Data integrity | Fee ledger append-only (no silent edits; corrections are reversing entries); attendance edits audited; results versioned on republish. |
| NFR-11 | Compliance | GST-compliant fee receipts (invoice fields); DPDP Act 2023 for consent & data rights; school data export; audit trails retained ≥ 3 years. |
| NFR-12 | Localization | English + Hindi first; currency INR; date formats configurable; timezone Asia/Kolkata default, per-school configurable. |

---

## 5. User Stories

**Platform / Admin**
1. As a platform admin, I can onboard a school with branding and session config so that the school is live without engineering help.
2. As a school admin, I can add 500 students from a CSV so that I don't retype records.
3. As a school admin, I can see revenue and collection rates on my dashboard so that I can act on defaults early.
4. As a school admin, I can publish results so that parents see report cards instantly.
5. As an accountant, I can record a cash payment of ₹2,500 against a ₹10,000 fee and issue a receipt so that partial payments are tracked.

**Teacher**
6. As a teacher, I can mark attendance for my class in under a minute so that I don't spend the period on registers.
7. As a teacher, I can assign homework with a file and deadline so that students know what to do.
8. As a teacher, I can enter marks for my subject so that results are computed automatically.
9. As a teacher, I only see my classes and subjects so that I can't accidentally touch other teachers' data.

**Parent / Student**
10. As a parent, I get a push when homework is assigned or fees are due so that I don't miss deadlines.
11. As a parent, I can pay fees online and download receipts so that I don't visit the school office.
12. As a student, I can see my attendance and marks so that I can track my progress.

---

## 6. Use Cases (primary flows)

| UC | Name | Primary actor | Flow summary |
|---|---|---|---|
| UC-1 | Tenant onboarding | Platform admin | Fill school details → set session → branding → create super-admin → school active |
| UC-2 | Student enrollment | School admin | Add/edit student → assign class/division+session → upload docs → guardian link |
| UC-3 | Bulk student import | School admin | Upload CSV → validate → preview errors → commit |
| UC-4 | Mark daily attendance | Teacher | Pick class/division/date → bulk mark → adjust exceptions → save (audited) |
| UC-5 | Attendance analytics | School admin | Filter class/date-range → chart + export |
| UC-6 | Assign homework | Teacher | Create homework → target class/subject → attach file → set due date → publish → notify |
| UC-7 | Submit & evaluate homework | Student/Teacher | Student submits → teacher reviews → remarks/grade → notify student |
| UC-8 | Create & schedule exam | School admin | Define exam + subject schedule → admit cards → marks entry → publish results → report cards |
| UC-9 | Enter marks | Teacher | Select exam/subject/class → grid entry → validate → publish subject result |
| UC-10 | Set fee structure | Accountant | Define heads, amounts, frequency, class scope, due dates → generate student ledgers |
| UC-11 | Collect fee (offline) | Accountant | Find student → outstanding shown → partial/full payment → receipt → ledger update |
| UC-12 | Pay fee online | Parent | Dashboard due fees → Razorpay order → UPI/card → webhook → ledger credit + receipt + notification |
| UC-13 | Defaulter list | Accountant | Filter class/date/amount → list + export + send reminders |
| UC-14 | Run monthly payroll | Accountant | Generate salary slips → mark paid → notify teachers + payslips |
| UC-15 | Targeted announcement | School admin | Draft → choose audience (school/class/division/student) → schedule → publish → read receipts |
| UC-16 | Parent registration | Parent | OTP verify phone → link children (invite code/approval) → dashboard |
| UC-17 | Result publication alert | System | Admin publishes results → notification service fans out push+email to parents |
| UC-18 | Fee-due reminder | System (scheduled) | Daily job → due within N days → notify via preference channels |

---

## 7. Out of Scope (v1.0)

- Full web admin console (deferred; admin actions run on tablet-optimized mobile app first).
- Library management, transport tracking (GPS), hostel management, HR leave module — designed for later phases (see Phase 8).
- Offline-first sync (reads cached; writes require connectivity in v1).
- Multi-language beyond English/Hindi.
- SMS delivery (adapter interface only).
