# How deployment gets environment variables

Env vars reach Cloud Run when you deploy with `deploy.sh`. The script builds a YAML env file and passes it with `--env-vars-file`. Cloud Build (`cloudbuild.yaml`) **does not** set env vars; it only updates the image.

## Primary source: `.env.production` per app

- **Laravel:** `niobe/.env.production` — copy from `niobe/.env.production.example`, fill in (e.g. `APP_KEY`, `APP_URL`, `GEMINI_API_KEY`).
- **Agent:** `agent/.env.production` — copy from `agent/.env.production.example`, fill in.

`deploy.sh` reads these files and merges them with **script overrides** (DB_*, Cloud SQL connection, and Cloud Run–specific vars). Overrides always win so DB connection is correct on Cloud Run.

## Flow

```
niobe/.env.production     ─┐
agent/.env.production     ─┼─► deploy.sh ─► temp YAML ─► gcloud run deploy --env-vars-file=...
Terraform (DB_*, conn)    ─┘       (overrides: DB_HOST=/cloudsql/..., APP_KEY, etc.)
```

- **With `.env.production`:** All vars from that file are sent to Cloud Run, except those overridden by the script (DB_*, APP_KEY, SESSION_DRIVER, etc.).
- **Without `.env.production`:** Only the script overrides are sent (DB + minimal Laravel/Agent vars). Set `APP_KEY` in shell: `export APP_KEY=base64:...` so the script can inject it.
- **Cloud Build:** Only updates the image; env vars stay as last set by `deploy.sh`.

## Overrides (always set by the script)

**Laravel:** `DB_CONNECTION`, `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD`, `APP_KEY` (from shell if not in .env.production), `SESSION_DRIVER`, `CACHE_STORE`, `QUEUE_CONNECTION`, `LOG_STACK`.

**Agent:** `DB_HOST`, `DB_PORT`, `DB_DATABASE`, `DB_USERNAME`, `DB_PASSWORD`.

So you can put everything else (e.g. `APP_URL`, `GEMINI_API_KEY`, `VOICE_AGENT_URL`, `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `AGENT_SHARED_SECRET`) in the app’s `.env.production` and they’ll be used as-is.

## Setting up

1. **Per-app `.env.production`** (recommended):

   ```bash
   cp niobe/.env.production.example niobe/.env.production
   cp agent/.env.production.example agent/.env.production
   # Edit both; add APP_KEY, URLs, API keys. Do not commit (they're gitignored).
   ./deploy/scripts/deploy.sh
   ```

2. **No file:** Export at least `APP_KEY` before deploy so the script can inject it: `export APP_KEY=base64:...`.

3. **GCP Console:** Cloud Run → service → Edit → Variables & secrets. Next `deploy.sh` will replace env vars from the YAML it builds; for secrets-only tweaks use Secret Manager or Console.

## Summary

- **Deployment gets envs** from **`niobe/.env.production`** and **`agent/.env.production`**, merged with script overrides (DB + Cloud Run–specific), then passed via `--env-vars-file` to `gcloud run deploy`.
- **Cloud Build** does not change env vars.
- Add or change vars in the app’s `.env.production` (or in the script’s override list if you need to force a value).
