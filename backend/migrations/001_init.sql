-- ============================================================================
-- SchoolOS — canonical production schema (Phase 3 deliverable)
-- Applied by golang-migrate in CI/CD. In dev, GORM AutoMigrate covers the
-- implemented scaffold modules; this file is the single source of truth for
-- the full design and for production.
--
-- Conventions:
--   * UUIDv4 primary keys
--   * tenant tables lead with school_id and index it
--   * money stored as bigint (INR, integer units)
--   * timestamps timestamptz, UTC
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ---------------------------------------------------------------------------
-- Platform & tenant
-- ---------------------------------------------------------------------------

CREATE TABLE plans (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT NOT NULL,
    price_inr      BIGINT NOT NULL DEFAULT 0,
    max_students   INT NOT NULL DEFAULT 0,
    max_teachers   INT NOT NULL DEFAULT 0,
    feature_flags  JSONB NOT NULL DEFAULT '{}',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE schools (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id    UUID REFERENCES plans(id),
    name       TEXT NOT NULL,
    code       TEXT NOT NULL UNIQUE,
    board      TEXT,
    address    TEXT,
    phone      TEXT,
    email      TEXT,
    timezone   TEXT NOT NULL DEFAULT 'Asia/Kolkata',
    currency   TEXT NOT NULL DEFAULT 'INR',
    branding   JSONB NOT NULL DEFAULT '{}',      -- {primary_color, logo_url, letterhead}
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE academic_sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID NOT NULL REFERENCES schools(id),
    name       TEXT NOT NULL,
    starts_on  DATE NOT NULL,
    ends_on    DATE NOT NULL,
    is_active  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_school ON academic_sessions (school_id);

-- ---------------------------------------------------------------------------
-- RBAC catalog (platform-level) + membership (tenant-scoped)
-- ---------------------------------------------------------------------------

CREATE TABLE roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL,
    description TEXT
);

CREATE TABLE permissions (
    id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code   TEXT NOT NULL UNIQUE,
    name   TEXT NOT NULL,
    module TEXT NOT NULL
);

CREATE TABLE role_permissions (
    role_id       UUID NOT NULL REFERENCES roles(id),
    permission_id UUID NOT NULL REFERENCES permissions(id),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id     UUID REFERENCES schools(id),   -- NULL = platform admin
    full_name     TEXT NOT NULL,
    email         TEXT,
    phone         TEXT,
    password_hash TEXT NOT NULL,
    avatar_url    TEXT,
    status        TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','suspended')),
    is_superadmin BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, phone),
    UNIQUE (school_id, email)
);
CREATE INDEX idx_users_email ON users (email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_phone ON users (phone) WHERE phone IS NOT NULL;

CREATE TABLE user_roles (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id               UUID NOT NULL REFERENCES schools(id),
    user_id                 UUID NOT NULL REFERENCES users(id),
    role_id                 UUID NOT NULL REFERENCES roles(id),
    scope_class_division_id UUID REFERENCES class_divisions(id), -- nullable: teacher class scope
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_user_roles_user ON user_roles (school_id, user_id);

CREATE TABLE user_devices (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    user_id     UUID NOT NULL REFERENCES users(id),
    device_id   TEXT NOT NULL,
    device_name TEXT,
    platform    TEXT,
    fcm_token   TEXT,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_devices_user ON user_devices (school_id, user_id);

CREATE TABLE auth_sessions (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id          UUID NOT NULL REFERENCES schools(id),
    user_id            UUID NOT NULL REFERENCES users(id),
    device_id          TEXT,
    refresh_token_hash TEXT NOT NULL,
    ip                 TEXT,
    user_agent         TEXT,
    expires_at         TIMESTAMPTZ NOT NULL,
    revoked_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_sessions_user ON auth_sessions (school_id, user_id, revoked_at);

CREATE TABLE otp_codes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    phone       TEXT,
    email       TEXT,
    purpose     TEXT NOT NULL,     -- parent_login | password_reset
    code_hash   TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    attempts    INT NOT NULL DEFAULT 0,
    verified_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_otp_lookup ON otp_codes (school_id, phone, purpose, created_at DESC);

-- ---------------------------------------------------------------------------
-- Academics
-- ---------------------------------------------------------------------------

CREATE TABLE class_groups (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID NOT NULL REFERENCES schools(id),
    name       TEXT NOT NULL,
    level      INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_class_groups_school ON class_groups (school_id);

CREATE TABLE divisions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID NOT NULL REFERENCES schools(id),
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_divisions_school ON divisions (school_id);

CREATE TABLE class_divisions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    session_id        UUID NOT NULL REFERENCES academic_sessions(id),
    class_group_id    UUID NOT NULL REFERENCES class_groups(id),
    division_id       UUID NOT NULL REFERENCES divisions(id),
    class_teacher_id  UUID REFERENCES teachers(id),
    strength          INT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, session_id, class_group_id, division_id)
);
CREATE INDEX idx_class_divisions_session ON class_divisions (school_id, session_id);

CREATE TABLE subjects (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    name        TEXT NOT NULL,
    code        TEXT,
    is_language BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_subjects_school ON subjects (school_id);

CREATE TABLE teachers (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id              UUID NOT NULL REFERENCES schools(id),
    user_id                UUID NOT NULL REFERENCES users(id),
    employee_code          TEXT NOT NULL,
    designation            TEXT,
    join_date              DATE,
    qualification          TEXT,
    bank_details_encrypted TEXT,               -- AES-256-GCM app-level
    status                 TEXT NOT NULL DEFAULT 'active',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_teachers_school ON teachers (school_id);

CREATE TABLE class_subjects (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    session_id        UUID NOT NULL REFERENCES academic_sessions(id),
    class_division_id UUID NOT NULL REFERENCES class_divisions(id),
    subject_id        UUID NOT NULL REFERENCES subjects(id),
    teacher_id        UUID REFERENCES teachers(id),
    is_elective       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_class_subjects_class ON class_subjects (school_id, class_division_id);

-- ---------------------------------------------------------------------------
-- Students
-- ---------------------------------------------------------------------------

CREATE TABLE students (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id     UUID NOT NULL REFERENCES schools(id),
    admission_no  TEXT NOT NULL,
    first_name    TEXT NOT NULL,
    last_name     TEXT,
    dob           DATE NOT NULL,
    gender        TEXT,
    blood_group   TEXT,
    address       TEXT,
    photo_url     TEXT,
    medical_notes TEXT,
    status        TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','dropped')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, admission_no)
);
CREATE INDEX idx_students_school ON students (school_id);

CREATE TABLE guardians (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID NOT NULL REFERENCES schools(id),
    user_id    UUID REFERENCES users(id),      -- set when parent registers
    name       TEXT NOT NULL,
    relation   TEXT,
    phone      TEXT NOT NULL,
    email      TEXT,
    occupation TEXT,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_guardians_school ON guardians (school_id);

CREATE TABLE student_guardians (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    student_id  UUID NOT NULL REFERENCES students(id),
    guardian_id UUID NOT NULL REFERENCES guardians(id),
    relation    TEXT,
    priority    INT NOT NULL DEFAULT 0,
    UNIQUE (school_id, student_id, guardian_id)
);
CREATE INDEX idx_student_guardians_student ON student_guardians (school_id, student_id);

CREATE TABLE student_enrollments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    session_id        UUID NOT NULL REFERENCES academic_sessions(id),
    student_id        UUID NOT NULL REFERENCES students(id),
    class_division_id UUID NOT NULL REFERENCES class_divisions(id),
    roll_no           INT,
    admission_date    DATE NOT NULL DEFAULT CURRENT_DATE,
    status            TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','promoted','dropped')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, session_id, student_id),
    UNIQUE (school_id, session_id, class_division_id, roll_no)
);
CREATE INDEX idx_enrollments_class ON student_enrollments (school_id, session_id, class_division_id);

CREATE TABLE student_documents (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID NOT NULL REFERENCES schools(id),
    student_id UUID NOT NULL REFERENCES students(id),
    doc_type   TEXT NOT NULL,
    file_url   TEXT NOT NULL,
    issued_on  DATE,
    expires_on DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_student_docs_student ON student_documents (school_id, student_id);

-- ---------------------------------------------------------------------------
-- Attendance
-- ---------------------------------------------------------------------------

CREATE TABLE attendance_records (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    class_division_id UUID NOT NULL REFERENCES class_divisions(id),
    subject_id        UUID REFERENCES subjects(id),
    student_id        UUID NOT NULL REFERENCES students(id),
    date              DATE NOT NULL,
    status            TEXT NOT NULL CHECK (status IN ('present','absent','late','half_day','leave')),
    marked_by         UUID NOT NULL REFERENCES users(id),
    edited_by         UUID REFERENCES users(id),
    edited_at         TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, class_division_id, student_id, date)
);
CREATE INDEX idx_attendance_class_date ON attendance_records (school_id, class_division_id, date);
CREATE INDEX idx_attendance_student ON attendance_records (school_id, student_id, date DESC);
CREATE INDEX idx_attendance_school_date ON attendance_records (school_id, date);

-- ---------------------------------------------------------------------------
-- Homework & assignments
-- ---------------------------------------------------------------------------

CREATE TABLE homeworks (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    session_id        UUID NOT NULL REFERENCES academic_sessions(id),
    class_division_id UUID NOT NULL REFERENCES class_divisions(id),
    subject_id        UUID NOT NULL REFERENCES subjects(id),
    teacher_id        UUID NOT NULL REFERENCES teachers(id),
    title             TEXT NOT NULL,
    description       TEXT,
    due_at            TIMESTAMPTZ NOT NULL,
    max_marks         INT,
    published_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_homework_class_due ON homeworks (school_id, class_division_id, due_at);

CREATE TABLE homework_attachments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    homework_id UUID NOT NULL REFERENCES homeworks(id),
    file_url    TEXT NOT NULL,
    file_name   TEXT,
    size        BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX idx_hw_attachments_hw ON homework_attachments (school_id, homework_id);

CREATE TABLE homework_submissions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    homework_id  UUID NOT NULL REFERENCES homeworks(id),
    student_id   UUID NOT NULL REFERENCES students(id),
    content      TEXT,
    file_url     TEXT,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    remarks      TEXT,
    marks        INT,
    graded_by    UUID REFERENCES users(id),
    graded_at    TIMESTAMPTZ,
    UNIQUE (school_id, homework_id, student_id)
);
CREATE INDEX idx_hw_submissions_hw ON homework_submissions (school_id, homework_id);

CREATE TABLE assignments (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    session_id        UUID NOT NULL REFERENCES academic_sessions(id),
    class_division_id UUID NOT NULL REFERENCES class_divisions(id),
    subject_id        UUID NOT NULL REFERENCES subjects(id),
    teacher_id        UUID NOT NULL REFERENCES teachers(id),
    title             TEXT NOT NULL,
    description       TEXT,
    due_at            TIMESTAMPTZ NOT NULL,
    rubric            JSONB NOT NULL DEFAULT '{}',
    resources         JSONB NOT NULL DEFAULT '[]',
    max_marks         INT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_assignments_class_due ON assignments (school_id, class_division_id, due_at);

CREATE TABLE assignment_submissions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    assignment_id     UUID NOT NULL REFERENCES assignments(id),
    student_id        UUID NOT NULL REFERENCES students(id),
    content           TEXT,
    file_url          TEXT,
    submitted_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    resubmission_count INT NOT NULL DEFAULT 0,
    remarks           TEXT,
    marks             INT,
    graded_by         UUID REFERENCES users(id),
    graded_at         TIMESTAMPTZ,
    UNIQUE (school_id, assignment_id, student_id)
);

-- ---------------------------------------------------------------------------
-- Exams & results
-- ---------------------------------------------------------------------------

CREATE TABLE exams (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    session_id        UUID NOT NULL REFERENCES academic_sessions(id),
    name              TEXT NOT NULL,
    exam_type         TEXT NOT NULL CHECK (exam_type IN ('unit','term','final')),
    class_division_id UUID NOT NULL REFERENCES class_divisions(id),
    starts_on         DATE NOT NULL,
    ends_on           DATE NOT NULL,
    status            TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','scheduled','ongoing','published')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_exams_school ON exams (school_id, session_id);

CREATE TABLE exam_schedules (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    exam_id           UUID NOT NULL REFERENCES exams(id),
    subject_id        UUID NOT NULL REFERENCES subjects(id),
    date              DATE NOT NULL,
    start_time        TIME,
    end_time          TIME,
    max_marks         INT NOT NULL,
    pass_marks        INT NOT NULL DEFAULT 0,
    UNIQUE (school_id, exam_id, subject_id)
);

CREATE TABLE grading_systems (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID NOT NULL REFERENCES schools(id),
    name       TEXT NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE grade_ranges (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    grading_system_id UUID NOT NULL REFERENCES grading_systems(id),
    grade             TEXT NOT NULL,
    grade_point       NUMERIC(4,2) NOT NULL,
    min_percent       NUMERIC(5,2) NOT NULL,
    max_percent       NUMERIC(5,2) NOT NULL,
    remark            TEXT
);

CREATE TABLE exam_marks (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    exam_schedule_id  UUID NOT NULL REFERENCES exam_schedules(id),
    student_id        UUID NOT NULL REFERENCES students(id),
    marks_obtained    NUMERIC(7,2),
    grade             TEXT,
    is_absent         BOOLEAN NOT NULL DEFAULT FALSE,
    entered_by        UUID NOT NULL REFERENCES users(id),
    status            TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','published')),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, exam_schedule_id, student_id)
);
CREATE INDEX idx_exam_marks_schedule ON exam_marks (school_id, exam_schedule_id);

CREATE TABLE result_publishes (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id      UUID NOT NULL REFERENCES schools(id),
    exam_id        UUID NOT NULL REFERENCES exams(id),
    published_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_by   UUID NOT NULL REFERENCES users(id),
    report_card_url TEXT
);
CREATE INDEX idx_result_publishes_exam ON result_publishes (school_id, exam_id);

CREATE TABLE report_cards (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    exam_id     UUID NOT NULL REFERENCES exams(id),
    student_id  UUID NOT NULL REFERENCES students(id),
    pdf_url     TEXT NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_report_cards_student ON report_cards (school_id, student_id, exam_id);

-- ---------------------------------------------------------------------------
-- Fees & payments (append-only money paths)
-- ---------------------------------------------------------------------------

CREATE TABLE fee_heads (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    name        TEXT NOT NULL,
    code        TEXT NOT NULL,
    category    TEXT NOT NULL CHECK (category IN ('tuition','transport','hostel','lab','exam','misc')),
    is_recurring BOOLEAN NOT NULL DEFAULT TRUE,
    frequency   TEXT NOT NULL DEFAULT 'monthly' CHECK (frequency IN ('monthly','quarterly','term','yearly')),
    tax_rate    NUMERIC(6,2) NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_fee_heads_school ON fee_heads (school_id);

CREATE TABLE fee_structures (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    session_id        UUID NOT NULL REFERENCES academic_sessions(id),
    fee_head_id       UUID NOT NULL REFERENCES fee_heads(id),
    class_division_id UUID REFERENCES class_divisions(id), -- NULL = whole school
    amount_inr        BIGINT NOT NULL CHECK (amount_inr >= 0),
    due_day           INT NOT NULL DEFAULT 10,
    applicable_from   DATE,
    applicable_to     DATE,
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_fee_structures_session ON fee_structures (school_id, session_id);

CREATE TABLE fee_ledger (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    session_id      UUID NOT NULL REFERENCES academic_sessions(id),
    student_id      UUID NOT NULL REFERENCES students(id),
    fee_head_id     UUID NOT NULL REFERENCES fee_heads(id),
    amount_inr      BIGINT NOT NULL CHECK (amount_inr >= 0),
    due_date        DATE NOT NULL,
    paid_amount_inr BIGINT NOT NULL DEFAULT 0,
    concession_inr  BIGINT NOT NULL DEFAULT 0,
    status          TEXT NOT NULL DEFAULT 'due' CHECK (status IN ('due','partial','paid','waived')),
    waived_by       UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_fee_ledger_student ON fee_ledger (school_id, student_id);
CREATE INDEX idx_fee_ledger_due ON fee_ledger (school_id, status, due_date);
CREATE INDEX idx_fee_ledger_session_head ON fee_ledger (school_id, session_id, fee_head_id);

CREATE TABLE fee_installments (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id      UUID NOT NULL REFERENCES schools(id),
    fee_ledger_id  UUID NOT NULL REFERENCES fee_ledger(id),
    installment_no INT NOT NULL,
    amount_inr     BIGINT NOT NULL,
    due_date       DATE NOT NULL,
    paid_at        TIMESTAMPTZ
);
CREATE INDEX idx_fee_installments_ledger ON fee_installments (school_id, fee_ledger_id);

CREATE TABLE fee_payment_orders (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id        UUID NOT NULL REFERENCES schools(id),
    student_id       UUID NOT NULL REFERENCES students(id),
    amount_inr       BIGINT NOT NULL,
    currency         TEXT NOT NULL DEFAULT 'INR',
    gateway          TEXT NOT NULL DEFAULT 'razorpay',
    gateway_order_id TEXT,
    status           TEXT NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending','processing','captured','failed','refunded')),
    idempotency_key  TEXT NOT NULL,
    created_by       UUID NOT NULL REFERENCES users(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, idempotency_key)
);
CREATE INDEX idx_orders_gateway ON fee_payment_orders (gateway_order_id);

CREATE TABLE payments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    student_id  UUID NOT NULL REFERENCES students(id),
    order_id    UUID REFERENCES fee_payment_orders(id),
    receipt_no  TEXT NOT NULL UNIQUE,
    amount_inr  BIGINT NOT NULL CHECK (amount_inr > 0),
    mode        TEXT NOT NULL CHECK (mode IN ('cash','upi','card','netbanking','cheque','online')),
    gateway_ref TEXT,
    paid_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    recorded_by UUID NOT NULL REFERENCES users(id),
    notes       TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payments_school_time ON payments (school_id, paid_at DESC);
CREATE INDEX idx_payments_student ON payments (school_id, student_id);

CREATE TABLE payment_allocations (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    payment_id   UUID NOT NULL REFERENCES payments(id),
    fee_ledger_id UUID NOT NULL REFERENCES fee_ledger(id),
    amount_inr   BIGINT NOT NULL CHECK (amount_inr > 0)
);
CREATE INDEX idx_allocations_payment ON payment_allocations (school_id, payment_id);

CREATE TABLE receipts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID NOT NULL REFERENCES schools(id),
    payment_id UUID NOT NULL REFERENCES payments(id),
    receipt_no TEXT NOT NULL UNIQUE,
    pdf_url    TEXT,
    issued_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_receipts_payment ON receipts (school_id, payment_id);

CREATE TABLE refunds (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    payment_id   UUID NOT NULL REFERENCES payments(id),
    amount_inr   BIGINT NOT NULL CHECK (amount_inr > 0),
    reason       TEXT,
    status       TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','processed')),
    processed_by UUID REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- Payroll
-- ---------------------------------------------------------------------------

CREATE TABLE salary_structures (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id      UUID NOT NULL REFERENCES schools(id),
    teacher_id     UUID NOT NULL REFERENCES teachers(id),
    effective_from DATE NOT NULL,
    basic_inr      BIGINT NOT NULL,
    components     JSONB NOT NULL DEFAULT '{}',   -- allowances/deductions breakdown
    monthly_net    BIGINT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_salary_structures_teacher ON salary_structures (school_id, teacher_id);

CREATE TABLE salary_runs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    month        INT NOT NULL CHECK (month BETWEEN 1 AND 12),
    year         INT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','finalized','paid')),
    total_amount BIGINT NOT NULL DEFAULT 0,
    run_by       UUID NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, month, year)
);

CREATE TABLE salary_slips (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id             UUID NOT NULL REFERENCES schools(id),
    salary_run_id         UUID NOT NULL REFERENCES salary_runs(id),
    teacher_id            UUID NOT NULL REFERENCES teachers(id),
    gross_inr             BIGINT NOT NULL,
    deductions_inr        BIGINT NOT NULL DEFAULT 0,
    net_inr               BIGINT NOT NULL,
    component_breakdown   JSONB NOT NULL DEFAULT '{}',
    slip_url              TEXT,
    paid_at               TIMESTAMPTZ,
    mode                  TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, salary_run_id, teacher_id)
);

-- ---------------------------------------------------------------------------
-- Communication
-- ---------------------------------------------------------------------------

CREATE TABLE announcements (
    id                         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id                  UUID NOT NULL REFERENCES schools(id),
    title                      TEXT NOT NULL,
    body                       TEXT NOT NULL,
    audience_scope             TEXT NOT NULL CHECK (audience_scope IN ('school','class','division','student','staff')),
    audience_class_division_id UUID REFERENCES class_divisions(id),
    audience_student_id        UUID REFERENCES students(id),
    publish_at                 TIMESTAMPTZ,
    expires_at                 TIMESTAMPTZ,
    status                     TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','scheduled','published','archived')),
    created_by                 UUID NOT NULL REFERENCES users(id),
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                 TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_announcements_status ON announcements (school_id, status, publish_at);

CREATE TABLE announcement_reads (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    announcement_id UUID NOT NULL REFERENCES announcements(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    read_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, announcement_id, user_id)
);

CREATE TABLE holidays_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id     UUID NOT NULL REFERENCES schools(id),
    session_id    UUID NOT NULL REFERENCES academic_sessions(id),
    title         TEXT NOT NULL,
    type          TEXT NOT NULL CHECK (type IN ('holiday','event','activity')),
    starts_on     DATE NOT NULL,
    ends_on       DATE NOT NULL,
    audience_scope TEXT NOT NULL DEFAULT 'school',
    description   TEXT,
    created_by    UUID NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_holidays_start ON holidays_events (school_id, starts_on);

CREATE TABLE notification_templates (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id     UUID REFERENCES schools(id), -- NULL = platform default
    event_type    TEXT NOT NULL,
    channel       TEXT NOT NULL CHECK (channel IN ('inapp','push','email','sms')),
    subject       TEXT,
    body_template TEXT
);

CREATE TABLE notification_preferences (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID NOT NULL REFERENCES schools(id),
    user_id    UUID NOT NULL REFERENCES users(id),
    event_type TEXT NOT NULL,
    channel    TEXT NOT NULL CHECK (channel IN ('inapp','push','email','sms')),
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    UNIQUE (school_id, user_id, event_type, channel)
);

CREATE TABLE notifications (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id  UUID NOT NULL REFERENCES schools(id),
    user_id    UUID NOT NULL REFERENCES users(id),
    type       TEXT NOT NULL,
    title      TEXT NOT NULL,
    body       TEXT,
    data       JSONB NOT NULL DEFAULT '{}',
    channel    TEXT NOT NULL DEFAULT 'inapp',
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notifications_user ON notifications (school_id, user_id, created_at DESC);

CREATE TABLE notification_deliveries (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id       UUID NOT NULL REFERENCES schools(id),
    notification_id UUID NOT NULL REFERENCES notifications(id),
    channel         TEXT NOT NULL CHECK (channel IN ('inapp','push','email','sms')),
    status          TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','sent','failed')),
    attempts        INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    delivered_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (school_id, notification_id, channel)
);
CREATE INDEX idx_deliveries_status ON notification_deliveries (status, created_at);

-- ---------------------------------------------------------------------------
-- Audit & ops
-- ---------------------------------------------------------------------------

CREATE TABLE audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID REFERENCES schools(id),   -- NULL for platform ops
    user_id     UUID REFERENCES users(id),
    action      TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   TEXT,
    changes     JSONB NOT NULL DEFAULT '{}',
    ip          TEXT,
    user_agent  TEXT,
    device_id   TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_entity ON audit_logs (school_id, entity_type, entity_id, created_at DESC);
CREATE INDEX idx_audit_user ON audit_logs (user_id, created_at DESC);

-- ============================================================================
-- Production hardening (apply in prod only): Row-Level Security
-- ============================================================================
-- ALTER TABLE ... ENABLE ROW LEVEL SECURITY;  (per tenant table)
-- CREATE POLICY tenant_isolation ON students
--   USING (school_id = NULLIF(current_setting('app.school_id', true), '')::uuid);
-- The app sets app.school_id per transaction; repositories enforce scoping
-- anyway, RLS is defense-in-depth.
-- ============================================================================
