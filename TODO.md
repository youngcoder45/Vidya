# SchoolOS — TODO

> Development roadmap and implementation checklist for SchoolOS.

Legend:

* `[ ]` Not started
* `[-]` In progress
* `[x]` Completed
* `🔴` Critical
* `🟡` High priority
* `🟢` Normal priority
* `⚪` Nice to have

---

> **Progress note (2026-09-26):** A hardening pass landed across
> authentication, payments, attendance, the event bus, CI, and migrations.
> See [`report.md`](report.md) for the full finding-by-finding audit and the
> remediation progress section (75 findings, many now fixed).

---

# 🚀 Current Focus

The immediate goal is to move the existing architecture and scaffold toward a complete, production-ready School ERP.

* [ ] 🟡 Complete Homework module
* [ ] 🟡 Complete Examinations module
* [ ] 🟡 Complete Payroll module
* [ ] 🟡 Expand Flutter application beyond authentication/dashboard
* [ ] 🟡 Improve automated test coverage
* [ ] 🟡 Harden production security
* [ ] 🟡 Complete production deployment configuration

---

# 🏫 Core ERP Modules

## Authentication

* [x] JWT authentication
* [x] Refresh token support
* [x] Device tracking
* [x] RBAC foundation
* [x] Tenant-aware authentication
* [ ] 🔴 Production secret management
* [ ] 🟡 Password reset flow
* [ ] 🟡 Email verification
* [ ] 🟡 Parent OTP authentication
* [ ] 🟢 Session/device management UI
* [ ] 🟢 Account lockout / suspicious login detection

---

## Multi-Tenancy

* [x] `school_id` tenant model
* [x] Tenant context
* [x] Repository-level tenant scoping
* [x] PostgreSQL RLS architecture
* [ ] 🔴 Full RLS policy implementation
* [ ] 🔴 Cross-tenant access test suite
* [ ] 🟡 Tenant administration UI
* [ ] 🟡 School onboarding workflow
* [ ] 🟢 Tenant usage metrics
* [ ] 🟢 Tenant configuration management

---

# 👨‍🎓 Students

* [x] Student schema
* [x] Student API foundation
* [ ] 🟡 Student CRUD UI
* [ ] 🟡 Student profile
* [ ] 🟡 Class/section assignment
* [ ] 🟡 Parent/student relationship
* [ ] 🟢 Student search and filtering
* [ ] 🟢 Bulk student import
* [ ] 🟢 Student document management
* [ ] 🟢 Student transfer workflow
* [ ] 🟢 Student archival

---

# 📋 Attendance

* [x] Attendance schema
* [x] Attendance API foundation
* [ ] 🟡 Teacher attendance UI
* [ ] 🟡 Daily attendance workflow
* [ ] 🟡 Class-wise attendance
* [ ] 🟢 Attendance history
* [ ] 🟢 Attendance reports
* [ ] 🟢 Monthly attendance summaries
* [ ] 🟢 Parent attendance notifications
* [ ] ⚪ Attendance export

---

# 💰 Fees & Payments

* [x] Fee schema
* [x] Append-only ledger design
* [x] Offline payments
* [x] Razorpay order creation
* [x] Razorpay webhook architecture
* [x] Idempotent payment handling
* [ ] 🔴 Production Razorpay configuration
* [ ] 🟡 Fee structure management
* [ ] 🟡 Student fee dashboard
* [ ] 🟡 Parent payment UI
* [ ] 🟡 Payment history
* [ ] 🟢 Receipts
* [ ] 🟢 Refund workflow
* [ ] 🟢 Late fee support
* [ ] 🟢 Discounts / scholarships
* [ ] 🟢 Fee reports
* [ ] ⚪ Financial exports

---

# 📚 Homework

**Status: Schema + API stub**

* [ ] 🟡 Homework CRUD
* [ ] 🟡 Teacher homework creation
* [ ] 🟡 Class/subject assignment
* [ ] 🟡 Homework deadlines
* [ ] 🟡 Student homework view
* [ ] 🟡 Submission workflow
* [ ] 🟢 File attachments
* [ ] 🟢 Teacher evaluation
* [ ] 🟢 Parent notifications
* [ ] ⚪ Homework analytics

---

# 📝 Examinations

**Status: Schema + API stub**

* [ ] 🟡 Examination CRUD
* [ ] 🟡 Exam scheduling
* [ ] 🟡 Subjects and papers
* [ ] 🟡 Marks entry
* [ ] 🟡 Marks validation
* [ ] 🟡 Student results
* [ ] 🟢 Grade calculation
* [ ] 🟢 Report cards
* [ ] 🟢 Result publishing
* [ ] 🟢 Parent result access
* [ ] ⚪ Result analytics
* [ ] ⚪ Rank/position calculations

---

# 👨‍🏫 Teachers & Staff

* [ ] 🟡 Teacher profiles
* [ ] 🟡 Staff profiles
* [ ] 🟡 Teacher-class assignments
* [ ] 🟡 Subject assignments
* [ ] 🟢 Staff attendance
* [ ] 🟢 Teacher timetable
* [ ] 🟢 Staff leave management
* [ ] 🟢 Teacher dashboard
* [ ] ⚪ Staff performance reports

---

# 💼 Payroll

**Status: Schema + API stub**

* [ ] 🟡 Employee salary structure
* [ ] 🟡 Payroll periods
* [ ] 🟡 Salary calculation
* [ ] 🟡 Deductions
* [ ] 🟢 Bonuses
* [ ] 🟢 Payslips
* [ ] 🟢 Payroll history
* [ ] 🟢 Staff payment records
* [ ] ⚪ Payroll reports

---

# 📢 Announcements

* [x] Announcement backend
* [ ] 🟡 Announcement management UI
* [ ] 🟡 School-wide announcements
* [ ] 🟡 Class-specific announcements
* [ ] 🟢 Teacher announcements
* [ ] 🟢 Scheduled announcements
* [ ] 🟢 Announcement expiry
* [ ] 🟢 Read/unread tracking

---

# 🔔 Notifications

* [x] Notification module foundation
* [ ] 🟡 Push notification provider
* [ ] 🟡 Notification preferences
* [ ] 🟡 Notification center
* [ ] 🟢 Email notifications
* [ ] 🟢 SMS integration
* [ ] 🟢 Parent notification workflows
* [ ] 🟢 Notification templates
* [ ] ⚪ Notification analytics

---

# 📊 Dashboard

* [x] Dashboard API
* [x] Dashboard module
* [ ] 🟡 Student statistics
* [ ] 🟡 Attendance statistics
* [ ] 🟡 Fee statistics
* [ ] 🟡 Homework statistics
* [ ] 🟢 Examination statistics
* [ ] 🟢 Teacher statistics
* [ ] 🟢 Role-specific dashboards
* [ ] ⚪ Custom dashboard widgets

---

# 📱 Flutter Application

## Core

* [x] Flutter project structure
* [x] Riverpod
* [x] go_router
* [x] Theme system
* [x] Light mode
* [x] Dark mode
* [ ] 🟡 API client layer
* [ ] 🟡 Global error handling
* [ ] 🟡 Loading states
* [ ] 🟡 Offline state handling
* [ ] 🟢 Local caching
* [ ] 🟢 Secure token storage

## UI

* [x] Authentication screens
* [x] Dashboard foundation
* [ ] 🟡 Student screens
* [ ] 🟡 Attendance screens
* [ ] 🟡 Fees screens
* [ ] 🟡 Announcements
* [ ] 🟡 Notifications
* [ ] 🟡 Homework
* [ ] 🟡 Examinations
* [ ] 🟡 Teacher portal
* [ ] 🟡 Parent portal
* [ ] 🟢 Payroll
* [ ] 🟢 Settings
* [ ] 🟢 Profile management

---

# 🎨 Design System

* [x] Design tokens
* [x] Semantic theme roles
* [x] Light theme
* [x] Dark theme
* [x] Theme tests
* [ ] 🟡 Typography scale
* [ ] 🟡 Spacing system
* [ ] 🟡 Component library
* [ ] 🟢 Form components
* [ ] 🟢 Data tables
* [ ] 🟢 Empty states
* [ ] 🟢 Error states
* [ ] 🟢 Loading/skeleton components
* [ ] ⚪ Accessibility audit

---

# 🔌 API

* [x] API foundation
* [x] Authentication endpoints
* [x] Tenant endpoints
* [x] Student endpoints
* [x] Attendance endpoints
* [x] Fee endpoints
* [x] Announcement endpoints
* [x] Notification endpoints
* [x] Dashboard endpoints
* [ ] 🟡 Complete API documentation
* [ ] 🟡 OpenAPI specification
* [ ] 🟡 Request validation improvements
* [ ] 🟢 API versioning strategy
* [ ] 🟢 Pagination standardization
* [ ] 🟢 Filtering/sorting conventions
* [ ] 🟢 API rate limiting
* [ ] ⚪ GraphQL evaluation

---

# 🗄️ Database

* [x] PostgreSQL schema
* [x] Initial migration
* [x] Core indexes
* [x] Foreign keys
* [x] Multi-tenant schema design
* [ ] 🔴 Production RLS policies
* [ ] 🟡 Migration rollback strategy
* [ ] 🟡 Database backup strategy
* [ ] 🟢 Query performance audit
* [ ] 🟢 Index optimization
* [ ] 🟢 Database seed improvements
* [ ] ⚪ Read replica evaluation

---

# ⚡ Redis & Background Jobs

* [x] Redis integration foundation
* [ ] 🟡 Production Redis configuration
* [ ] 🟡 Rate limiting
* [ ] 🟡 Dashboard caching
* [ ] 🟢 Background job system
* [ ] 🟢 Job retries
* [ ] 🟢 Scheduled jobs
* [ ] 🟢 Notification queues
* [ ] ⚪ Redis-backed distributed locking

---

# 🔒 Security

* [ ] 🔴 Production JWT key management
* [ ] 🔴 Secrets management
* [ ] 🔴 Full tenant-isolation audit
* [ ] 🔴 Security headers
* [ ] 🔴 Production CORS configuration
* [ ] 🔴 Rate limiting
* [ ] 🟡 Input validation audit
* [ ] 🟡 Authorization audit
* [ ] 🟡 Dependency vulnerability scanning
* [ ] 🟡 Audit logging
* [ ] 🟢 Security test suite
* [ ] 🟢 Automated dependency updates
* [ ] ⚪ External penetration testing

---

# 🧪 Testing

## Backend

* [ ] 🟡 Unit test coverage
* [ ] 🟡 Repository tests
* [ ] 🟡 Service tests
* [ ] 🟡 API integration tests
* [ ] 🔴 Tenant isolation tests
* [ ] 🔴 Authentication security tests
* [ ] 🟢 Payment webhook tests
* [ ] 🟢 Database migration tests
* [ ] 🟢 Load testing

## Flutter

* [x] Theme tests
* [ ] 🟡 Widget test coverage
* [ ] 🟡 Unit tests
* [ ] 🟢 Integration tests
* [ ] 🟢 Authentication flow tests
* [ ] 🟢 Payment flow tests
* [ ] 🟢 Offline-mode tests

---

# 🚢 Deployment

* [x] Docker configuration
* [x] Docker Compose
* [x] GitHub Actions foundation
* [x] Prometheus configuration
* [ ] 🟡 Production Docker configuration
* [ ] 🟡 Production environment configuration
* [ ] 🟡 HTTPS/TLS
* [ ] 🟡 Automated deployments
* [ ] 🟡 Database backups
* [ ] 🟡 Disaster recovery
* [ ] 🟢 Log aggregation
* [ ] 🟢 Alerting
* [ ] 🟢 Health monitoring
* [ ] 🟢 Deployment rollback strategy
* [ ] ⚪ Kubernetes evaluation

---

# 📈 Observability

* [x] Health endpoints
* [x] Prometheus foundation
* [ ] 🟡 API metrics
* [ ] 🟡 Database metrics
* [ ] 🟡 Redis metrics
* [ ] 🟢 Structured logging
* [ ] 🟢 Distributed request IDs
* [ ] 🟢 Error tracking
* [ ] 🟢 Alerting
* [ ] ⚪ Distributed tracing

---

# 📖 Documentation

* [x] Architecture documentation
* [x] Requirements documentation
* [x] Database documentation
* [x] API documentation foundation
* [x] Frontend architecture documentation
* [x] Backend architecture documentation
* [x] Deployment documentation
* [x] Roadmap
* [x] README
* [x] CONTRIBUTING.md
* [x] CODE_OF_CONDUCT.md
* [x] LICENSE
* [x] SETUP.md
* [ ] 🟡 Complete API reference
* [ ] 🟡 Developer onboarding guide
* [ ] 🟢 Deployment runbook
* [ ] 🟢 Troubleshooting guide
* [ ] 🟢 Architecture decision records
* [ ] ⚪ User documentation

---

# 🧹 Code Quality

* [ ] 🟡 Standardize error handling
* [ ] 🟡 Improve logging
* [ ] 🟡 Remove technical debt
* [ ] 🟡 Review dependency usage
* [ ] 🟢 Improve naming consistency
* [ ] 🟢 Refactor duplicated logic
* [ ] 🟢 Improve code documentation
* [ ] 🟢 Add static analysis
* [ ] ⚪ Benchmark critical paths

---

# 🚀 Performance

* [ ] 🟡 API response-time benchmarks
* [ ] 🟡 Database query profiling
* [ ] 🟡 Dashboard query optimization
* [ ] 🟢 Redis cache optimization
* [ ] 🟢 Flutter rendering performance
* [ ] 🟢 Large-list virtualization
* [ ] 🟢 API pagination
* [ ] ⚪ Load testing at realistic tenant scale

---

# 🌍 Future Ideas

These are intentionally not committed to the immediate roadmap.

* [ ] Multi-language support
* [ ] Regional education-board support
* [ ] Advanced analytics
* [ ] AI-assisted administrative workflows
* [ ] Parent mobile application improvements
* [ ] Teacher mobile application improvements
* [ ] Web administration portal
* [ ] Attendance QR system
* [ ] Biometric attendance integration
* [ ] Transport management
* [ ] Library management
* [ ] Inventory management
* [ ] Hostel management
* [ ] Timetable management
* [ ] Certificate generation
* [ ] Digital document storage
* [ ] WhatsApp notification integration
* [ ] Advanced reporting
* [ ] Public API
* [ ] Third-party integrations

---

# 🏁 Milestones

## Milestone 1 — Core Foundation

* [x] Architecture
* [x] Database
* [x] Authentication
* [x] Tenant isolation foundation
* [x] Students
* [x] Attendance
* [x] Fees
* [x] Announcements
* [x] Notifications
* [x] Dashboard

## Milestone 2 — Complete ERP

* [ ] Homework
* [ ] Examinations
* [ ] Payroll
* [ ] Teacher workflows
* [ ] Parent workflows
* [ ] Expanded Flutter UI

## Milestone 3 — Production Readiness

* [ ] Security audit
* [ ] Tenant isolation audit
* [ ] Automated testing
* [ ] Production infrastructure
* [ ] Monitoring
* [ ] Backups
* [ ] Disaster recovery
* [ ] Production payment configuration

## Milestone 4 — Scale

* [ ] Performance optimization
* [ ] Background jobs
* [ ] Advanced caching
* [ ] Observability
* [ ] Load testing
* [ ] Scaling strategy

---

# 📌 Notes

This TODO is intentionally broader than the immediate development sprint.

Tasks may be moved, split, reprioritized, or removed as the architecture and product requirements evolve.

For the detailed implementation roadmap, see:

[`docs/08-roadmap.md`](docs/08-roadmap.md)

---

**SchoolOS** — Building a scalable, multi-tenant school management platform.
