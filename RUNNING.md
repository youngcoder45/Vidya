# Running Vidya locally

Two ways to try it: **A) everything in Docker** (recommended — no Go/Flutter needed
for the API), or **B) local dev** (Go on host). The Flutter app always needs the
Flutter SDK.

---

## Prerequisites

| Tool | Needed for | Get it |
|---|---|---|
| Docker (with Compose v2) | API + Postgres + Redis (path A) | https://docs.docker.com/get-docker/ |
| Go 1.23+ | API (path B only) | https://go.dev/dl/ |
| Flutter SDK (stable) | Mobile app | https://docs.flutter.dev/get-started/install |

---

## A. Docker (fastest)

```bash
# from the repo root
docker compose -f deploy/docker-compose.yml up --build
```

First boot does three things automatically (dev mode):
1. Creates the schema (GORM AutoMigrate for scaffold modules)
2. Seeds roles/permissions + a demo school with 2 students
3. Starts the API on `http://localhost:8080`

Wait until healthy, then verify:

```bash
curl -s http://localhost:8080/health/ready
# {"status":"ready","db":"up"}

# Demo login → token
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"principal@greenwood.edu","password":"admin12345","device":{"device_id":"demo-device"}}'
```

Copy `data.access_token` from the response into a variable and explore:

```bash
TOKEN="<paste access_token here>"

# Dashboard
curl -s http://localhost:8080/api/v1/dashboard/summary -H "Authorization: Bearer $TOKEN"

# Students
curl -s "http://localhost:8080/api/v1/students?session_id=<session id>" -H "Authorization: Bearer $TOKEN"

# Try the fee flow:
#   1. POST /api/v1/fees/heads        {"name":"Tuition Fee","code":"TUITION","category":"tuition","frequency":"monthly"}
#   2. POST /api/v1/fees/structures   {"session_id":..., "fee_head_id":..., "amount_inr":25000, "due_day":10}
#   3. POST /api/v1/fees/structures/generate  {"session_id":...}
#   4. POST /api/v1/fees/payments/offline     {"student_id":..., "mode":"cash", "amount_inr":25000}
```

Stop everything: `Ctrl+C`, then `docker compose -f deploy/docker-compose.yml down`
(add `-v` to also wipe the database volumes).

> **One-time demo creds** (seeded): school admin `principal@greenwood.edu` / `admin12345`,
> platform admin `admin@vidya.app` / `admin12345`.

---

## B. Local dev (Go on host, DB in Docker)

```bash
# 1. Postgres + Redis only
docker compose -f deploy/docker-compose.yml up -d postgres redis

# 2. API
cd backend
cp .env.example .env        # defaults match the compose DB credentials
make run                    # or: go run ./cmd/server
# APP_SEED_ON_START=true in .env seeds the demo school on first boot
```

---

## Mobile app (Flutter)

```bash
cd mobile

# First time only: generate the android/ios platform folders
flutter create . --platforms=android,ios

flutter pub get
flutter run                # pick a device when prompted
```

**API URL** the app talks to (`API_BASE_URL`, default `http://10.0.2.2:8080/api/v1`):

| Device | Command |
|---|---|
| Android emulator | default works (10.0.2.2 = host machine) |
| iOS simulator / macOS | `flutter run --dart-define=API_BASE_URL=http://localhost:8080/api/v1` |
| Physical phone | `flutter run --dart-define=API_BASE_URL=http://<your-LAN-IP>:8080/api/v1` |

Android note: plain `http://` is blocked by default on Android 9+ — for dev add
`android:usesCleartextTraffic="true"` to the `<application>` tag in
`android/app/src/main/AndroidManifest.xml` (or use a debug network security config).

---

## Troubleshooting

| Symptom | Fix |
|---|---|
| `curl: connection refused` on 8080 | Wait for `api` container to be healthy (`docker compose ps`) |
| Login returns 401 `INVALID_CREDENTIALS` | Seed didn't run — set `APP_SEED_ON_START=true`, recreate containers (`down -v`, `up --build`) |
| `role "vidya" does not exist` | The `postgres` volume predates the DB — `docker compose down -v` and up again |
| Port 5432/6379 already in use | Change `ports` in `deploy/docker-compose.yml` or stop the conflicting service |
| Flutter: no `android/` folder | Run `flutter create . --platforms=android,ios` in `mobile/` |

For deeper detail: `docs/07-deployment.md` (deploy), `docs/04-api.md` (endpoints).
