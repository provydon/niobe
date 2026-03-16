terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Enable required APIs
resource "google_project_service" "apis" {
  for_each = toset([
    "run.googleapis.com",
    "sqladmin.googleapis.com",
    "artifactregistry.googleapis.com",
    "cloudbuild.googleapis.com",
    "servicenetworking.googleapis.com",
    "compute.googleapis.com",
    "secretmanager.googleapis.com",
  ])
  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

# Artifact Registry repository for Laravel and Agent images
resource "google_artifact_registry_repository" "repo" {
  location      = var.region
  repository_id = var.artifact_repo
  description   = "Container images for Niobe Laravel app and Go agent"
  format        = "DOCKER"

  depends_on = [google_project_service.apis]
}

# Cloud SQL PostgreSQL instance (public IP for simplicity; use Private IP + VPC for production)
resource "google_sql_database_instance" "main" {
  name             = "niobe-db"
  database_version = "POSTGRES_15"
  region           = var.region

  settings {
    tier = var.db_tier

    database_flags {
      name  = "max_connections"
      value = "100"
    }

    ip_configuration {
      ipv4_enabled    = true
      ssl_mode        = "ALLOW_UNENCRYPTED_AND_ENCRYPTED"
      private_network = null
    }

    backup_configuration {
      enabled                        = true
      start_time                     = "03:00"
      point_in_time_recovery_enabled = false
    }
  }

  deletion_protection = false

  depends_on = [google_project_service.apis]
}

resource "google_sql_database" "db" {
  name     = var.db_name
  instance = google_sql_database_instance.main.name
}

# DB user with generated password (store in Secret Manager in production)
resource "random_password" "db" {
  length  = 24
  special = true
}

resource "google_sql_user" "db_user" {
  name     = "niobe"
  instance = google_sql_database_instance.main.name
  password = random_password.db.result
}

# Allow Cloud Run (default compute SA) to connect to Cloud SQL
data "google_project" "project" {
  project_id = var.project_id
}

resource "google_project_iam_member" "cloudrun_sqlclient" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

# Secret Manager: DB password for Cloud Build (so cloudbuild-with-env.yaml can deploy with env)
resource "google_secret_manager_secret" "db_password" {
  secret_id = "niobe-db-password"
  replication {
    auto {}
  }
  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "db_password" {
  secret      = google_secret_manager_secret.db_password.id
  secret_data = random_password.db.result
}

# Laravel APP_KEY; replace the placeholder version with your key: echo -n "base64:YOUR_KEY" | gcloud secrets versions add niobe-app-key --data-file=-
resource "google_secret_manager_secret" "app_key" {
  secret_id = "niobe-app-key"
  replication {
    auto {}
  }
  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "app_key_placeholder" {
  secret      = google_secret_manager_secret.app_key.id
  secret_data = "replace-me-with-php-artisan-key-generate--show"
}

# Grant Cloud Build service account access to secrets (for cloudbuild-with-env.yaml)
resource "google_secret_manager_secret_iam_member" "cloudbuild_db_password" {
  secret_id = google_secret_manager_secret.db_password.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}@cloudbuild.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "cloudbuild_app_key" {
  secret_id = google_secret_manager_secret.app_key.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}@cloudbuild.gserviceaccount.com"
}

# Full app env (paste .env.production content); script merges with DB_* overrides
resource "google_secret_manager_secret" "niobe_app_env" {
  secret_id = "niobe-app-env"
  replication {
    auto {}
  }
  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "niobe_app_env_placeholder" {
  secret      = google_secret_manager_secret.niobe_app_env.id
  secret_data = "# Paste Laravel .env.production content (KEY=VALUE). DB_* and APP_KEY overridden by build."
}

resource "google_secret_manager_secret" "agent_app_env" {
  secret_id = "agent-app-env"
  replication {
    auto {}
  }
  depends_on = [google_project_service.apis]
}

resource "google_secret_manager_secret_version" "agent_app_env_placeholder" {
  secret      = google_secret_manager_secret.agent_app_env.id
  secret_data = "# Paste Agent .env.production content. DB_* overridden by build."
}

resource "google_secret_manager_secret_iam_member" "cloudbuild_niobe_app_env" {
  secret_id = google_secret_manager_secret.niobe_app_env.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}@cloudbuild.gserviceaccount.com"
}

resource "google_secret_manager_secret_iam_member" "cloudbuild_agent_app_env" {
  secret_id = google_secret_manager_secret.agent_app_env.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${data.google_project.project.number}@cloudbuild.gserviceaccount.com"
}

# Outputs for scripts and Cloud Build
output "artifact_registry_repo" {
  value       = "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.repo.repository_id}"
  description = "Full Artifact Registry repo path for docker push"
}

output "cloud_sql_connection_name" {
  value       = google_sql_database_instance.main.connection_name
  description = "Cloud SQL connection name (for Cloud Run with Cloud SQL connector)"
}

output "cloud_sql_public_ip" {
  value       = google_sql_database_instance.main.public_ip_address
  description = "Cloud SQL public IP (for DB_HOST in env)"
}

output "project_id" {
  value = var.project_id
}

output "region" {
  value = var.region
}

output "db_username" {
  value     = google_sql_user.db_user.name
  sensitive = false
}

output "db_password" {
  value     = random_password.db.result
  sensitive = true
}
