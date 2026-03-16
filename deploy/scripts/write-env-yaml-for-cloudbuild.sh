#!/usr/bin/env bash
# Write laravel-env.yaml and agent-env.yaml for Cloud Build (env from Secret Manager).
# Optional: LARAVEL_BASE_ENV_FILE and AGENT_BASE_ENV_FILE (.env format); script overrides DB_*, APP_KEY, etc.
# Usage: CONN_NAME=... DB_PASS=... [APP_KEY=...] [LARAVEL_BASE_ENV_FILE=path] [AGENT_BASE_ENV_FILE=path] write-env-yaml-for-cloudbuild.sh

set -euo pipefail

CONN_NAME="${CONN_NAME:?}"
DB_USER="${DB_USER:-niobe}"
DB_NAME="${DB_NAME:-niobe}"
DB_PASS="${DB_PASS:?}"
APP_KEY="${APP_KEY:-}"
APP_URL="${APP_URL:-}"
VOICE_AGENT_URL="${VOICE_AGENT_URL:-}"
LARAVEL_BASE_ENV_FILE="${LARAVEL_BASE_ENV_FILE:-}"
AGENT_BASE_ENV_FILE="${AGENT_BASE_ENV_FILE:-}"

yaml_quote() {
  local v="$1"
  v="${v//\\/\\\\}"
  v="${v//\"/\\\"}"
  printf '%s' "$v"
}

# Parse .env format into associative array (key=value, skip # and empty). Writes YAML to stdout; skip keys in override list.
env_to_yaml() {
  local env_file="$1"
  local override_keys="$2"
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%%#*}"
    line="${line%"${line##*[![:space:]]}"}"
    [[ -z "$line" ]] && continue
    if [[ "$line" == *=* ]]; then
      key="${line%%=*}"
      key="${key%"${key##*[![:space:]]}"}"
      [[ " $override_keys " == *" $key "* ]] && continue
      val="${line#*=}"
      val="${val#"${val%%[![:space:]]*}"}"
      val="${val%\"}"
      val="${val#\"}"
      echo "${key}: \"$(yaml_quote "$val")\""
    fi
  done < "$env_file"
}

# APP_KEY can come from base env; we add it from variable (secret) at end so secret overrides
LARAVEL_OVERRIDE_KEYS="DB_CONNECTION DB_HOST DB_PORT DB_DATABASE DB_USERNAME DB_PASSWORD APP_URL VOICE_AGENT_URL SESSION_DRIVER CACHE_STORE QUEUE_CONNECTION LOG_STACK"
# PORT is reserved by Cloud Run; never pass it in agent-env.yaml (Cloud Run sets PORT=8080)
# Agent uses same DB_* as Laravel (builds DSN from DB_HOST, DB_PORT, DB_DATABASE, DB_USERNAME, DB_PASSWORD)
AGENT_OVERRIDE_KEYS="DATABASE_URL DB_URL DB_CONNECTION DB_HOST DB_PORT DB_DATABASE DB_USERNAME DB_PASSWORD PORT"

# Laravel env: base file (if any) then overrides
> laravel-env.yaml
if [[ -n "$LARAVEL_BASE_ENV_FILE" && -f "$LARAVEL_BASE_ENV_FILE" ]]; then
  env_to_yaml "$LARAVEL_BASE_ENV_FILE" "$LARAVEL_OVERRIDE_KEYS" >> laravel-env.yaml
fi
cat >> laravel-env.yaml << EOF
DB_CONNECTION: "pgsql"
DB_HOST: "/cloudsql/$CONN_NAME"
DB_PORT: "5432"
DB_DATABASE: "$DB_NAME"
DB_USERNAME: "$DB_USER"
DB_PASSWORD: "$(yaml_quote "$DB_PASS")"
SESSION_DRIVER: "database"
CACHE_STORE: "database"
QUEUE_CONNECTION: "database"
LOG_STACK: "single,stderr"
EOF
[[ -n "$APP_KEY" ]] && echo "APP_KEY: \"$(yaml_quote "$APP_KEY")\"" >> laravel-env.yaml
[[ -n "$APP_URL" ]] && echo "APP_URL: \"$(yaml_quote "$APP_URL")\"" >> laravel-env.yaml
[[ -n "$VOICE_AGENT_URL" ]] && echo "VOICE_AGENT_URL: \"$(yaml_quote "$VOICE_AGENT_URL")\"" >> laravel-env.yaml

# Worker env: same as Laravel + DEPLOYMENT_TYPE=worker (for Cloud Run worker pool)
cp laravel-env.yaml worker-env.yaml
echo 'DEPLOYMENT_TYPE: "worker"' >> worker-env.yaml

# Agent env: base file (if any) then overrides. DB_* set exactly like Laravel so agent shares same DB config.
> agent-env.yaml
if [[ -n "$AGENT_BASE_ENV_FILE" && -f "$AGENT_BASE_ENV_FILE" ]]; then
  env_to_yaml "$AGENT_BASE_ENV_FILE" "$AGENT_OVERRIDE_KEYS" >> agent-env.yaml
fi
# Same DB block as Laravel (Cloud SQL socket); agent config builds DSN from these
cat >> agent-env.yaml << EOF
DB_CONNECTION: "pgsql"
DB_HOST: "/cloudsql/$CONN_NAME"
DB_PORT: "5432"
DB_DATABASE: "$DB_NAME"
DB_USERNAME: "$DB_USER"
DB_PASSWORD: "$(yaml_quote "$DB_PASS")"
EOF

echo "Wrote laravel-env.yaml, worker-env.yaml, and agent-env.yaml"
