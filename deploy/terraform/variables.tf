variable "project_id" {
  description = "GCP project ID"
  type        = string
}

variable "region" {
  description = "GCP region (e.g. us-central1)"
  type        = string
  default     = "us-central1"
}

variable "db_tier" {
  description = "Cloud SQL instance tier (e.g. db-f1-micro for dev, db-g1-small for prod)"
  type        = string
  default     = "db-f1-micro"
}

variable "db_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "niobe"
}

variable "artifact_repo" {
  description = "Artifact Registry repository name for container images"
  type        = string
  default     = "niobe"
}

# GitHub repo URI for linking to the existing connection (e.g. https://github.com/provydon/niobe).
# Terraform will create the repository link and trigger. Leave empty to skip trigger.
variable "github_repo_uri" {
  description = "GitHub repo URI (e.g. https://github.com/owner/niobe). Used to create repository link and trigger. Leave empty to skip."
  type        = string
  default     = ""
}

variable "cloud_build_connection_name" {
  description = "Name of the 2nd gen Cloud Build connection (created in Console). Default niobe."
  type        = string
  default     = "niobe"
}

# Custom domains for Cloud Run (optional). Leave empty to skip domain mapping.
# You must verify domain ownership in GCP (Console → Cloud Run → Domain mappings or Webmaster Central) before apply.
variable "custom_domain_web" {
  description = "Custom domain for the Laravel web service (e.g. niobe.live). Leave empty to skip."
  type        = string
  default     = ""
}

variable "custom_domain_agent" {
  description = "Custom domain for the Agent service (e.g. agent.niobe.live). Leave empty to skip."
  type        = string
  default     = ""
}
