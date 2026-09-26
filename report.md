# SchoolOS — Audit, Bug Report & Improvement Plan

**Date:** 2026-09-26
**Scope reviewed:** `backend/` (Go modular monolith), `mobile/` (Flutter), `deploy/`, `.github/workflows/`, root docs.
**Method:** static review of the full backend Go source, Flutter source, migration SQL, seeds, compose, Dockerfile and CI. The Go toolchain is not installed in this environment, so findings were derived by reading code rather than compiling; items marked `[BUILD]` are deterministic config mismatches.

**Baseline quality:** This is an unusually well-structured scaffold — clear module boundaries, a real multi-tenant model, an append-only fee ledger, token-driven theming, and honest documentation. The findings below are mostly *correctness and hardening* gaps rather than architectural rewrites. Several are genuine, exploitable bugs.

---

## How to read this

- **Severity:** 🔴 Critical (security/data-loss/build-breaking) · 🟠 High · 🟡 Medium · 🟢 Low
- Every finding lists the file, the concrete problem, and the fix.
- Findings are numbered 1–75 so the plan below can reference them.

**Verification note:** Permission-wildcard, rate-limiter, receipt-numbering, webhook-signature, tenant-login and Go-version issues were traced line-by-line through the code and are high-confidence. DB-constraint findings (foreign keys, unique indexes) depend on whether dev `AutoMigrate` or the production SQL schema is in use — which is itself the source of finding #58.

---

## Executive summary of the highest-risk issues

1. **Platform superadmins are locked out of every guarded endpoint** (#1) — the `"*"` wildcard permission never matches, so `/platform/schools/*` returns 403.
2. **Login is not tenant-scoped** (#2) — `FindUserByIdentity` is called with `nil` school, so a duplicate email/phone can authenticate into the wrong school.
3. **Razorpay webhooks are unauthenticated when the secret is unset, and production never requires it** (#4) — forged `payment.captured` events can mark fees paid.
4. **The dev in-memory rate limiter never resets** (#30) — after the limit is hit, a client is blocked indefinitely.
5. **The build is broken by a Go version mismatch** (#53) — `go.mod` requires Go 1.25 while CI and the Dockerfile pin 1.23.
6. **OTPs are written to logs in every environment** (#5).

---

## Remediation progress (2026-09-26)

Started applying the plan, one small commit per fix. Landed so far:

- **Build/CI/deploy:** #51 Go version, #52 phantom `worker` service, #53 real CI migration check, #54 SQL migration runner + `make migrate`, #57 `.dockerignore`, #58 forward-reference FKs in `001_init.sql`, #60 govulncheck + golangci-lint.
- **Auth/security:** #1 wildcard permission (+ test), #2 ambiguous login, #3 nullable platform session, #4 required payment/webhook secrets, #5 OTP logging, #6 non-bypassable auth limit, #8 JWT issuer/alg, #9 password-reset guard, #11 refresh limit, #12 bcrypt OTPs, #13 webhook body cap, #14 prod CORS, #15 login timing, #16 config validation, #17 security headers, #18/#19 audit logging + fee-payment audit.
- **Multi-tenancy/correctness:** #21 announcement visibility, #23 distinct count, #25 student mass assignment, #26 attendance roster, #27 student existence, #28 duplicate ledgers, #29 promotion target, #30 rate-limiter reset, #31/#32 receipt numbering, #33/#34/#35 webhook ack semantics, #36 unique indexes, #37 attendance edit trail, #38 dashboard timezone, #39 dashboard errors, #42 empty-device logout, #44 bounded event pool + real `fees.paid` publish, #45 preference-aware fan-out, #47 prior-OTP invalidation, #48 scheduled publication, #49 due-date anchoring.
- **Mobile/docs:** #61 bounded refresh retry, #62 token cleanup, #64 random device id, #66 PATCH helper, #68 radius token, #70/#71 README license, #75 migrate target.
- **Retracted:** #20 (the `[TEMPLATE]` line does not exist; read-tool artifact).

Still open (larger, feature-shaped): #7 access-token revocation, #10 refresh device binding, #22 RLS enablement + cross-tenant tests, #24 scoped role grants, #40 role-scoped dashboard, #41 `/auth/me` status check, #43 refresh-reuse audit, #46 parent auto-provision race, #50 orphan orders, plus the mobile screens (#63, #67, #72) and the deferred TODO backlog.

## Phased improvement plan

### Phase 0 — Stop the bleeding (do first, small diffs)
Fix the build/CI mismatch (#53, #54, #55, #56), the superadmin lockout (#1), webhook secret enforcement (#4), OTP logging (#5), the non-resetting rate limiter (#30), and tenant-scoped login (#2). These are each one-to-few lines and unblock everything else.

### Phase 1 — Correctness & financial integrity
Receipt numbering (#31, #32), webhook idempotency/ack semantics (#33, #34, #35), attendance upsert index mismatch (#36), mass-assignment on student update (#25), attendance input validation (#26), dashboard timezone (#39), announcement visibility (#21), dev/prod schema drift (#58), and scheduled-announcement publication (#49).

### Phase 2 — Security hardening
Account lockout and non-bypassable auth throttling (#6), device binding on refresh (#10), token revocation (#7), password-reset flow (#9), security headers (#17), CORS/secret config validation (#14, #16), webhook body limits (#13), OTP throttling (#48), audit coverage (#18, #19), RLS enablement (#22).

### Phase 3 — Mobile product completeness
Bound the refresh retry (#61), fix cold-start restore (#62, #65), wire bottom-nav routing (#67), implement parent OTP login (#72), add error/loading/offline states (#63), add widget + unit tests (#69).

### Phase 4 — Scale & operations
Background jobs (#45), notification fan-out with preferences/templates (#46), dashboard caching, structured metrics, backup/DR, load testing, dependency scanning (#60), deployment runbook.

---

## Findings

### A. Authentication, authorization & security

1. 🔴 **Superadmin wildcard permission never matches.** `HasPermission` (`backend/internal/pkg/ctxuser/ctxuser.go:44`) returns true only for an exact match or a `module.*` suffix. A superadmin is issued `permissions: ["*"]` (`auth/service.go` `issueTokens`), but `"*"` has length 1, so the `len(p) > 2` branch can never match `students.read`. Result: `RequirePermission` denies the platform admin everywhere. **Fix:** treat `p == "*"` as a global grant and add a unit test for it.

2. 🔴 **Login is not tenant-scoped.** `Service.Login` calls `repo.FindUserByIdentity(ctx, nil, identifier)` (`auth/service.go`), and the repository applies `school_id` only when non-nil (`auth/repository.go:42`). The schema allows `UNIQUE(school_id, email)`, i.e. the same email in two schools. `First()` then returns an arbitrary row, potentially authenticating a user into the wrong tenant. **Fix:** require a tenant hint (school code/subdomain) at login, scope the lookup, or make email/phone globally unique for staff.

3. 🟠 **Superadmin sessions violate the production FK.** For a user with `SchoolID == nil`, `issueTokens` creates `AuthSession{SchoolID: uuid.Nil}`. The production schema declares `auth_sessions.school_id NOT NULL REFERENCES schools(id)` (`migrations/001_init.sql`). The zero UUID is not a valid school, so platform-admin login fails at insert in production. **Fix:** make `AuthSession.SchoolID` nullable, or add a platform pseudo-tenant.

4. 🔴 **Unsigned Razorpay webhooks accepted; production never requires the secret.** `VerifyWebhookSignature` returns `true` when `secret == ""` (`fees/service.go:195`), and `config.Load` validates only `APP_JWT_SECRET` and `APP_ENC_KEY` for production — never `APP_RAZORPAY_WEBHOOK_SECRET`. With the secret unset, anyone who knows a `gateway_order_id` can POST a forged `payment.captured`. **Fix:** fail production boot when the webhook secret is missing, and reject unsigned webhooks outside dev.

5. 🔴 **OTP codes are logged in all environments.** `RequestOtp` logs `"otp", code` at info level (`auth/service.go:103`). In production this leaks a working credential into log aggregation. **Fix:** log only a masked reference, or gate behind `AppEnv == "local"`.

6. 🟠 **Auth rate limiting is bypassable and there is no account lockout.** `AuthRateLimit` keys on `c.GetHeader("X-Identifier")` (`middleware/ratelimit.go`), a client-controlled value, so rotating the header defeats the limit. There is also no per-account failed-attempt counter. **Fix:** key on the IP plus a hash of the *parsed* identifier from the request body, and add lockout/backoff after N failures.

7. 🟠 **Access tokens cannot be revoked.** Logout revokes the refresh session but the already-issued JWT stays valid for its full TTL (`JWT_ACCESS_TTL`, 15m). Password change has the same gap. **Fix:** add a token-version/`jti` denylist (Redis) or shorten access TTL with sliding refresh; check version on parse.

8. 🟡 **JWT not validated for issuer/audience.** `jwtutil.Parse` accepts any HS256 token signed with the secret and never checks `iss`/`aud`. **Fix:** set and validate `Issuer`/`Audience`.

9. 🟠 **Advertised password reset is not implemented and conflates with login.** `otpRequest` accepts `purpose=password_reset`, but no reset endpoint exists; `VerifyOtp` issues a full login token for that purpose. **Fix:** block or implement the flow; never issue a normal session for a reset-purpose OTP.

10. 🟠 **Refresh token is not bound to the device.** `Service.Refresh` (`auth/service.go:160`) ignores the relationship between `session.DeviceID` and the request `device_id`, accepting any value and creating the new session under it. A stolen refresh token can be rotated onto a new device id. **Fix:** require `session.DeviceID == deviceID` (or re-bind only after re-auth).

11. 🟡 **Refresh endpoint is not rate-limited.** `/auth/refresh` is registered without `AuthRateLimit` (`auth/handler.go` `RegisterPublic`), unlike login/OTP. **Fix:** apply a limiter.

12. 🟡 **OTP hashing is unsalted SHA-256 over a 6-digit space.** A DB leak allows trivial brute-force of valid OTPs. **Fix:** use a slow KDF (bcrypt/argon2) or HMAC with a server pepper, and keep the 5-minute expiry + attempt count.

13. 🟡 **Webhook body is read without a size limit.** `io.ReadAll(c.Request.Body)` (`fees/handler.go`) has no cap — an attacker can stream a huge body and exhaust memory. **Fix:** `http.MaxBytesReader` (e.g. 256 KB).

14. 🟡 **CORS defaults to `*` and is not validated for production.** `APP_CORS_ORIGINS` defaults to `*` (`config.go`), and `CORS` reflects any origin (`middleware/cors.go`). **Fix:** warn/fail in production when `*` is configured.

15. 🟢 **Login reveals user existence via timing.** `bcrypt.Verify` runs only when a user is found, so a missing identifier returns faster. **Fix:** run a dummy bcrypt comparison on the not-found path.

16. 🟡 **Config validation is incomplete.** Production requires only JWT secret and encryption key. Rate limits of 0 silently block all traffic, and Razorpay/CORS aren't validated. **Fix:** validate `RateLimitPerMin > 0`, `AuthRateLimitPerMin > 0`, and required production integrations.

17. 🟡 **No security headers.** Responses lack `Strict-Transport-Security`, `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`, `Content-Security-Policy`. **Fix:** add a security-headers middleware (or terminate at the proxy with documented config).

18. 🟢 **Audit writer silently swallows failures.** `audit.Record` ignores errors with an empty block and no log (`pkg/audit/audit.go`). **Fix:** log at warn and emit a metric; optionally fail sensitive ops when audit cannot be written.

19. 🟡 **Audit coverage is thin.** Only OTP login writes an audit entry. Fee capture, offline payments, student edits, role changes, and school creation are not audited. **Fix:** record `audit.Entry` for all money/RBAC/PII mutations.

20. ✅ **RETRACTED.** The reported `[TEMPLATE]` line in `backend/.env.example` does not exist — it was a read-tool artifact, not file content. Verified with `cat -A`; no change required.

### B. Multi-tenancy & data isolation

21. 🟠 **Announcements leak drafts and non-applicable audiences.** `Announcements.list` and `get` (`announcements/repository.go`) filter only by school and optional status; they never enforce audience scope or `status = published` for readers. Roles like `parent`/`student` hold `announcements.read`, so they can read drafts and staff-only/other-student announcements. **Fix:** role-aware visibility: non-staff see published, non-expired announcements whose audience matches them.

22. 🟠 **RLS is designed but not enabled.** `migrations/001_init.sql` ships the RLS section commented out; there are no `CREATE POLICY` statements and the app never sets `app.school_id`. Tenant isolation rests entirely on repository discipline. **Fix:** implement the per-request `SET LOCAL app.school_id` and enable policies; add a cross-tenant test suite (also listed in `TODO.md`).

23. 🟡 **`ListStudents` count is inflated by joins.** `ListStudents` joins `student_enrollments` but calls `Count(&total)` without `Distinct`, while the data query does use `Distinct("students.*")` (`students/repository.go`). A student with multiple matching enrollment rows is counted more than once. **Fix:** `Distinct("students.id").Count`.

24. 🟢 **Role lookup is not tenant-aware.** `FindRoleByCode` queries the global roles catalog without `school_id` (`auth/repository.go`). Acceptable for a global catalog, but the intent should be documented, or use `user_roles.scope_class_division_id` for scoped grants (currently unused).

25. 🔴 **Mass assignment on student update.** `students.update` calls `c.ShouldBindJSON(existing)` (`students/handler.go`) then persists. This (a) lets a client overwrite fields such as `status` (e.g. un-drop a student), `id`, and `school_id`-adjacent state, and (b) zeroes any field omitted from the JSON because `UpdateStudent` writes every column from the map. **Fix:** bind into a dedicated request DTO with field allow-list and use a partial-update map.

26. 🟠 **Attendance mark accepts untrusted students and dates.** `attendance.mark` (`attendance/handler.go`) validates the status string but never checks that each `student_id` belongs to the school/class, nor that `date` is not in the future or absurdly far in the past. Cross-tenant/enrollment garbage can be inserted. **Fix:** verify roster membership and a sane date window before upsert.

27. 🟡 **Offline payment relies on incidental protection.** `CaptureOfflinePayment` doesn't confirm the student belongs to the school; it happens to fail only because a foreign student has no ledgers. Any student with ledger rows always belongs to the school, so it's safe today but fragile. **Fix:** explicit `GetStudent(schoolID, id)` check first.

28. 🟡 **Ledger generation can double-charge.** `GenerateLedgers` de-dupes on `(student, fee_head)` only. A whole-school structure (nil class) plus a class-specific structure for the same head both match, and recurring monthly heads never produce new periods. **Fix:** de-dupe on `(student, head, period)` and define recurrence semantics.

29. 🟡 **`Promote` doesn't validate targets.** It doesn't confirm `toClassDivisionID` belongs to `toSessionID`/school, nor guard against an existing active enrollment for the same session. **Fix:** validate the target and enforce the enrollment uniqueness constraint.

### C. Correctness bugs (backend)

30. 🔴 **In-memory rate limiter never resets its window.** `allowLocal` (`middleware/ratelimit.go:78`) increments a counter and only clears the map when it exceeds 10,000 keys. Once a key hits `limit`, that client is permanently blocked until the map is swept. **Fix:** store `{count, windowStart}` and reset per window.

31. 🔴 **Receipt numbers are not concurrency-safe.** `nextReceiptNo` (`fees/service.go:287`) computes `count + 1` for the year. Two concurrent payments get the same number; the unique index on `receipt_no` then rejects one and potentially rolls back a legitimate payment. **Fix:** use a DB sequence or `INSERT ... ON CONFLICT` retry, or an atomic per-school counter row.

32. 🔴 **Receipt numbering collides across tenants.** `receipt_no` is globally `UNIQUE` in the migration, but the number is generated per school (`RCP-2026-00001`). The second school to issue receipt #1 fails. **Fix:** include the school code in the number or make the unique key `(school_id, receipt_no)`.

33. 🟡 **Webhook for an unknown order returns 500.** `HandleWebhook` → `GetOrderByGatewayID` returns `gorm.ErrRecordNotFound`, which `WriteError` maps to a generic 500. Razorpay treats that as a delivery failure and retries forever. **Fix:** ack 200 for unmatchable events (and log), or return a specific 4xx contract.

34. 🟡 **Over-allocation on webhook returns 500.** If the captured amount exceeds outstanding dues, `computeAllocations` returns `ErrOutstandingMismatch`, producing a 500 and endless retries. **Fix:** record the payment as unallocated/overpaid and ack 200, then resolve manually.

35. 🟡 **Webhook trusts payload amount over the stored order.** `amountINR := p.Amount / 100` comes from the webhook entity; it's never reconciled with `order.AmountINR`. **Fix:** verify `p.Amount/100 == order.AmountINR` and alert on mismatch.

36. 🔴 **Attendance upsert targets a unique index that dev never creates.** `UpsertDaily` uses `clause.OnConflict{Columns: [school_id, class_division_id, student_id, date]}` (`attendance/repository.go`), but the `AttendanceRecord` struct declares only non-unique indexes. In dev, `AutoMigrate` therefore has no matching constraint and the upsert errors. Same pattern in `notifications.UpsertPreference` and `announcements.MarkRead`. **Fix:** declare matching `uniqueIndex` tags on the structs, or rely solely on SQL migrations.

37. 🟢 **Attendance edits lose provenance.** `UpsertDaily` updates `edited_by`/`edited_at`, but the handler never sets them (it sets `MarkedBy`), so re-marks silently overwrite the editor trail. **Fix:** set `EditedBy`/`EditedAt` on conflict/update paths.

38. 🟡 **Dashboard "today" is computed in UTC, not tenant time.** `build` uses `time.Now().UTC().Truncate(24h)` (`dashboard/handler.go`) while tenants are `Asia/Kolkata`. Attendance marked before 05:30 IST counts as the previous day. **Fix:** resolve "today" using the school timezone.

39. 🟡 **Dashboard swallows all query errors.** Every count/scan ignores its `Error`, so a missing table or query failure yields silent zeros. **Fix:** propagate errors (or at least log + mark partial).

40. 🟡 **Dashboard exposes whole-school data to every role.** `dashboard.read` is granted to parent and student roles, but the payload contains school-wide counts, revenue, and pending fees. **Fix:** role-scoped summaries or a separate permission.

41. 🟢 **`/auth/me` returns the user regardless of status.** `Me` doesn't check `Status`, unlike login/refresh. A suspended user with a live token still gets their profile. **Fix:** check status.

42. 🟢 **Empty-device logout revokes unrelated sessions.** `Logout` calls `RevokeSessionsByDevice` even when `device_id` is empty, matching every device-less session for the user. **Fix:** skip when empty.

43. 🟢 **Refresh reuse-detection is unaudited.** Revoking the device family on suspected reuse is correct but not recorded and doesn't notify the user. **Fix:** audit + notify.

44. 🟡 **Event bus has no backpressure and the payment seam is dead.** `Bus.Publish` spawns an unbounded goroutine per handler per event (`events/bus.go`), and no module (including fees) ever publishes `fees.paid`. The notifications subscriber therefore never fires for payments. **Fix:** bounded worker queue + actually publish domain events.

45. 🟡 **Notification subscriber ignores preferences, templates, and channels.** `SubscribeToEvents` unconditionally writes an in-app notification; `notification_preferences` and `notification_templates` are unused. **Fix:** honor preferences per event/channel and render from templates.

46. 🟡 **Parent auto-provisioning can duplicate users.** On first OTP verify, `FindUserByIdentity` then `CreateUser` races; the DB has no unique constraint that catches it because `NULL` email/phone combinations are allowed. **Fix:** unique partial index + `ON CONFLICT`, or transaction with `SELECT ... FOR UPDATE`.

47. 🟡 **OTP request has no per-phone throttle and doesn't invalidate prior codes.** Many valid codes can coexist; `FindLatestOtp` picks the newest but older ones remain valid. **Fix:** invalidate previous codes and throttle per phone.

48. 🟡 **Scheduled announcements never publish.** `create` sets `status=scheduled` for future `publish_at`, but there is no scheduler/job to flip them to published; `list` for readers doesn't consider timing. **Fix:** add a background job or publish lazily on read.

49. 🟡 **`dueDateFor` ignores session applicability.** Due dates always anchor to the *current* month, ignoring `applicable_from`/`applicable_to` and multi-month recurring heads. **Fix:** derive due dates from the structure period.

50. 🟢 **`fees.CreateOrder` can orphan pending orders.** The DB order is created before the gateway call; a gateway failure leaves a `pending` row with no gateway id. **Fix:** create-after-gateway, or mark failed/cleanup on error.

### D. Build, CI, deployment

51. 🔴 `[BUILD]` **Go version mismatch breaks the build.** `backend/go.mod` declares `go 1.25`, but CI sets `GO_VERSION: "1.23"` (`.github/workflows/ci.yml:14`) and the image uses `FROM golang:1.23-alpine` (`backend/Dockerfile:8`). Go refuses to build a module requiring a newer toolchain. README/SETUP also say "Go 1.23+". **Fix:** pin all three to the same version.

52. 🟠 **Deploy job references a nonexistent service.** The deploy step runs `docker compose up -d --build api worker` (`.github/workflows/ci.yml:96`), but `deploy/docker-compose.yml` defines no `worker` service. **Fix:** remove `worker` or add the service.

53. 🟡 **The "SQL migration sanity" CI step is a no-op.** It starts Postgres and `pg_isready`s it, but never applies `001_init.sql`; `|| true` masks everything. **Fix:** run the migration against the container and fail on SQL errors.

54. 🟠 **Production never applies migrations.** The Dockerfile copies `migrations/` but nothing runs `golang-migrate`; `AutoMigrate` is dev-only (`db.go`). Deployed production would run against an unmigrated DB. **Fix:** add a migrate init step/entrypoint (and the CI step above to validate it).

55. 🟡 **Dev/prod schema drift.** Dev uses `AutoMigrate` over GORM structs; production uses `001_init.sql`. They differ (FKs, `CHECK`s, `DATE` vs timestamp, unique indexes #36, `branding jsonb`). Bugs pass dev and fail prod. **Fix:** use migrations in dev too, or generate migrations from the models and diff-check them in CI.

56. 🟡 **Compose publishes datastores and uses dev secrets.** `postgres:5432` and `redis:6379` are exposed on the host and the API runs with `APP_JWT_SECRET=dev-secret-change-me...`; there is no production override file. **Fix:** a compose override for prod (no DB ports, secrets from env/secret store).

57. 🟡 **No `.dockerignore`.** The backend build context can include `.env`, local binaries, and test artifacts. **Fix:** add `.dockerignore` (`.env`, `bin/`, `.git`).

58. 🟡 **No dependency/vulnerability scanning.** `TODO.md` lists it as planned but CI runs neither `govulncheck` nor `flutter pub outdated`/`dart pub audit`. **Fix:** add scans to CI.

### E. Mobile (Flutter)

59. 🔴 **Token-refresh retry is unbounded.** `ApiClient._request` retries the original request after `_refreshTokens()` with no attempt counter (`core/network/api_client.dart`). If the refreshed token is still rejected (revoked device, server bug), it recurses indefinitely. **Fix:** retry at most once, then surface `UNAUTHENTICATED`.

60. 🟡 **Cold-start restore loses tenant/role context and doesn't clear bad tokens.** `restoreSession` returns `school_id: null` and `roles: []`; on failure `_restore` sets unauthenticated but leaves tokens in storage. **Fix:** return the full session from `/auth/me` (or decode the JWT) and clear tokens on hard failure.

61. 🟡 **No global error, loading, or offline handling.** The TODOs are real: only the dashboard shows error text; there are no skeletons, retry affordances, or connectivity handling. **Fix:** a shared `AsyncValue`-aware view layer and an offline banner.

62. 🟡 **Bottom navigation is inert.** `AppScaffold` renders Dashboard/Students/Fees/Notices, but `onBottomNavTap` is a no-op in `DashboardScreen` and none of those routes exist. **Fix:** add go_router branches and handlers.

63. 🟠 **Parent OTP login is a dead-end.** The login screen's "Parent? Login with OTP" button does nothing, while the backend OTP endpoints exist. **Fix:** implement the OTP screen/flow or remove the affordance.

64. 🟡 **Test coverage is one theme test.** `mobile/test/theme_test.dart` is the only test. No widget, controller, or API-client tests (and no backend tests at all beyond `go vet`). **Fix:** prioritize auth/API-client and dashboard widget tests, plus backend handler/service tests.

65. 🟢 **Device id is a timestamp.** `SecureStore.deviceId` uses `microsecondsSinceEpoch`; predictable and easily spoofed. **Fix:** use a random UUID stored once.

66. 🟢 **`ApiClient` has no `patch` and the interceptor rebuilds Dio per provider watch.** Minor; add `patch` for parity with CORS methods and cache the Dio instance.

67. 🟢 **`AppConfig` ships an emulator default in all builds.** `10.0.2.2` is the default base URL; a release build without `--dart-define` would point at a dev host. **Fix:** require `API_BASE_URL` in release or fail fast.

68. 🟢 **Theme token misuse.** `StatTile` uses `context.spaceMd` as an `InkWell` border radius (spacing vs radii token). **Fix:** use `AppRadii`.

69. 🟢 **Hardcoded-color rule is convention only.** `analysis_options.yaml` excludes platform dirs and adds no custom lint despite comments claiming CI/review enforcement. **Fix:** add a `custom_lint` rule or a grep-based check in CI.

### F. Documentation & DX

70. 🟡 **README's "compiling, runnable" claim is currently false** given the Go toolchain mismatch (#51), and it lists a Go version that disagrees with `go.mod`. **Fix:** align versions and wording.

71. 🟢 **License ambiguity.** README says "This project is currently unlicensed" while `LICENSE` exists; `CONTRIBUTING.md` refers to "the project's applicable license." **Fix:** pick a license (or mark it proprietary) and make all docs consistent.

72. 🟢 **API docs are prose, not a spec.** There is no OpenAPI document; clients hand-roll models (already drifting, e.g. dashboard models). **Fix:** generate OpenAPI and typed clients.

73. 🟢 **No migration rollback or backup/DR runbook**, both tracked as TODO. **Fix:** write the runbook and add rollback files per migration.

74. 🟢 **Observability is a placeholder.** `/metrics` returns a text stub; no request/db/redis metrics, structured error tracking, or alerting. **Fix:** wire `prometheus/client_golang` and error tracking.

75. 🟢 **Contributor onboarding gaps.** `Makefile` has no `migrate` target (docs reference one), and `SETUP.md` doesn't mention the Go 1.25 requirement. **Fix:** add `make migrate`, `make seed`, and align setup docs.

---

## Suggested first pull requests

| PR | Findings | Why first |
| -- | -------- | --------- |
| `fix: align Go toolchain versions` | #51, #52 | Restores CI/build; everything else is blocked until this lands. |
| `fix: grant wildcard permissions to platform admins` | #1 | One-line logic fix with a unit test; unblocks tenant administration. |
| `fix: require razorpay webhook secret in production` | #4, #20 | Prevents forged payment capture. |
| `fix: reset in-memory rate-limit windows` | #30 | Removes a self-inflicted denial of service. |
| `fix: reject unsigned webhooks` | #4, #13, #33, #34 | Small, high-value payment-integrity batch. |
| `fix: tenant-scope login lookups` | #2, #3 | Closes cross-tenant authentication. |
| `fix: atomic, tenant-unique receipt numbers` | #31, #32 | Prevents payment failures and collisions. |

---

## Notes for the follow-up work

- Add `backend/internal/**/*_test.go` tests as each fix lands; the repo currently has none, so regressions are invisible.
- Enable RLS (#22) together with a cross-tenant integration test harness — that combination is worth more than either alone.
- Treat `TODO.md` as the long-range backlog; this report is the *correctness* delta between the scaffold and its own claims.
