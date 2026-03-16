# Map custom domains to Cloud Run services (niobe-web and niobe-agent).
# Services must exist (deploy at least once via Cloud Build). Domain ownership must be verified in GCP first.
# See: https://cloud.google.com/run/docs/mapping-custom-domains

resource "google_cloud_run_domain_mapping" "web" {
  count    = var.custom_domain_web != "" ? 1 : 0
  location = var.region
  name     = var.custom_domain_web

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = "niobe-web"
  }
}

resource "google_cloud_run_domain_mapping" "agent" {
  count    = var.custom_domain_agent != "" ? 1 : 0
  location = var.region
  name     = var.custom_domain_agent

  metadata {
    namespace = var.project_id
  }

  spec {
    route_name = "niobe-agent"
  }
}

output "custom_domain_web" {
  value       = var.custom_domain_web != "" ? var.custom_domain_web : null
  description = "Custom domain for niobe-web (if set)"
}

output "custom_domain_agent" {
  value       = var.custom_domain_agent != "" ? var.custom_domain_agent : null
  description = "Custom domain for niobe-agent (if set)"
}
