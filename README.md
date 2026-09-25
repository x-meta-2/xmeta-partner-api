# xmeta-partner-api

Backend service for the X-Meta Partner Program — a partner affiliate system layered on top of the main X-Meta exchange. Partners earn futures-trading commissions from their referred users.

Written in Go (Gin + GORM), backed by PostgreSQL, fronted by xmeta-partner-web. Plugs into xmeta-monorepo via internal events for trade/deposit/registration hooks.

## Tech stack

| Layer | Choice |
|---|---|
| Language | Go 1.25 |
| HTTP | Gin |
| ORM | GORM (PostgreSQL driver) |
| Auth | AWS Cognito (separate pools for partner-portal vs admin) |
| API docs | swaggo/swag (generated to `docs/`) |
| Object storage | MinIO (S3-compatible) |
| Container | Docker + docker-compose |

## Project layout

```
.
├── cmd/futures-commission-sync/ # Daily cron — syncs futures closed positions into commissions
├── cmd/monthly-tier-review/ # Monthly cron — reviews previous-month tier assignment
├── controllers/            # HTTP handlers, grouped by audience
│   ├── admin/              # Admin back-office endpoints (Bearer + RBAC)
│   ├── partner/            # Partner self-service (Bearer)
│   ├── public/             # No-auth (referral click tracking, public tier list)
│   └── system/             # Internal events from xmeta-monorepo (X-Internal-API-Key)
├── services/               # Business logic, mirrors controllers/ layout
├── database/               # GORM models + migration helpers
├── middlewares/            # Auth, activity log, CORS, RBAC
├── structs/                # Request/response shapes (separate from DB models)
├── utils/                  # Cognito, AWS config, PII helpers, referral codes
├── docs/                   # Swagger output (regenerated on build)
├── nginx/                  # Reverse-proxy config for production
├── Dockerfile
├── docker-compose.yml
├── env.example             # Copy to .env and fill in
└── main.go
```

## Routes (high-level)

All routes are mounted under `/api/v1`.

### Public (`/api/v1/public/partner/*`)
- `GET /ref/:code` — track click, redirect to register page
- `GET /tiers` — list partner tiers for the marketing landing page

### Partner self-service (`/api/v1/partner/*`) — Bearer (partner Cognito)
- `auth/*` — status, apply, info, profile, tier
- `dashboard/*` — summary, charts, tier progress
- `referrals/*` — list (PII-masked), detail, stats
- `links/*` — list, create (max 5 codes per partner)
- `commissions/*` — list, breakdown, daily summary
- `payouts/*` — list, detail, pending balance

### Admin back-office (`/api/v1/admin/partner/*`) — Bearer (admin Cognito) + RBAC
- `applications/*` — list, detail, approve, reject
- `partners/*` — list, detail, update tier, update status
- `config/tiers/*` — CRUD on tiers
- `payouts/*` — list, pending, approve, reject
- `analytics/*` — summary, commission trend, top partners, referral funnel

### Internal (`/api/v1/internal/*`) — X-Internal-API-Key
Called by xmeta-monorepo when domain events fire.
- Futures commissions are synced from `futures_closed_positions` by the daily
  `futures-commission-sync` job.
- `GET /referral-links/:code` — preview a code (sender identity + active flag)
- `POST /link-referral` — attach a user to a partner's code (registration or settings flow)
- `POST /unlink-referral` — detach a user (account closure / compliance)
- `POST /user-deposited` — record first deposit on the active referral

Full Swagger UI: `http://<host>:<port>/swagger/index.html` after the server is up.

## Local setup

### Prerequisites
- Go 1.25+
- PostgreSQL 14+
- Docker (optional, for the compose setup)
- AWS Cognito user pools (one for partners, one for admins)

### 1. Clone & install
```bash
git clone git@github.com-work:x-meta-2/xmeta-partner-api.git
cd xmeta-partner-api
go mod download
```

### 2. Environment
```bash
cp env.example .env
# fill in DB_*, COGNITO_*, AWS_*, INTERNAL_API_KEY, MINIO_*, ALLOWED_ORIGINS
```

### 3. Database
```bash
# Option A — local Postgres
createdb xmeta_partner

# Option B — docker-compose (brings up Postgres + MinIO)
docker-compose up -d postgres minio
```

Set `DB_AUTO_MIGRATE=true` and `DB_RUN_MIGRATIONS=true` on first boot to create tables and seed the default tier.

### 4. Run
```bash
go run main.go
# → listens on APP_PORT (default 8080)
```

Hot reload during development:
```bash
go install github.com/cosmtrek/air@latest
air
```

### 5. Regenerate Swagger
```bash
~/go/bin/swag init --parseDependency --parseInternal -o ./docs
```
Run this after any change to controller annotations or struct shapes that surface in the API.

## Migrations

`database/migrations.go` runs idempotent SQL on startup (when `DB_RUN_MIGRATIONS=true`) for things AutoMigrate can't safely do — drop columns, partial unique indexes, backfill data. Each migration is keyed by date; once it has been run everywhere, the call is left in place but commented as a changelog entry.

To add a new migration:
1. Add a function in `database/migrations.go` (use `addColumnIfMissing`, `dropColumn`, `dropIndex` helpers)
2. Call it from `RunMigrations`
3. Boot once, verify, deploy
4. Once rolled out everywhere, comment the call (keep the helper as record)

## Futures commission sync

`cmd/futures-commission-sync/main.go` is a separate one-off binary that reads `futures_closed_positions` from the shared admin database and creates pending partner commission rows.

Commission formula:

```text
fee = close_notional * FUTURES_FEE_RATE
partner_rebate = fee * partner_tier.commission_rate
```

Both calculated amounts are rounded down to 4 decimal places before being
stored.

Tier rules:

- Real-time upgrade uses the current Asia/Ulaanbaatar calendar month's
  commission metrics only.
- Monthly downgrade/reassignment runs on day 1 using the previous calendar
  month's metrics.
- Monthly active clients are counted as distinct referred users with at least
  one commission in that month.

Default production command:

```bash
docker compose --profile jobs run --rm futures-commission-sync
```

The default run scans the last `FUTURES_SYNC_LOOKBACK_DAYS` closed days and skips duplicates via the unique `commissions.position_id` key.

Manual backfill:

```bash
FUTURES_SYNC_STARTED_AT=2026-09-01 FUTURES_SYNC_ENDED_AT=2026-09-24 \
  docker compose --profile jobs run --rm futures-commission-sync
```

Install the cron jobs:

```bash
./scripts/install-futures-sync-cron.sh
```

The script installs a daily futures commission sync at `08:00 UTC+8` and a
monthly tier review on day 1 at `08:30 UTC+8`.

## Internal events contract

xmeta-monorepo posts to `/api/v1/internal/*` for referral lookups, referral linking, and unlink requests. All requests must carry header:

```
X-Internal-API-Key: <INTERNAL_API_KEY from .env>
```

Example — registering a referral:
```bash
curl -X POST http://partner-api:8080/api/v1/internal/link-referral \
  -H "X-Internal-API-Key: $INTERNAL_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"userId":"abc-123","referralCode":"AB12CDE"}'
```

## Build & deploy

### Docker image
```bash
docker build -t xmeta-partner-api:latest .
```

### docker-compose (single host)
```bash
docker-compose up -d
```
Brings up `app`, `postgres`, `minio`, `nginx`. Nginx terminates TLS and reverse-proxies to `app:8080`.

### Production checklist
- [ ] Rotate `INTERNAL_API_KEY` from the env.example placeholder
- [ ] Use managed Postgres (RDS / CloudSQL) — not the docker-compose Postgres
- [ ] `DB_AUTO_MIGRATE=false` on prod (apply migrations explicitly)
- [ ] `ALLOWED_ORIGINS` whitelisted to actual partner-portal + admin domains
- [ ] Cognito pools configured with the right callback URLs
- [ ] Futures commission sync scheduled at 08:00 UTC+8
- [ ] Daily futures commission sync and monthly tier review cron jobs installed
- [ ] Logs shipped to CloudWatch / Sentry to catch `log.Printf` errors

## Conventions

- **DB models** live in `database/`; **wire types** (request/response) live in `structs/`. Never expose `database.*` directly through partner-facing endpoints — wrap in a sanitized DTO (see `partner.ReferralListItem` for the PII-masking pattern).
- **One service per controller folder.** Cross-team helpers go in `services/` root.
- **Status strings are constants** (`database.PartnerStatusActive`, `database.ReferralStatusUnlinked`, etc.) — never hardcode `"active"` in business logic.
- **Migrations are idempotent.** Use the helpers in `database/migrations.go`; assume a function may run on every boot.

## Related repos

- **xmeta-partner-web** — partner-facing dashboard (TanStack Start)
- **xmeta-admin** — back-office console (React + TanStack Router)
- **xmeta-monorepo** — main exchange backend; sends internal events here

## License

Internal — X-Meta proprietary.
