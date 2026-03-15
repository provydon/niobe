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
