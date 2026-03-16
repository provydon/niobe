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

# Full resource name of the 2nd gen Cloud Build repo (after linking in Console).
# Example: projects/niobe-489920/locations/us-central1/connections/CONN_ID/repositories/niobe
# Get from: gcloud builds repositories list --region=us-central1 --project=PROJECT_ID
variable "cloud_build_repository" {
  description = "2nd gen Cloud Build repository resource name (for trigger). Leave empty to skip creating the trigger."
  type        = string
  default     = ""
}
