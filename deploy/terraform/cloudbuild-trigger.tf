# Cloud Build trigger: run cloudbuild-with-env.yaml when pushing to main (source = GitHub).
# Requires the repo to be linked first (Console → Cloud Build → Repositories → Link repository).
# If you get 400 "invalid argument", the repository path may be wrong. In Console open the repo
# and copy the exact "Resource name" (or use the format below with the repo name shown there).

resource "google_cloudbuild_trigger" "deploy_niobe" {
  count = var.cloud_build_repository != "" ? 1 : 0

  name            = "deploy-niobe-with-env"
  location        = var.region
  project        = var.project_id
  service_account = "projects/${var.project_id}/serviceAccounts/${data.google_project.project.number}@cloudbuild.gserviceaccount.com"

  repository_event_config {
    repository = var.cloud_build_repository
    push {
      branch = "^main$"
    }
  }

  filename = "cloudbuild-with-env.yaml"
}

output "cloud_build_trigger_id" {
  value       = var.cloud_build_repository != "" ? google_cloudbuild_trigger.deploy_niobe[0].id : null
  description = "Cloud Build trigger ID (when repository is set)"
}
