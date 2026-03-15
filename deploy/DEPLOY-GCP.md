# Deploy Niobe on Google Cloud

End-to-end steps to run Laravel (niobe) + Go agent + PostgreSQL on GCP (Cloud Run + Cloud SQL).

## Prerequisites

- **Google Cloud account** with billing enabled
- **gcloud** CLI: [install](https://cloud.google.com/sdk/docs/install) and run `gcloud auth login`
- **Terraform** >= 1.0: [install](https://developer.hashicorp.com/terraform/install)
- **Docker** (for local build; or use Cloud Build only)

## 1. Create or pick a GCP project

```bash
# Create a project (or use existing)
gcloud projects create YOUR_PROJECT_ID --name="Niobe"
gcloud config set project YOUR_PROJECT_ID

# Enable billing for the project in Cloud Console: Billing → Link a billing account
```

## 2. Set environment variables

**Option A – use `.env.production` per app (recommended)**

```bash
cp niobe/.env.production.example niobe/.env.production
cp agent/.env.production.example agent/.env.production
# Edit both: set APP_KEY, APP_URL, GEMINI_API_KEY, VOICE_AGENT_URL, etc. Do not commit (gitignored).
```

**Option B – shell only**

From the **repo root**:

```bash
export GCP_PROJECT_ID=YOUR_PROJECT_ID
export GCP_REGION=us-central1
export APP_KEY=base64:your-laravel-app-key   # required if not in niobe/.env.production
```

Deploy script **overrides** DB_* and Cloud Run–specific vars; everything else can live in `niobe/.env.production` and `agent/.env.production`. See **`deploy/ENV-VARS.md`**.

Generate a Laravel key if needed:

```bash
cd niobe && php artisan key:generate --show && cd ..
```

## 3. Enable APIs

```bash
./deploy/scripts/setup.sh
```

## 4. Provision infrastructure (Terraform)

Creates: Artifact Registry repo, Cloud SQL PostgreSQL instance, database, and user.

```bash
cd deploy/terraform
terraform init
terraform plan -var="project_id=$GCP_PROJECT_ID" -var="region=$GCP_REGION"
terraform apply -var="project_id=$GCP_PROJECT_ID" -var="region=$GCP_REGION"
```

Type `yes` when prompted. Save the outputs (or re-run `terraform output` later).

```bash
cd ../..
```

## 5. Deploy the apps

Builds Docker images, pushes to Artifact Registry, deploys to Cloud Run (with Cloud SQL connection).

```bash
./deploy/scripts/deploy.sh
```

First run can take several minutes (Docker build + push). You’ll get two URLs at the end:

- **Laravel (niobe-web)** – main app
- **Agent (niobe-agent)** – Go API (e.g. `/health`, `/live`)

## 6. (Optional) Allow Cloud Run to reach Cloud SQL

If you used Terraform’s Cloud SQL with **public IP**, ensure the Cloud Run service account can connect. If you see connection errors:

1. In **Cloud Console → IAM**: find the Cloud Run service account  
   `PROJECT_NUMBER-compute@developer.gserviceaccount.com`
2. Grant it **Cloud SQL Client** (or add the role in Terraform).

Terraform does not set this by default; you can add it in `deploy/terraform/main.tf` or via Console.

## 7. Later: redeploy only one app

```bash
./deploy/scripts/deploy.sh laravel   # only Laravel
./deploy/scripts/deploy.sh agent    # only Agent
```

## Alternative: deploy via Cloud Build (no local Docker)

From repo root, after Terraform has been applied at least once (so Cloud Run services and Cloud SQL exist):

```bash
gcloud builds submit --config=cloudbuild.yaml .
```

This builds and deploys both services in the cloud. For the **first** deploy you still need to run `deploy.sh` once so Cloud Run gets the Cloud SQL connection and env vars; after that, `gcloud builds submit` updates the images only.

## Teardown

Remove apps:

```bash
./deploy/scripts/teardown.sh
```

Remove all infrastructure (DB, registry, etc.):

```bash
cd deploy/terraform
terraform destroy -var="project_id=$GCP_PROJECT_ID" -var="region=$GCP_REGION"
```
