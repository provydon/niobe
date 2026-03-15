#!/usr/bin/env bash
# Enable GCP APIs and ensure Artifact Registry exists.
# Run from repo root. Terraform creates the repo; this script enables APIs if you skip Terraform initially.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
cd "$REPO_ROOT"

PROJECT_ID="${GCP_PROJECT_ID:-}"
REGION="${GCP_REGION:-us-central1}"

if [[ -z "$PROJECT_ID" ]]; then
  echo "Set GCP_PROJECT_ID (e.g. export GCP_PROJECT_ID=my-project)"
  exit 1
fi

echo "Enabling required APIs for project $PROJECT_ID..."
gcloud services enable \
  run.googleapis.com \
  sqladmin.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com \
  servicenetworking.googleapis.com \
  compute.googleapis.com \
  secretmanager.googleapis.com \
  --project="$PROJECT_ID"

echo "APIs enabled. Create Artifact Registry and Cloud SQL via Terraform:"
echo "  cd deploy/terraform && terraform init && terraform apply -var=project_id=$PROJECT_ID -var=region=$REGION"
echo "Then run ./deploy/scripts/deploy.sh to build and deploy the apps."
