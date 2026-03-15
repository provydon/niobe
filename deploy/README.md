# Cloud deployment automation (Hackathon bonus)

This folder contains **automated Cloud Deployment** using **scripts** and **infrastructure-as-code** so it can be included in our public repository for the hackathon bonus.

## What's included

| Artifact | Type | Purpose |
|----------|------|---------|
| **`terraform/`** | Infrastructure as Code | Provisions GCP: APIs, Artifact Registry, Cloud SQL (PostgreSQL), VPC, Cloud Run services |
| **`scripts/setup.sh`** | Script | Enables required APIs and creates Artifact Registry repo |
| **`scripts/deploy.sh`** | Script | Builds images, pushes to Artifact Registry, deploys to Cloud Run |
| **`scripts/teardown.sh`** | Script | Destroys Terraform state and optionally Cloud SQL (cleanup) |
| **`../cloudbuild.yaml`** | IaC (YAML) | Full pipeline: build both images and deploy (manual or one trigger) |
| **`../cloudbuild-laravel.yaml`** | IaC (YAML) | Build and deploy **only Laravel** (for path-based CI/CD) |
| **`../cloudbuild-agent.yaml`** | IaC (YAML) | Build and deploy **only Agent** (for path-based CI/CD) |

## Prerequisites

- [Google Cloud SDK (gcloud)](https://cloud.google.com/sdk/docs/install) installed and logged in (`gcloud auth login`)
- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.0
- A GCP project and billing enabled

## Quick start

1. **Set variables** (or export env vars):

   ```bash
   export GCP_PROJECT_ID=your-gcp-project-id
   export GCP_REGION=us-central1
   ```

2. **Bootstrap (one-time)** – enable APIs and create Artifact Registry:

   ```bash
   ./scripts/setup.sh
   ```

3. **Provision infrastructure** with Terraform:

   ```bash
   cd terraform
   terraform init
   terraform plan -var="project_id=$GCP_PROJECT_ID" -var="region=$GCP_REGION"
   terraform apply -var="project_id=$GCP_PROJECT_ID" -var="region=$GCP_REGION"
   cd ..
   ```

4. **Deploy applications** (build + push + deploy):

   ```bash
   ./scripts/deploy.sh          # both Laravel and Agent
   ./scripts/deploy.sh laravel  # only Laravel
   ./scripts/deploy.sh agent    # only Agent
   ./scripts/deploy.sh all      # same as no argument
   ```

   Or trigger via Cloud Build (full deploy):

   ```bash
   gcloud builds submit --config=cloudbuild.yaml .
   ```

## Monorepo: deploy individually and path-based CI/CD

In a monorepo you can deploy **only what changed** and run **separate CI/CD pipelines** per app.

### Deploy from your machine (per app)

```bash
./deploy/scripts/deploy.sh laravel   # build + push + deploy only Laravel
./deploy/scripts/deploy.sh agent     # build + push + deploy only Agent
```

### CI/CD: path-based Cloud Build triggers

Create **two triggers** so that pushes under `niobe/**` only build/deploy Laravel, and pushes under `agent/**` only build/deploy the Agent.

1. **Cloud Build → Triggers → Create trigger**

2. **Trigger 1 – Laravel**
   - Name: `deploy-laravel`
   - Event: Push to branch (e.g. `^main$`)
   - **Included files filter:** `niobe/**, deploy/**, cloudbuild-laravel.yaml`
   - **Config:** `cloudbuild-laravel.yaml` (repo root)
   - So: when you change only `niobe/` or `deploy/`, only Laravel is built and deployed.

3. **Trigger 2 – Agent**
   - Name: `deploy-agent`
   - Event: Push to branch (e.g. `^main$`)
   - **Included files filter:** `agent/**, cloudbuild-agent.yaml`
   - **Config:** `cloudbuild-agent.yaml` (repo root)
   - So: when you change only `agent/`, only the Agent is built and deployed.

4. **Optional – full deploy trigger**
   - Name: `deploy-all`
   - Included files: e.g. `cloudbuild.yaml` or leave broad
   - Config: `cloudbuild.yaml`
   - Use when you want to redeploy both on any change (or run manually).

With this setup, the monorepo still has one repo, but each app is **deployed individually** and CI/CD runs only the pipeline for the code that changed.

## Teardown

To remove deployed apps and optionally the database:

```bash
 ./scripts/teardown.sh
```

Then in `terraform/`: `terraform destroy -var="project_id=$GCP_PROJECT_ID" -var="region=$GCP_REGION"` (when you want to remove all infra).

## Repository note

This code is intended to be **included in our public repository** as proof of automated cloud deployment for the hackathon bonus (scripts + infrastructure-as-code).
