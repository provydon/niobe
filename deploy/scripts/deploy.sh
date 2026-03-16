#!/usr/bin/env bash
# Build and deploy Laravel and/or Agent. Run from repo root.
# Usage: deploy.sh [laravel|agent|all]   (default: all)
# Env: from niobe/.env.production and agent/.env.production if present; script overrides DB_* and Cloud Run vars.
# Requires: docker, gcloud, Terraform applied (for repo + Cloud SQL).

set -euo pipefail

# Write YAML env file: .env file + overrides (override keys replace any from .env). Keys in OVERRIDE_KEYS are skipped when reading .env.
write_env_yaml() {
  local env_file="$1"
  local yaml_file="$2"
  local override_keys="$3"  # space-separated list of keys to override
  shift 3
  local override_pairs=("$@")  # KEY=VAL KEY2=VAL2 ...

  > "$yaml_file"
  if [[ -f "$env_file" ]]; then
    while IFS= read -r line || [[ -n "$line" ]]; do
      line="${line%%#*}"
      line="${line%"${line##*[![:space:]]}"}"
      [[ -z "$line" ]] && continue
      if [[ "$line" == *=* ]]; then
        key="${line%%=*}"
        key="${key%"${key##*[![:space:]]}"}"
        # Skip if this key will be overridden
        if [[ " $override_keys " == *" $key "* ]]; then
          continue
        fi
        val="${line#*=}"
        val="${val#"${val%%[![:space:]]*}"}"
        val="${val%\"}"
        val="${val#\"}"
        val="${val//\\/\\\\}"
        val="${val//\"/\\\"}"
        echo "${key}: \"${val}\"" >> "$yaml_file"
      fi
    done < "$env_file"
  fi
  for pair in "${override_pairs[@]}"; do
    if [[ "$pair" == *=* ]]; then
      key="${pair%%=*}"
      val="${pair#*=}"
      val="${val//\\/\\\\}"
      val="${val//\"/\\\"}"
      echo "${key}: \"${val}\"" >> "$yaml_file"
    fi
  done
}

TARGET="${1:-all}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

PROJECT_ID="${GCP_PROJECT_ID:-}"
REGION="${GCP_REGION:-us-central1}"
REPO_NAME="${ARTIFACT_REPO:-niobe}"

if [[ -z "$PROJECT_ID" ]]; then
  echo "Set GCP_PROJECT_ID"
  exit 1
fi

case "$TARGET" in
  laravel|agent|all) ;;
  *)
    echo "Usage: $0 [laravel|agent|all]"
    exit 1
    ;;
esac

IMAGE_BASE="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPO_NAME}"
LARAVEL_IMAGE="${IMAGE_BASE}/laravel:latest"
AGENT_IMAGE="${IMAGE_BASE}/agent:latest"

gcloud auth configure-docker "${REGION}-docker.pkg.dev" --quiet

# Cloud SQL (needed for deploy of either service)
CONN_NAME="${CLOUD_SQL_CONNECTION_NAME:-}"
if [[ -z "$CONN_NAME" ]]; then
  CONN_NAME=$(cd deploy/terraform && terraform output -raw cloud_sql_connection_name 2>/dev/null || true)
fi
if [[ -z "$CONN_NAME" ]]; then
  echo "Set CLOUD_SQL_CONNECTION_NAME or run Terraform and use its output."
  exit 1
fi

DB_USER="${DB_USERNAME:-niobe}"
DB_PASS="${DB_PASSWORD:-}"
DB_NAME="${DB_DATABASE:-niobe}"
if [[ -z "$DB_PASS" ]] && command -v terraform &>/dev/null; then
  DB_PASS=$(cd deploy/terraform && terraform output -raw db_password 2>/dev/null || true)
fi

if [[ "$TARGET" == "laravel" || "$TARGET" == "all" ]]; then
  echo "Building Laravel image..."
  docker build -f niobe/docker/Dockerfile --build-arg DEPLOYMENT_TYPE=web -t "$LARAVEL_IMAGE" niobe
  echo "Pushing Laravel image..."
  docker push "$LARAVEL_IMAGE"

  # Auto-detect Cloud Run URLs so Laravel gets correct APP_URL and VOICE_AGENT_URL (overrides .env.production)
  APP_URL_AUTO=$(gcloud run services describe niobe-web --region="$REGION" --project="$PROJECT_ID" --format='value(status.url)' 2>/dev/null || true)
  VOICE_AGENT_URL_AUTO=$(gcloud run services describe niobe-agent --region="$REGION" --project="$PROJECT_ID" --format='value(status.url)' 2>/dev/null || true)

  LARAVEL_OVERRIDE_KEYS="DB_CONNECTION DB_HOST DB_PORT DB_DATABASE DB_USERNAME DB_PASSWORD APP_KEY APP_URL VOICE_AGENT_URL SESSION_DRIVER CACHE_STORE QUEUE_CONNECTION LOG_STACK"
  LARAVEL_OVERRIDES=(
    "DB_CONNECTION=pgsql"
    "DB_HOST=/cloudsql/$CONN_NAME"
    "DB_PORT=5432"
    "DB_DATABASE=$DB_NAME"
    "DB_USERNAME=$DB_USER"
    "DB_PASSWORD=$DB_PASS"
    "SESSION_DRIVER=database"
    "CACHE_STORE=database"
    "QUEUE_CONNECTION=database"
    "LOG_STACK=single,stderr"
  )
  [[ -n "${APP_KEY:-}" ]] && LARAVEL_OVERRIDES+=("APP_KEY=$APP_KEY")
  [[ -n "${APP_URL_AUTO:-}" ]] && LARAVEL_OVERRIDES+=("APP_URL=$APP_URL_AUTO")
  [[ -n "${VOICE_AGENT_URL_AUTO:-}" ]] && LARAVEL_OVERRIDES+=("VOICE_AGENT_URL=$VOICE_AGENT_URL_AUTO")
  LARAVEL_ENV_YAML="${REPO_ROOT}/deploy/.env.laravel.run.yaml"
  write_env_yaml "${REPO_ROOT}/niobe/.env.production" "$LARAVEL_ENV_YAML" "$LARAVEL_OVERRIDE_KEYS" "${LARAVEL_OVERRIDES[@]}"

  echo "Deploying Laravel to Cloud Run..."
  gcloud run deploy niobe-web \
    --image="$LARAVEL_IMAGE" \
    --region="$REGION" \
    --platform=managed \
    --allow-unauthenticated \
    --add-cloudsql-instances="$CONN_NAME" \
    --env-vars-file="$LARAVEL_ENV_YAML" \
    --project="$PROJECT_ID"
  rm -f "$LARAVEL_ENV_YAML"
fi

if [[ "$TARGET" == "agent" || "$TARGET" == "all" ]]; then
  echo "Building Agent image..."
  docker build -f agent/Dockerfile -t "$AGENT_IMAGE" agent
  echo "Pushing Agent image..."
  docker push "$AGENT_IMAGE"

  # Agent uses a single DATABASE_URL (PostgreSQL URL; Cloud SQL socket via host param)
  AGENT_DATABASE_URL="postgres://${DB_USER}:${DB_PASS}@/${DB_NAME}?host=/cloudsql/${CONN_NAME}"
  AGENT_OVERRIDE_KEYS="DATABASE_URL DB_URL"
  AGENT_OVERRIDES=(
    "DATABASE_URL=$AGENT_DATABASE_URL"
  )
  AGENT_ENV_YAML="${REPO_ROOT}/deploy/.env.agent.run.yaml"
  write_env_yaml "${REPO_ROOT}/agent/.env.production" "$AGENT_ENV_YAML" "$AGENT_OVERRIDE_KEYS" "${AGENT_OVERRIDES[@]}"

  echo "Deploying Agent to Cloud Run..."
  gcloud run deploy niobe-agent \
    --image="$AGENT_IMAGE" \
    --region="$REGION" \
    --platform=managed \
    --allow-unauthenticated \
    --add-cloudsql-instances="$CONN_NAME" \
    --env-vars-file="$AGENT_ENV_YAML" \
    --project="$PROJECT_ID"
  rm -f "$AGENT_ENV_YAML"
fi

echo "Done."
if [[ "$TARGET" == "laravel" || "$TARGET" == "all" ]]; then
  gcloud run services describe niobe-web --region="$REGION" --format='value(status.url)' --project="$PROJECT_ID"
fi
if [[ "$TARGET" == "agent" || "$TARGET" == "all" ]]; then
  gcloud run services describe niobe-agent --region="$REGION" --format='value(status.url)' --project="$PROJECT_ID"
fi
