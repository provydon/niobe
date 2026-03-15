#!/usr/bin/env bash
# Remove Cloud Run services and optionally clean up. Run from repo root.

set -euo pipefail

PROJECT_ID="${GCP_PROJECT_ID:-}"
REGION="${GCP_REGION:-us-central1}"

if [[ -z "$PROJECT_ID" ]]; then
  echo "Set GCP_PROJECT_ID"
  exit 1
fi

echo "Deleting Cloud Run services..."
gcloud run services delete niobe-web --region="$REGION" --project="$PROJECT_ID" --quiet 2>/dev/null || true
gcloud run services delete niobe-agent --region="$REGION" --project="$PROJECT_ID" --quiet 2>/dev/null || true

echo "Teardown complete. To remove database and all infra, run:"
echo "  cd deploy/terraform && terraform destroy -var=project_id=$PROJECT_ID -var=region=$REGION"
