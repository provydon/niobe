#!/usr/bin/env bash
# Build and deploy Laravel and/or Agent. Run from repo root.
# Usage: deploy.sh [laravel|agent|all]   (default: all)
# Requires: docker, gcloud, Terraform applied (for repo + Cloud SQL).

set -euo pipefail

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

  LARAVEL_ENV="DB_CONNECTION=pgsql,DB_HOST=/cloudsql/$CONN_NAME,DB_PORT=5432,DB_DATABASE=$DB_NAME,DB_USERNAME=$DB_USER,DB_PASSWORD=$DB_PASS,APP_KEY=${APP_KEY:-},SESSION_DRIVER=database,CACHE_STORE=database,QUEUE_CONNECTION=database,LOG_STACK=single,stderr"
  echo "Deploying Laravel to Cloud Run..."
  gcloud run deploy niobe-web \
    --image="$LARAVEL_IMAGE" \
    --region="$REGION" \
    --platform=managed \
    --allow-unauthenticated \
    --add-cloudsql-instances="$CONN_NAME" \
    --set-env-vars="$LARAVEL_ENV" \
    --project="$PROJECT_ID"
fi

if [[ "$TARGET" == "agent" || "$TARGET" == "all" ]]; then
  echo "Building Agent image..."
  docker build -f agent/Dockerfile -t "$AGENT_IMAGE" agent
  echo "Pushing Agent image..."
  docker push "$AGENT_IMAGE"

  AGENT_ENV="DB_HOST=/cloudsql/$CONN_NAME,DB_PORT=5432,DB_DATABASE=$DB_NAME,DB_USERNAME=$DB_USER,DB_PASSWORD=$DB_PASS"
  echo "Deploying Agent to Cloud Run..."
  gcloud run deploy niobe-agent \
    --image="$AGENT_IMAGE" \
    --region="$REGION" \
    --platform=managed \
    --allow-unauthenticated \
    --add-cloudsql-instances="$CONN_NAME" \
    --set-env-vars="$AGENT_ENV" \
    --project="$PROJECT_ID"
fi

echo "Done."
if [[ "$TARGET" == "laravel" || "$TARGET" == "all" ]]; then
  gcloud run services describe niobe-web --region="$REGION" --format='value(status.url)' --project="$PROJECT_ID"
fi
if [[ "$TARGET" == "agent" || "$TARGET" == "all" ]]; then
  gcloud run services describe niobe-agent --region="$REGION" --format='value(status.url)' --project="$PROJECT_ID"
fi
