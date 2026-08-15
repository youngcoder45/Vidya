# Phase 4 — API Design

**Base URL:** `https://api.schoolos.app/api/v1` · **Content-Type:** `application/json` · **Auth:** `Authorization: Bearer <access_token>`
**Tenant header:** `X-School-ID` (optional — JWT claim is authoritative; header used for pre-auth routes & webhooks)

---

## 1. Conventions

- **Envelope** — success: `{"data": ...}`; error: `{"error": {"code": "FEES_DUE_EXISTS", "message": "...", "details": {...}}}`
- **Errors** — HTTP statuses: 400 validation, 401 unauthenticated, 403 forbidden, 404 not found, 409 conflict, 422 business-rule violation, 429 rate limited, 500 internal.
- **Pagination** — `?page=1&limit=20` → `{"data": [...], "meta": {"page": 1, "limit": 20, "total": 342}}`
- **Filtering** — `?class_division_id=...&status=active&search=ram`
- **Sorting** — `?sort=-created_at` (minus = desc)
- **Idempotency** — money-creating endpoints accept `Idempotency-Key` header.
- **Dates** — RFC3339 UTC; **money** — integer paise (INR) or `{amount_inr, currency}` with 2dp serialized as string to avoid float drift. v1 uses integer paise.
- **Audit** — sensitive mutations recorded server-side; client never trusts client-supplied identity.

---

## 2. Authentication Flow

### 2.1 Staff password login (email/phone + password)

```mermaid
sequenceDiagram
    participant C as Client
    participant API as API
    participant DB as PostgreSQL
    participant RD as Redis

    C->>API: POST /auth/login {identifier, password, device}
    API->>RD: rate limit check (5/min per identifier+IP)
    API->>DB: load user by email/phone + school
    API->>API: verify bcrypt; check status
    API->>DB: create auth_session (refresh hash), upsert device
    API->>RD: cache role/permission set (TTL 15m)
    API-->>C: 200 {access_token (15m), refresh_token (30d), user, school, roles, permissions}
```

- Access token: JWT HS256/RS256, claims `sub` (user), `school_id`, `roles[]`, `permissions[]`, `device_id`, `jti`, `exp`, `iat`.
- Refresh token: opaque 256-bit random stored **hashed** in `auth_sessions`; rotated on every use; reuse of a revoked token revokes the whole session family (detection log).
- Device tracking: `user_devices` upserted on login; `POST /auth/devices/:id/revoke` kills session; push token updates via `PUT /auth/devices/me`.

### 2.2 Parent OTP login (phone)

```
1. POST /auth/otp/request  {phone, purpose: parent_login}
   → 200 {expires_in: 300}   (OTP stored hashed in otp_codes, 3 attempts, 60s resend)
2. POST /auth/otp/verify   {phone, otp, device}
   → 200 {access_token, refresh_token, user, children: [...]}
```

### 2.3 Token refresh

```
POST /auth/refresh {refresh_token, device_id}
→ 200 {access_token, refresh_token}   (old refresh rotated & invalidated)
```

### 2.4 Webhooks (Razorpay) — signed, no JWT

```
POST /webhooks/razorpay   (X-Razorpay-Signature verified with webhook secret)
```

---

## 3. Endpoint Catalog (v1)

### Auth (`/auth`)
| Method | Path | Access | Description |
|---|---|---|---|
| POST | `/auth/login` | public | staff login |
| POST | `/auth/otp/request` | public | send OTP |
| POST | `/auth/otp/verify` | public | verify OTP → tokens |
| POST | `/auth/refresh` | public | rotate refresh token |
| POST | `/auth/logout` | auth | revoke session + device |
| GET | `/auth/me` | auth | profile + roles + permissions |
| GET | `/auth/devices` | auth | list own devices |
| PUT | `/auth/devices/me` | auth | update fcm token / name |
| POST | `/auth/devices/:id/revoke` | auth | logout a device |
| POST | `/auth/password/change` | auth | change password |
| POST | `/auth/password/forgot` | public | email/OTP reset |

### Tenants / School (`/school` — platform admin)
| Method | Path | Access |
|---|---|---|
| POST | `/schools` | platform_admin |
| GET | `/schools/:id` | platform_admin |
| PUT | `/schools/:id` | platform_admin |
| POST | `/schools/:id/suspend` | platform_admin |

### Sessions & Setup (`/academic`)
| Method | Path | Access |
|---|---|---|
| GET/POST | `/sessions` | school_admin |
| PUT | `/sessions/:id` | school_admin |
| POST | `/sessions/:id/activate` | school_admin |
| GET/POST | `/classes` | school_admin |
| GET/POST | `/divisions` | school_admin |
| GET/POST | `/classes/:id/divisions` | school_admin (class-division mapping per session) |
| GET/POST | `/subjects` | school_admin |
| POST | `/classes/:id/subjects` | school_admin (assign subject/teacher) |
| GET/POST | `/teachers` | school_admin |
| PUT | `/teachers/:id` | school_admin |

### Students (`/students`)
| Method | Path | Access |
|---|---|---|
| GET | `/students` | school_admin, teacher |
| POST | `/students` | school_admin |
| GET | `/students/:id` | school_admin, teacher, parent(own child) |
| PUT | `/students/:id` | school_admin |
| DELETE | `/students/:id` | school_admin (soft) |
| POST | `/students/import` | school_admin (CSV) |
| POST | `/students/:id/enroll` | school_admin (session/class/roll) |
| POST | `/students/:id/promote` | school_admin (next class, next session) |
| GET | `/students/:id/documents` | school_admin |
| POST | `/students/:id/documents` | school_admin (multipart → S3) |
| DELETE | `/students/:id/documents/:docId` | school_admin |
| GET | `/students/:id/guardians` | school_admin, parent |
| POST | `/students/:id/guardians` | school_admin |
| GET | `/students/:id/history` | school_admin (academic history) |

### Attendance (`/attendance`)
| Method | Path | Access |
|---|---|---|
| GET | `/attendance/classes/:cdId/daily?date=` | teacher (own class), school_admin |
| POST | `/attendance/daily` | teacher, school_admin (bulk upsert) |
| PUT | `/attendance/records/:id` | teacher, school_admin (edit + audit) |
| GET | `/attendance/students/:id?from=&to=` | school_admin, parent(own child) |
| GET | `/attendance/analytics?class_division_id=&from=&to=` | school_admin, teacher |
| GET | `/attendance/export?class_division_id=&month=` | school_admin |

### Homework (`/homework`)
| Method | Path | Access |
|---|---|---|
| GET/POST | `/homework` | teacher (create), school_admin (view), parent/student (view own) |
| GET | `/homework/:id` | as above, scoped |
| PUT | `/homework/:id` | teacher (own), school_admin |
| DELETE | `/homework/:id` | teacher (own), school_admin |
| POST | `/homework/:id/submissions` | student (own), parent(own child) |
| GET | `/homework/:id/submissions` | teacher (own class), school_admin |
| PUT | `/homework/submissions/:id` | teacher (grade/remarks) |
| GET | `/homework/analytics?class_division_id=` | school_admin, teacher |

### Assignments (`/assignments`) — mirrors homework with rubric/resources
| Method | Path | Access |
|---|---|---|
| GET/POST | `/assignments` | teacher/school_admin |
| PUT/DELETE | `/assignments/:id` | teacher (own)/school_admin |
| POST | `/assignments/:id/submissions` | student |
| PUT | `/assignments/submissions/:id` | teacher |
| GET | `/assignments/analytics` | school_admin, teacher |

### Exams & Results (`/exams`)
| Method | Path | Access |
|---|---|---|
| GET/POST | `/exams` | school_admin (create), teacher (own classes view) |
| PUT | `/exams/:id` | school_admin |
| POST | `/exams/:id/schedules` | school_admin (subject schedule) |
| GET | `/exams/:id/admit-cards` | school_admin (PDF) |
| GET/POST | `/grading-systems` | school_admin |
| GET | `/exams/:id/marks?subject_id=` | teacher (own subject), school_admin |
| PUT | `/exams/:id/marks` | teacher (grid upsert, draft) |
| POST | `/exams/:id/subjects/:subjectId/publish` | teacher (publish subject marks) |
| POST | `/exams/:id/publish` | school_admin (publish full results) |
| GET | `/exams/:id/report-cards` | school_admin (PDF batch) |
| GET | `/exams/:id/analytics` | school_admin |

### Fees (`/fees`)
| Method | Path | Access |
|---|---|---|
| GET/POST | `/fees/heads` | school_admin, accountant |
| PUT | `/fees/heads/:id` | school_admin, accountant |
| GET/POST | `/fees/structures` | school_admin, accountant |
| PUT | `/fees/structures/:id` | school_admin, accountant |
| POST | `/fees/structures/generate` | accountant (create student ledgers for session) |
| GET | `/fees/students/:id/ledger` | accountant, school_admin, parent(own child) |
| GET | `/fees/dues?class_division_id=&status=` | accountant, school_admin |
| GET | `/fees/dues/defaulter?class_division_id=&days=&amount=` | accountant, school_admin |
| POST | `/fees/orders` | parent(own child), accountant, school_admin (Razorpay order) |
| POST | `/fees/payments/offline` | accountant (cash/cheque/upi capture) |
| POST | `/fees/payments/allocate` | accountant (split across heads) |
| GET | `/fees/students/:id/payments` | accountant, parent(own child) |
| GET | `/fees/receipts/:paymentId` | accountant, parent(own child) (PDF) |
| GET | `/fees/reports/collection?from=&to=` | accountant, school_admin |
| GET | `/fees/reports/outstanding-ageing` | accountant, school_admin |
| POST | `/fees/refunds` | school_admin |

### Payroll (`/payroll`)
| Method | Path | Access |
|---|---|---|
| GET/POST | `/payroll/structures` | school_admin, accountant |
| PUT | `/payroll/structures/:id` | school_admin, accountant |
| POST | `/payroll/runs` | accountant (generate month) |
| POST | `/payroll/runs/:id/finalize` | accountant |
| POST | `/payroll/runs/:id/pay` | accountant (mark paid, notify) |
| GET | `/payroll/slips?teacher_id=&month=` | accountant, teacher (own) |
| GET | `/payroll/reports?from=&to=` | accountant, school_admin |

### Announcements & Calendar (`/announcements`, `/calendar`)
| Method | Path | Access |
|---|---|---|
| GET/POST | `/announcements` | school_admin (create+target), all (read own feed) |
| PUT/DELETE | `/announcements/:id` | school_admin |
| POST | `/announcements/:id/publish` | school_admin (schedule honored) |
| POST | `/announcements/:id/read` | all (mark read) |
| GET | `/announcements/:id/read-receipts` | school_admin |
| GET/POST | `/calendar/events` | school_admin (create), all (view) |
| PUT/DELETE | `/calendar/events/:id` | school_admin |

### Notifications (`/notifications`)
| Method | Path | Access |
|---|---|---|
| GET | `/notifications?type=&unread=` | all (own) |
| POST | `/notifications/:id/read` | all |
| POST | `/notifications/read-all` | all |
| GET | `/notifications/preferences` | all |
| PUT | `/notifications/preferences` | all |

### Dashboard (`/dashboard`)
| Method | Path | Access |
|---|---|---|
| GET | `/dashboard/summary` | school_admin (counts, attendance %, pending fees, upcoming exams, revenue) |
| GET | `/dashboard/revenue?from=&to=` | school_admin, accountant |
| GET | `/dashboard/attendance-trend?days=` | school_admin |

### Platform (`/admin` — platform super admin)
| Method | Path |
|---|---|
| GET | `/admin/schools` |
| GET | `/admin/schools/:id/usage` |
| PUT | `/admin/schools/:id/plan` |

### Ops
| Method | Path | Notes |
|---|---|---|
| GET | `/health` | liveness + dependency status |
| GET | `/health/ready` | readiness |
| GET | `/metrics` | Prometheus (internal) |

---

## 4. Representative Request/Response Schemas

### POST /auth/login
```jsonc
// request
{
  "identifier": "principal@school.edu",
  "password": "••••••••",
  "device": {"device_id": "fcm-reg-xyz", "name": "Mi 11", "platform": "android", "fcm_token": "..."}
}
// 200 response
{
  "data": {
    "access_token": "eyJ...", "expires_in": 900, "token_type": "Bearer",
    "refresh_token": "a3f...", 
    "user": {"id": "u-1", "name": "Meera Iyer", "email": "principal@school.edu", "phone": "+9198...", "avatar_url": null},
    "school": {"id": "s-1", "name": "Greenwood Public School", "branding": {"primary_color": "#1B5E20", "logo_url": "..."}},
    "roles": ["school_admin"], "permissions": ["students.*", "fees.*", "payroll.*", "exams.*"]
  }
}
```

### POST /students
```jsonc
{
  "first_name": "Aarav", "last_name": "Sharma", "dob": "2015-04-12", "gender": "male",
  "blood_group": "B+", "admission_no": "ADM-2026-0142", "address": "12 MG Road, Pune",
  "class_division_id": "cd-9", "roll_no": 23, "session_id": "sess-26",
  "guardians": [{"name": "Rajesh Sharma", "relation": "father", "phone": "+9198...", "email": "rajesh@x.in", "is_primary": true}]
}
// 201 → student profile + enrollment
```

### POST /fees/orders (parent pays online)
```jsonc
{
  "student_id": "stu-77",
  "ledger_ids": ["fl-1", "fl-2"],
  "amount_inr": 25000,
  "currency": "INR"
}
// 201
{"data": {"order_id": "fpo-9", "razorpay_order_id": "order_P4x9...", "key_id": "rzp_live_...", "amount_inr": 25000}}
```

### POST /fees/payments/offline (accountant captures cash)
```jsonc
{
  "student_id": "stu-77",
  "mode": "cash",
  "amount_inr": 25000,
  "allocations": [{"fee_ledger_id": "fl-1", "amount_inr": 15000}, {"fee_ledger_id": "fl-2", "amount_inr": 10000}],
  "notes": "Cash collected at office"
}
// 201 → receipt
{"data": {"payment_id": "pay-4", "receipt_no": "RCP-2026-00042", "receipt_url": "https://cdn.schoolos.app/receipts/..."}}
```

### POST /attendance/daily
```jsonc
{
  "class_division_id": "cd-9", "date": "2026-08-14",
  "records": [
    {"student_id": "stu-77", "status": "present"},
    {"student_id": "stu-78", "status": "absent"},
    {"student_id": "stu-79", "status": "late"}
  ]
}
// 200 → {"data": {"saved": 42, "skipped_duplicates": 0}}
```

### GET /dashboard/summary → 200
```jsonc
{"data": {
  "students": {"active": 1482, "per_class": {"Class 6-A": 38}},
  "teachers": 54,
  "attendance_today": {"percentage": 94.2, "present": 1396, "total": 1482},
  "pending_fees": {"count": 214, "amount_inr": 1834500},
  "upcoming_exams": [{"id": "ex-3", "name": "Term 1", "starts_on": "2026-09-15"}],
  "recent_announcements": [{"id": "an-8", "title": "PTM on Sep 5", "published_at": "..."}],
  "revenue": {"collected_inr": 8420000, "due_inr": 10254500, "collection_rate": 82.1}
}}
```

### Webhook POST /webhooks/razorpay
```jsonc
{
  "entity": "event", "event": "payment.captured",
  "payload": {"payment": {"entity": {"id": "pay_Nx7...", "order_id": "order_P4x9...", "amount": 2500000, "currency": "INR", "status": "captured"}}}
}
```

---

## 5. Error Model

```jsonc
{
  "error": {
    "code": "FEE_LEDGER_ALREADY_SETTLED",
    "message": "This fee item is already fully paid.",
    "details": {"fee_ledger_id": "fl-1", "paid_amount_inr": 25000}
  }
}
```
Common codes: `VALIDATION_ERROR`, `UNAUTHENTICATED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `RATE_LIMITED`, `TENANT_MISMATCH`, `INVALID_OTP`, `OTP_EXPIRED`, `TOKEN_REVOKED`, `FEE_DUE_EXISTS`, `PAYMENT_ALREADY_CAPTURED`, `EXAM_RESULT_LOCKED`.

## 6. Versioning & Evolution

- URL versioning `/api/v1`; additive changes within v1; breaking changes bump version.
- Deprecation: header `Deprecation` + 6-month overlap window.
