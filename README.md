# Niobe

**AI waitress for restaurants.** Upload your menu, share one link. Customers talk to your waitress, have a conversation and order by voice.

## Repository structure

| Path | Description |
|------|-------------|
| **`niobe/`** | Laravel + Vue 3 + Inertia web app: dashboard, waitress management, menu upload, AI menu extraction (Gemini). |
| **`agent/`** | Go voice-agent service: live voice sessions, menu-aware ordering, connects to Laravel/DB. |
| **`deploy/`** | Deployment automation: Terraform (GCP), scripts, Cloud Build configs. See [deploy/README.md](deploy/README.md). |

## Architecture

How **Gemini** connects to the backend, database, and frontend:

```
┌─────────────┐     HTTPS      ┌──────────────────┐     Gemini API      ┌─────────────┐
│   Browser   │ ──────────────►│  Niobe (Laravel) │ ──────────────────► │   Gemini    │
│  (Vue/Inertia)               │  + menu extract  │                     │ (menu from  │
└──────┬──────┘                └────────┬─────────┘                     │  images)    │
       │                                │                                └─────────────┘
       │ WebSocket /live?niobe=slug     │
       ▼                                ▼
┌──────────────────┐            ┌──────────────┐
│  Voice Agent     │◄──────────►│  PostgreSQL  │
│  (Go)            │   read/    └──────────────┘
│  Proxy ◄───────► │   write
│  Gemini Live     │
└────────┬─────────┘
         │ real-time audio + tool calls
         ▼
┌─────────────────┐
│  Gemini Live    │
│  (voice model)  │
└─────────────────┘
```

- **Laravel** uses **Gemini API** for menu extraction from uploaded images; data is stored in **PostgreSQL**.
- **Agent** uses **Gemini Live API** for voice; it reads waitress/menu from **PostgreSQL** and runs tools (orders, email, webhooks) in-process, writing back to the same DB.

Full diagram and data flows: [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Prerequisites

- **Niobe (web):** PHP ≥ 8.2, Composer, Node.js 22+, npm. For tests: in-memory SQLite (no DB setup).
- **Agent:** Go 1.24+.
- **Deploy:** See [deploy/README.md](deploy/README.md) (gcloud, Terraform, GCP project).

---

## Reproducible testing

These steps give the same test results as CI and can be run locally without external services (no real DB or API keys for the test suites).

### 1. Niobe (Laravel) tests

Tests use **Pest** with PHPUnit, in-memory SQLite, and the config in `niobe/phpunit.xml`. No database or `.env` secrets required for the test run.

**Requirements:** PHP 8.4 or 8.5 (matches CI), Composer, Node 22.

**Steps (from repo root):**

```bash
cd niobe

# Install PHP and JS dependencies
composer install --no-interaction --prefer-dist --optimize-autoloader
npm ci

# Optional: copy env and generate key (needed for artisan, not for test DB – tests use :memory:)
cp .env.example .env
php artisan key:generate

# Run the test suite (same as CI)
./vendor/bin/pest
```

**What this runs:**

- **Unit tests:** `niobe/tests/Unit/`
- **Feature tests:** `niobe/tests/Feature/` (auth, dashboard, waitress management, public Niobe pages, etc.)

Tests use `APP_ENV=testing`, `DB_CONNECTION=sqlite`, `DB_DATABASE=:memory:` (set in `phpunit.xml`), so no database setup or migrations are required for a reproducible run.

**Optional (match CI exactly):**

```bash
npm run build
./vendor/bin/pest
```

### 2. Agent (Go) tests

Tests are in the `agent` package. No database or external APIs are required; tests are unit-style (e.g. proxy logic).

**Requirements:** Go 1.24 or later.

**Steps (from repo root):**

```bash
cd agent

# Download modules
go mod download

# Run all tests
go test ./...
```

**What this runs:**

- `agent/proxy/proxy_test.go` (and any other `*_test.go` under `agent/`).

To run with verbose output:

```bash
go test -v ./...
```

### 3. Run both (from repo root)

```bash
# Niobe
cd niobe && composer install --no-interaction --prefer-dist --optimize-autoloader && npm ci && cp .env.example .env && php artisan key:generate && ./vendor/bin/pest && cd ..

# Agent
cd agent && go test ./... && cd ..
```

Or run each `cd` block in a separate terminal.

### CI reference

- **Niobe:** [niobe/.github/workflows/tests.yml](niobe/.github/workflows/tests.yml) — PHP 8.4 & 8.5, Composer, Node 22, `cp .env.example .env`, `php artisan key:generate`, `npm run build`, `./vendor/bin/pest`.
- **Agent:** No GitHub Action in repo; run `go test ./...` in `agent/` for the same result locally.

---

## Local development (optional)

- **Niobe:** From `niobe/`, run `php artisan serve`, and in another terminal `npm run dev` for Vite. Use `.env` (e.g. `DB_CONNECTION=sqlite`, `DB_DATABASE=database/database.sqlite`) and run `php artisan migrate` if you use a real DB.
- **Agent:** From `agent/`, run `go run .` (requires DB and env; see `agent/.env.example` or deploy docs for production-style config).
- **Deploy:** See [deploy/README.md](deploy/README.md) and [deploy/DEPLOY-GCP.md](deploy/DEPLOY-GCP.md).

## License

MIT (or as specified in subdirectories).
