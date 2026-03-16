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

First run can take several minutes (Docker build + push). You’ll get URLs for the web and agent (the worker pool has no HTTP endpoint):

- **Laravel web (niobe-web)** – main app
- **Laravel worker (niobe-worker)** – Cloud Run worker pool, same image with `DEPLOYMENT_TYPE=worker` (queue:work + schedule:work)
- **Agent (niobe-agent)** – Go API (e.g. `/health`, `/live`)

## 6. (Optional) Custom domains (e.g. niobe.live, agent.niobe.live)

To serve the web app and agent on your own domain:

1. **Verify domain ownership** in Google Cloud (required before mapping):
   - In **Cloud Console** go to **Cloud Run** → **Domain mappings** (or **APIs & Services** → **Domain verification**).
   - Add and verify **niobe.live** and **agent.niobe.live** (e.g. via DNS TXT record or HTML file as instructed).

2. **Apply Terraform** with the custom domain variables (services **niobe-web** and **niobe-agent** must already exist from a prior deploy):
   ```bash
   cd deploy/terraform
   terraform apply \
     -var="project_id=$GCP_PROJECT_ID" \
     -var="region=$GCP_REGION" \
     -var='github_repo_uri=https://github.com/provydon/niobe' \
     -var='custom_domain_web=niobe.live' \
     -var='custom_domain_agent=agent.niobe.live'
   ```

3. **Configure DNS** at your registrar for **niobe.live** and **agent.niobe.live**:
   - After the domain mapping is created, Cloud Run shows the required **CNAME** (or **A**) target (e.g. `ghs.googlehosted.com` or a run.app redirect).
   - In **Cloud Run** → **Domain mappings**, open each mapping and follow the **DNS records** instructions.
   - Point each domain (and optionally `www.niobe.live`) to the given target.

4. **Set APP_URL / VOICE_AGENT_URL** so Laravel and the agent use the custom URLs: in Secret Manager (**niobe-app-env** / **agent-app-env**) or in your next deploy, set `APP_URL=https://niobe.live` and ensure the Laravel app and agent know the agent URL (e.g. `VOICE_AGENT_URL=https://agent.niobe.live`). Re-deploy or update env so the apps use these URLs.

To skip custom domains, omit the `custom_domain_web` and `custom_domain_agent` variables (default is empty).

## 7. (Optional) Allow Cloud Run to reach Cloud SQL

If you used Terraform’s Cloud SQL with **public IP**, ensure the Cloud Run service account can connect. If you see connection errors:

1. In **Cloud Console → IAM**: find the Cloud Run service account  
   `PROJECT_NUMBER-compute@developer.gserviceaccount.com`
2. Grant it **Cloud SQL Client** (or add the role in Terraform).

Terraform does not set this by default; you can add it in `deploy/terraform/main.tf` or via Console.

## 8. Later: redeploy only one app

```bash
./deploy/scripts/deploy.sh laravel   # only Laravel
./deploy/scripts/deploy.sh agent    # only Agent
```

## Alternative: build in Cloud with env (Cloud Build) (no local Docker)

You can build **and** deploy with env vars entirely in GCP (no Docker on your machine). Terraform creates Secret Manager secrets for the DB password and Laravel APP_KEY; Cloud Build reads them and deploys with `--env-vars-file` and Cloud SQL.

**One-time after Terraform:**

1. **Re-apply Terraform** so the new secrets exist (if you applied before this was added):
   ```bash
   cd deploy/terraform && terraform apply -var="project_id=$GCP_PROJECT_ID" -var="region=$GCP_REGION" && cd ../..
   ```

2. **Set your Laravel APP_KEY** in Secret Manager (replace the placeholder):
   ```bash
   cd niobe && KEY=$(php artisan key:generate --show) && cd ../..
   echo -n "$KEY" | gcloud secrets versions add niobe-app-key --data-file=- --project=YOUR_PROJECT_ID
   ```

3. **(Optional) Full app env** – To inject **all** env vars (e.g. GEMINI_API_KEY, GOOGLE_CLIENT_ID, etc.):
   - **Secret Manager** → **niobe-app-env** → **New version** → paste your **Laravel** `.env.production` content (KEY=VALUE, one per line). Build will merge this with DB_* and Cloud Run overrides.
   - **agent-app-env** → **New version** → paste your **Agent** `.env.production` content. Same merge for the Go app.
   - If you only set **niobe-app-key** (step 2), Laravel still gets DB + APP_KEY; add **niobe-app-env** / **agent-app-env** when you want the rest of your env in the build.

**Deploy (from repo root):**

```bash
gcloud builds submit --config=cloudbuild-with-env.yaml .
```

This builds both images in Cloud Build, fetches DB password and APP_KEY from Secret Manager, writes env YAML, and deploys to Cloud Run: **niobe-web** (service), **niobe-worker** (worker pool, same Laravel image with `DEPLOYMENT_TYPE=worker`), and **niobe-agent** (Go), all with `--add-cloudsql-instances` and `--env-vars-file`. No local Docker required.

### Let Terraform create the GitHub trigger (optional)

After you’ve created the **connection** in **Cloud Build → Connect host** (OAuth with GitHub), Terraform can **link the repo** and create the trigger so builds run on push to `main`.

1. Apply Terraform with your GitHub repo URI (and connection name if it’s not `niobe`):
   ```bash
   cd deploy/terraform
   terraform apply -var="project_id=niobe-489920" -var="region=us-central1" \
     -var='github_repo_uri=https://github.com/YOUR_GITHUB_USERNAME/niobe'
   cd ../..
   ```
   Use your actual GitHub org/user and repo name. If your connection has a different name, add e.g. `-var='cloud_build_connection_name=your-connection-name'`.

2. If Terraform reports that the repository **already exists** (you linked it in the Console), import it:
   ```bash
   terraform import 'google_cloudbuildv2_repository.niobe[0]' projects/niobe-489920/locations/us-central1/connections/niobe/repositories/niobe
   ```
   Then run `terraform apply` again.

3. Pushes to `main` will run `cloudbuild-with-env.yaml`; you can also run the trigger manually from **Cloud Build → Triggers**.

**Other Cloud Build configs**

- **`cloudbuild.yaml`** – Builds and deploys in cloud but **does not** set env vars or Cloud SQL; use after a first deploy with `deploy.sh` or `cloudbuild-with-env.yaml` if you only want to update images.
- **`cloudbuild-with-env.yaml`** – Full deploy in cloud with env (recommended when you don’t run Docker locally).

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
