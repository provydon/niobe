# Link the GitHub repo to the existing connection (created in Console), then create the trigger.
# Connection must exist: Console → Cloud Build → Connect host (e.g. name "niobe"). Then set github_repo_uri.

locals {
  # API requires https://github.com/{owner}/{repo}.git
  github_repo_uri_normalized = var.github_repo_uri != "" && !endswith(var.github_repo_uri, ".git") ? "${var.github_repo_uri}.git" : var.github_repo_uri
}

resource "google_cloudbuildv2_repository" "niobe" {
  count = var.github_repo_uri != "" ? 1 : 0

  name               = "niobe"
  location           = var.region
  parent_connection  = "projects/${var.project_id}/locations/${var.region}/connections/${var.cloud_build_connection_name}"
  remote_uri         = local.github_repo_uri_normalized
}

resource "google_cloudbuild_trigger" "deploy_niobe" {
  count = var.github_repo_uri != "" ? 1 : 0

  name             = "deploy-niobe-with-env"
  location         = var.region
  project          = var.project_id
  service_account  = "projects/${var.project_id}/serviceAccounts/${data.google_project.project.number}@cloudbuild.gserviceaccount.com"

  repository_event_config {
    repository = google_cloudbuildv2_repository.niobe[0].id
    push {
      branch = "^main$"
    }
  }

  filename = "cloudbuild-with-env.yaml"
}

# Trigger: build and deploy only Laravel (run when niobe/** or config changes, or run manually)
resource "google_cloudbuild_trigger" "deploy_laravel_only" {
  count = var.github_repo_uri != "" ? 1 : 0

  name             = "deploy-laravel-only"
  location         = var.region
  project          = var.project_id
  service_account  = "projects/${var.project_id}/serviceAccounts/${data.google_project.project.number}@cloudbuild.gserviceaccount.com"

  repository_event_config {
    repository = google_cloudbuildv2_repository.niobe[0].id
    push {
      branch = "^main$"
    }
  }

  included_files = ["niobe/**", "cloudbuild-laravel.yaml"]
  filename       = "cloudbuild-laravel.yaml"
}

# Trigger: build and deploy only Agent (run when agent/** or config changes, or run manually)
resource "google_cloudbuild_trigger" "deploy_agent_only" {
  count = var.github_repo_uri != "" ? 1 : 0

  name             = "deploy-agent-only"
  location         = var.region
  project          = var.project_id
  service_account  = "projects/${var.project_id}/serviceAccounts/${data.google_project.project.number}@cloudbuild.gserviceaccount.com"

  repository_event_config {
    repository = google_cloudbuildv2_repository.niobe[0].id
    push {
      branch = "^main$"
    }
  }

  included_files = ["agent/**", "cloudbuild-agent.yaml"]
  filename       = "cloudbuild-agent.yaml"
}

output "cloud_build_trigger_id" {
  value       = var.github_repo_uri != "" ? google_cloudbuild_trigger.deploy_niobe[0].id : null
  description = "Cloud Build trigger ID (when github_repo_uri is set)"
}
