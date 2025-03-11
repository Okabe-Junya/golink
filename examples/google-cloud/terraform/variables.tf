variable "project_id" {
  description = "The Google Cloud project ID"
  type        = string
}

variable "region" {
  description = "The Google Cloud region to deploy resources"
  type        = string
  default     = "asia-northeast1"
}

variable "frontend_service_name" {
  description = "Name for the frontend Cloud Run service"
  type        = string
  default     = "golink-frontend"
}

variable "backend_service_name" {
  description = "Name for the backend Cloud Run service"
  type        = string
  default     = "golink-backend"
}

variable "container_registry" {
  description = "Container registry location (Artifact Registry repository)"
  type        = string
}

variable "frontend_container_image" {
  description = "Frontend container image with tag"
  type        = string
}

variable "backend_container_image" {
  description = "Backend container image with tag"
  type        = string
}

variable "min_instance_count" {
  description = "Minimum number of instances for Cloud Run"
  type        = number
  default     = 0
}

variable "max_instance_count" {
  description = "Maximum number of instances for Cloud Run"
  type        = number
  default     = 10
}

variable "zone" {
  description = "The Google Cloud zone"
  type        = string
  default     = "asia-northeast1-a"
}

variable "service_name" {
  description = "The name of the service"
  type        = string
  default     = "golink"
}

variable "container_image" {
  description = "The container image to deploy"
  type        = string
}

variable "auth_enabled" {
  description = "Enable authentication"
  type        = bool
  default     = true
}

variable "google_client_id" {
  description = "Google OAuth Client ID"
  type        = string
  sensitive   = true
}

variable "google_client_secret" {
  description = "Google OAuth Client Secret"
  type        = string
  sensitive   = true
}

variable "google_allowed_domain" {
  description = "Google Workspace domain allowed to access the application"
  type        = string
  default     = ""
}

variable "session_secret_key" {
  description = "Secret key for session tokens"
  type        = string
  sensitive   = true
  default     = "" # Will be generated if not provided
}
