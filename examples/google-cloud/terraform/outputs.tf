output "frontend_url" {
  description = "The URL of the frontend Cloud Run service"
  value       = google_cloud_run_v2_service.frontend.uri
}

output "backend_url" {
  description = "The URL of the backend Cloud Run service"
  value       = google_cloud_run_v2_service.backend.uri
}

output "frontend_service_name" {
  description = "The name of the frontend Cloud Run service"
  value       = google_cloud_run_v2_service.frontend.name
}

output "backend_service_name" {
  description = "The name of the backend Cloud Run service"
  value       = google_cloud_run_v2_service.backend.name
}

output "service_account" {
  description = "The email of the service account used by Cloud Run services"
  value       = google_service_account.cloud_run_service_account.email
}

output "firestore_database" {
  description = "Firestore database name"
  value       = google_firestore_database.database.name
}
